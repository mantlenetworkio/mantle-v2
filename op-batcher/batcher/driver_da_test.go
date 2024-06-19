package batcher

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"testing"

	"github.com/Layr-Labs/datalayr/common/graphView"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/ethereum-optimism/optimism/l2geth/common/hexutil"
	"github.com/ethereum-optimism/optimism/l2geth/rlp"
	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestIsChannelFull(t *testing.T) {
	log := testlog.Logger(t, log.LvlCrit)
	m := NewChannelManager(log, metrics.NoopMetrics, ChannelConfig{
		ChannelTimeout: 100,
	}, nil)
	require.NoError(t, m.ensurePendingChannel(eth.BlockID{}))
	channelID := m.pendingChannel.ID()
	frame := frameData{
		data: []byte{},
		id: frameID{
			chID:        channelID,
			frameNumber: uint16(0),
		},
	}
	m.pendingChannel.PushFrame(frame)

	isChannelFull := m.pendingChannel != nil && m.pendingChannel.IsFull() && !m.pendingChannel.HasFrame()

	require.False(t, isChannelFull)
	m.nextTxData()

	isChannelFull = m.pendingChannel != nil && m.pendingChannel.IsFull() && !m.pendingChannel.HasFrame()
	require.False(t, isChannelFull)
	m.pendingChannel.setFullErr(m.pendingChannel.timeoutReason)

	isChannelFull = m.pendingChannel != nil && m.pendingChannel.IsFull() && !m.pendingChannel.HasFrame()

	require.True(t, isChannelFull)

}

func TestTxAggregator(t *testing.T) {
	log := testlog.Logger(t, log.LvlCrit)
	zeroLog := zerolog.Nop()
	graphLog := &logging.Logger{
		Logger: &zeroLog,
	}
	m := NewChannelManager(log, metrics.NoopMetrics, ChannelConfig{
		ChannelTimeout: 100,
	}, nil)
	graphClient := graphView.NewGraphClient("", graphLog)
	b := &BatchSubmitter{
		Config: Config{
			log:           log,
			RollupMaxSize: 100,
			GraphClient:   graphClient,
			metr:          metrics.NewMetrics("default"),
		},
		txMgr: nil,
		state: m,
	}

	require.NoError(t, b.state.ensurePendingChannel(eth.BlockID{}))

	bytes := make([]byte, 60)
	frame0 := frameData{
		data: bytes,
		id: frameID{
			chID:        b.state.pendingChannel.ID(),
			frameNumber: uint16(0),
		},
	}
	txData0 := txData{frame: frame0}
	frame1 := frameData{
		data: bytes,
		id: frameID{
			chID:        b.state.pendingChannel.ID(),
			frameNumber: uint16(1),
		},
	}
	txData1 := txData{frame: frame1}

	b.state.daPendingTxData[txData0.ID()] = txData0
	b.state.daPendingTxData[txData1.ID()] = txData1
	require.Equal(t, len(b.state.daPendingTxData), 2)
	by, err := b.txAggregator()
	require.NoError(t, err)
	require.True(t, len(by) < 100)
	require.Equal(t, len(b.state.daUnConfirmedTxID), 1)
	require.Equal(t, txData0.ID(), b.state.daUnConfirmedTxID[0])

}

func TestConfirmDataStore(t *testing.T) {
	_, opts, _, _, err := setupDataLayrServiceManage()
	require.NoError(t, err)
	abi, err := bindings.ContractDataLayrServiceManagerMetaData.GetAbi()
	require.NoError(t, err)
	searchData := &bindings.IDataLayrServiceManagerDataStoreSearchData{
		Duration:  1,
		Timestamp: new(big.Int).SetUint64(uint64(1530000000)),
		Index:     0,
		Metadata: bindings.IDataLayrServiceManagerDataStoreMetadata{
			HeaderHash:           [32]byte{},
			DurationDataStoreId:  1,
			GlobalDataStoreId:    1,
			ReferenceBlockNumber: 1,
			BlockNumber:          uint32(1),
			Fee:                  big.NewInt(100),
			Confirmer:            opts.From,
			SignatoryRecordHash:  [32]byte{},
		},
	}

	b := &BatchSubmitter{
		Config: Config{},
		txMgr:  nil,
	}
	var bytes = []byte("test")
	txD, err := b.confirmDataTxData(abi, bytes, searchData)
	require.NoError(t, err)
	require.True(t, len(txD) > 0)

}

func TestDataStoreTxData(t *testing.T) {
	_, opts, _, _, err := setupDataLayrServiceManage()
	require.NoError(t, err)
	abi, err := bindings.ContractDataLayrServiceManagerMetaData.GetAbi()
	require.NoError(t, err)

	var bytes = []byte("test")

	txD, err := abi.Pack(
		"initDataStore",
		opts.From,
		opts.From,
		uint8(1),
		uint32(1),
		uint32(1),
		bytes)
	require.NoError(t, err)
	require.True(t, len(txD) > 0)

}

func setupDataLayrServiceManage() (common.Address, *bind.TransactOpts, *backends.SimulatedBackend, *bindings.ContractDataLayrServiceManager, error) {
	privateKey, err := crypto.GenerateKey()
	from := crypto.PubkeyToAddress(privateKey.PublicKey)
	if err != nil {
		return common.Address{}, nil, nil, nil, err
	}
	opts, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(1337))
	if err != nil {
		return common.Address{}, nil, nil, nil, err
	}
	backend := backends.NewSimulatedBackend(core.GenesisAlloc{from: {Balance: big.NewInt(params.Ether)}}, 50_000_000)
	_, _, contract, err := bindings.DeployContractDataLayrServiceManager(
		opts,
		backend,
		from,
		from,
		from,
		from,
		from,
		from,
		from,
	)
	if err != nil {
		return common.Address{}, nil, nil, nil, err
	}
	return from, opts, backend, contract, nil
}

func TestBatchSubmitter_Send(t *testing.T) {
	tests := []struct {
		name    string
		bs      *BatchSubmitter
		want    bool
		wantErr bool
	}{
		{
			name: "t1",
		},
	}
	cfgData := `{
		"L1EthRpc": "http://localhost:9875",
		"L2EthRpc": "http://127.0.0.1:9874",
		"RollupRpc": "http://127.0.0.1:9876",
		"DisperserSocket": "127.0.0.1:31011",
		"DisperserTimeout": 60000000000,
		"DataStoreDuration": 1,
		"GraphPollingDuration": 60000000000,
		"GraphProvider": "http://127.0.0.1:8000/subgraphs/name/datalayr",
		"RollupMaxSize": 128000,
		"MantleDaNodes": 3,
		"MaxChannelDuration": 10,
		"SubSafetyMargin": 4,
		"PollInterval": 1000000000,
		"MaxPendingTransactions": 1,
		"MaxL1TxSize": 10000,
		"Stopped": false,
		"TxMgrConfig": {
			"L1RPCURL": "http://localhost:9875",
			"Mnemonic": "",
			"HDPath": "",
			"SequencerHDPath": "m/44'/60'/0'/0/3",
			"L2OutputHDPath": "",
			"PrivateKey": "89310b99c43b06741c142fe0e8f122dff6b15e4223aeb183b1e87c9a700f95b9",
			"SignerCLIConfig": {
				"Endpoint": "",
				"Address": "",
				"TLSConfig": {
					"TLSCaCert": "tls/ca.crt",
					"TLSCert": "tls/tls.crt",
					"TLSKey": "tls/tls.key"
				}
			},
			"NumConfirmations": 1,
			"SafeAbortNonceTooLowCount": 3,
			"FeeLimitMultiplier": 5,
			"FeeLimitThresholdGwei": 100,
			"ResubmissionTimeout": 48000000000,
			"ReceiptQueryInterval": 12000000000,
			"NetworkTimeout": 2000000000,
			"TxSendTimeout": 120000000000,
			"TxNotInMempoolTimeout": 120000000000,
			"EnableHsm": false,
			"HsmCreden": "7b0a20202274797065223a2022736572766963655f6163636f756e74222c0a20202270726f6a6563745f6964223a20226d616e746c652d333831333032222c0a202022707269766174655f6b65795f6964223a202239633661386163613562353730663236643932613332633431626663363162633865363632373833222c0a202022707269766174655f6b6579223a20222d2d2d2d2d424547494e2050524956415445204b45592d2d2d2d2d5c6e4d494945766749424144414e42676b71686b6947397730424151454641415343424b67776767536b41674541416f49424151445a6644614c6275327835774f445c6e484a546e2f7176462f774e786e477875307045537577785854716d4f3748626f73663047484b43394c39566b6c2b626a544455354e546359676569636d63384e5c6e496276343854523945644a544133794c456d4a7a6d66674973512f714f54557263454d516c2b552f646b4a41796d416c782b435734424a61334b397a31444a425c6e3456374b39574c61507a376a6e6f625669686656795751506a32345035395452366b31506b53776752445147704f79513632457439376a5347555478494f316f5c6e394e76517a75562f4535437a4346682f372b6251464a3232487939346274344f424b6d6c776947755769614a4575674b4d6c6931454e74736c5869447466726f5c6e6d637861455373395044705068485866346e5644477265484f4d4c4b4346686641423030775a685461575a4a6f474a39566c514e4574315136346b32547a4d375c6e5a4a754d7861636e41674d4241414543676745414c62574c69546168516e69354a6a39466c4a545436574d3169425647504f79496a52552b2f4d4b4e704870535c6e61346d74456a484748723045376f563267324d713949455975552f6b59625635374e71674e53774d7968534b7a654f33733073443469514547312b4c5a7344725c6e5364766f584835774d697861744639555964786d655939536a454a42706568395035647359742b384450367036784551616155436157355156327a66787675685c6e5030543071686b38563273534c64784a454144325a32614431494748746b417943384c696f31386c457463555a636233347470387a556f6f794d6a4d475667725c6e493037714a52494a635a4e53315a4d48416979476e2f75463745324d6f52517457325373476a484f7171636f78717074776657453831627964344b635069476c5c6e31646c5573676e61344d66653831344f3778573977467a585736307a4f72364d2b4b6275327a7a6a30514b4267514437555952646a3653594256716a662f504a5c6e4b613372635a7139494262434e526b4c6234692b7a506b784e6251502f793733765a4a3443587a455368543637334242427a6b35744e386164533152554c436d5c6e4f7163516a4b79327949657366352b45744d62632b4331734c6c48705873737a5733794f5056536a6850557444587050504f42577a3537775a3146556e3644775c6e357074534e5171414377424b6c424c643534504f654c706849774b42675144646956704e6c6f336266396747372b7077683173543939494d374965556e6a69535c6e5a425a5161786e6e312f3948716d48584f6d4a2b6b2b6b61303468313138474947395576745876426d37315a466f5875545234307453676d62615548787170625c6e354a787039544f76715353636c6d6c6f576230572f72614c534a774f774d7465686f6f76307a6f686e422f4f50386533306c644d4f4e7032647a7776383767375c6e4c48706e485235634c514b42675144446a61755562626769506c42483173456f4c314651576561513851346b63644b61446d42324c754a63417a436f486555375c6e436e7956414c546675394656624d694a49516a4c4f553038746837634868424757473830746e4753444c6c645a5455487575376564424a4d456b4c5564316c675c6e44666a2b6151535a39465165695655356f4f486a53737965765a59515a65474363623438476c2b67514738716d4d7552645a73664a74764878774b42674857735c6e4c70354632535839653062384375416f315a542b72734555706c4f6e303037584151394952474e6b31514642484756526174336e505174317a756368616e67635c6e714a6d463461324f527635614f31752f394d7030613159324b56472f45654272787a56302f445a544e744a434273315a31566d7768452f7069704d2f6a7761765c6e6d686b624c71614a6f6b395169346f316e513873703859445161494b364248755a7a6e384f704d6c416f4742414c6a355932496a552f534e416c694e456748665c6e58533551726434583672635139777a52373861457251424d636e364e5659432f424a534a4d6f344832666b5a385736617454544b75616d4a4e657276754a70645c6e77686d4a475a4e4e67634372334559634d366565476c455045616d6f36374d464b6d4b6945484257714d7037522f6f395061597a624b667774554c64546d71345c6e4749336250676d4954346a72442f3932452f414c427544475c6e2d2d2d2d2d454e442050524956415445204b45592d2d2d2d2d5c6e222c0a202022636c69656e745f656d61696c223a2022636c6f756468736d2d7161406d616e746c652d3338313330322e69616d2e67736572766963656163636f756e742e636f6d222c0a202022636c69656e745f6964223a2022313035393734343937353938323137383437323831222c0a202022617574685f757269223a202268747470733a2f2f6163636f756e74732e676f6f676c652e636f6d2f6f2f6f61757468322f61757468222c0a202022746f6b656e5f757269223a202268747470733a2f2f6f61757468322e676f6f676c65617069732e636f6d2f746f6b656e222c0a202022617574685f70726f76696465725f783530395f636572745f75726c223a202268747470733a2f2f7777772e676f6f676c65617069732e636f6d2f6f61757468322f76312f6365727473222c0a202022636c69656e745f783530395f636572745f75726c223a202268747470733a2f2f7777772e676f6f676c65617069732e636f6d2f726f626f742f76312f6d657461646174612f783530392f636c6f756468736d2d71612534306d616e746c652d3338313330322e69616d2e67736572766963656163636f756e742e636f6d222c0a202022756e6976657273655f646f6d61696e223a2022676f6f676c65617069732e636f6d220a7d0a",
			"HsmAddress": "0x14E4FF2909EEB2bc7379bcfF8263f04671B0afDA",
			"HsmAPIName": "projects/mantle-381302/locations/global/keyRings/qa/cryptoKeys/sequencer-qa/cryptoKeyVersions/1"
		},
		"RPCConfig": {
			"ListenAddr": "0.0.0.0",
			"ListenPort": 6545,
			"EnableAdmin": true
		},
		"LogConfig": {
			"Level": "info",
			"Color": false,
			"Format": "text"
		},
		"MetricsConfig": {
			"Enabled": true,
			"ListenAddr": "0.0.0.0",
			"ListenPort": 7302
		},
		"PprofConfig": {
			"Enabled": true,
			"ListenAddr": "0.0.0.0",
			"ListenPort": 6065
		},
		"CompressorConfig": {
			"TargetL1TxSizeBytes": 100000,
			"TargetNumFrames": 1,
			"ApproxComprRatio": 0.4,
			"Kind": "ratio"
		},
		"EigenLogConfig": {
			"Path": "",
			"Prefix": "",
			"FileLevel": "",
			"StdLevel": ""
		},
		"EigenDAConfig": {
			"RPC": "disperser-holesky.eigenda.xyz:444",
			"StatusQueryRetryInterval": 5000000000,
			"StatusQueryTimeout": 1200000000000
		}
	}`
	cfg := CLIConfig{}
	_ = json.Unmarshal([]byte(cfgData), &cfg)

	batchSubmitter, err := NewBatchSubmitterFromCLIConfig(cfg, log.New(), metrics.NewMetrics("test"))
	fmt.Println(batchSubmitter, err)
	tests[0].bs = batchSubmitter
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.bs
			var candidate *txmgr.TxCandidate
			var err error
			da := [358834]byte{}
			if candidate, err = l.blobTxCandidate(da[:]); err != nil {
				l.log.Error("failed to create blob tx candidate", "err", err)
				return
			}
			got, err := l.txMgr.Send(context.Background(), *candidate)
			if err != nil {
				t.Errorf("BatchSubmitter.loopEigenDa() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(got, err)
		})
	}
}

const ChannelIDLength = 16

type ChannelID [ChannelIDLength]byte

type Frame struct {
	ID          ChannelID `json:"id"`
	FrameNumber uint16    `json:"frame_number"`
	Data        []byte    `json:"data"`
	IsLast      bool      `json:"is_last"`
}

func (f *Frame) MarshalBinary(w io.Writer) error {
	_, err := w.Write(f.ID[:])
	if err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, f.FrameNumber); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, uint32(len(f.Data))); err != nil {
		return err
	}
	_, err = w.Write(f.Data)
	if err != nil {
		return err
	}
	if f.IsLast {
		if _, err = w.Write([]byte{1}); err != nil {
			return err
		}
	} else {
		if _, err = w.Write([]byte{0}); err != nil {
			return err
		}
	}
	return nil
}

type ByteReader interface {
	io.Reader
	io.ByteReader
}

const MaxFrameLen = 1_000_000

func (f *Frame) UnmarshalBinary(r ByteReader) error {
	if _, err := io.ReadFull(r, f.ID[:]); err != nil {
		// Forward io.EOF here ok, would mean not a single byte from r.
		return fmt.Errorf("reading channel_id: %w", err)
	}
	if err := binary.Read(r, binary.BigEndian, &f.FrameNumber); err != nil {
		return fmt.Errorf("reading frame_number: %w", (err))
	}

	var frameLength uint32
	if err := binary.Read(r, binary.BigEndian, &frameLength); err != nil {
		return fmt.Errorf("reading frame_data_length: %w", (err))
	}

	// Cap frame length to MaxFrameLen (currently 1MB)
	if frameLength > MaxFrameLen {
		return fmt.Errorf("frame_data_length is too large: %d", frameLength)
	}
	f.Data = make([]byte, int(frameLength))
	if _, err := io.ReadFull(r, f.Data); err != nil {
		return fmt.Errorf("reading frame_data: %w", (err))
	}

	if isLastByte, err := r.ReadByte(); err != nil {
		return fmt.Errorf("reading final byte (is_last): %w", (err))
	} else if isLastByte == 0 {
		f.IsLast = false
	} else if isLastByte == 1 {
		f.IsLast = true
	} else {
		return errors.New("invalid byte as is_last")
	}
	return nil
}

func ParseFrames(data []byte) ([]Frame, error) {
	if len(data) == 0 {
		return nil, errors.New("data array must not be empty")
	}
	if data[0] != 0 {
		return nil, fmt.Errorf("invalid derivation format byte: got %d", data[0])
	}
	buf := bytes.NewBuffer(data[1:])
	var frames []Frame
	for buf.Len() > 0 {
		var f Frame
		if err := f.UnmarshalBinary(buf); err != nil {
			return nil, fmt.Errorf("parsing frame %d: %w", len(frames), err)
		}
		frames = append(frames, f)
	}
	if buf.Len() != 0 {
		return nil, fmt.Errorf("did not fully consume data: have %d frames and %d bytes left", len(frames), buf.Len())
	}
	if len(frames) == 0 {
		return nil, errors.New("was not able to find any frames")
	}
	return frames, nil
}

func TestUnmashal(t *testing.T) {
	rawData, _ := hexutil.Decode("0x009a46d4128b0a0ca5a14d01f98534a12600000000027578dadae1c3f0c373816836abd825fffbafce3b78f408dddd3bf9cf69ab4a978b256ce6666702abe6af7b19b6806bf1f9092617df2ffba6b896e9ff5c8959ff58a7c91eb4f1bfb0fd5982be9471e1c596b404d1e0033b6e31fcb8be20ef78f0d3a2e54e87394edf7ef8f7d92bb98bdf9a4e1f7b75ef58d0ac851ecf4e047b872f58fa74d7e26d8e79a5a9dddcab56ab3f7559915b15e3785a82e59f7677b7ea3375719069a13f7a7774ffe8e468b19e758aa185f3fa119d294e0c1880bf61c9e3be893f19f0027e17b705750ddf949604ac799f25713e85d99ca526312f6f9dc3491f6f0719afe59f7fc5142fd07929b7ebd69d851ee552fe8cd27f2f5cda58c71ad89c93d25f9d507b805b6ff13f48485d36fbf7b738a97475fd8ae0a592765f03ad96143d909f79f6d83c8b1bed1f34176f22ce6fe10720a6ddfb70d6dcf581f9e32b02a5e7afc932cfe4e42b986495feba64a6e2e7578ce7181889332d126ada37939f6dcdd699df7d5edd7df7cafaa2bff915a1b3b94a3db785d39ab787a4775c8a58b06526f77f7387049f2f76130a153f3c7bfed37507e3bd402539f3a7efe4c576b357824c8b869a76274aefba789791815de6fdd49273996c8b0558f7acb6fbeda665f6957bffee599388332d166a5a52c91b7be5ed7a7f0c22da2557353df7cffbbd4e4e7356f0a505ba0cef0215181489332d1e6a1af3e524894a8e1f99eed3f7cd7a6f27afd6af55182617d4af9f75ae34f62efbbc33910b84d6f0c72eb879e6ef9e85513c2181b26c2b765e0c7c1cb041a87642fddee017767c20d312a1a6c5d85aacbe59667863e76f41d392a32ff654a4f3ccfc9558b35ceaa3db82dd9bbf2c20ceb4e40380000000ffff5759847e01")
	//rawData, _ := hexutil.Decode("0xed12f0010a2098efdcc50cd8f59297d7688adb274d02a4eae30c8f199be6c9dd44571c44a06210661886a2662202000128a066a206bd01633836363438666337623636656239323934343938383865343634623134636434353133393331336439323465386631313864356330666566633431316662392d33313337333133373335333733383337333833303330333833373336333133323337333833323266333032663333333332663331326633333333326665336230633434323938666331633134396166626634633839393666623932343237616534316534363439623933346361343935393931623738353262383535")
	calldataFrame := &op_service.CalldataFrame{}
	fmt.Println(rawData[0])
	err := proto.Unmarshal(rawData[1:], calldataFrame)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(calldataFrame)
	}
}

func TestCallEigenDA(t *testing.T) {
	cfgData := `{
		"L1EthRpc": "http://127.0.0.1:8545",
		"L2EthRpc": "http://127.0.0.1:9545",
		"RollupRpc": "http://127.0.0.1:7545",
		"DisperserSocket": "127.0.0.1:31011",
		"DisperserTimeout": 60000000000,
		"DataStoreDuration": 1,
		"GraphPollingDuration": 60000000000,
		"GraphProvider": "http://127.0.0.1:8000/subgraphs/name/datalayr",
		"RollupMaxSize": 128000,
		"MantleDaNodes": 3,
		"MaxChannelDuration": 10,
		"SubSafetyMargin": 4,
		"PollInterval": 1000000000,
		"MaxPendingTransactions": 1,
		"MaxL1TxSize": 10000,
		"Stopped": false,
		"TxMgrConfig": {
			"L1RPCURL": "https://eth-sepolia.g.alchemy.com/v2/XMS1J6f654XZolfd7oaMe-kaNPEpWifX",
			"Mnemonic": "",
			"HDPath": "",
			"SequencerHDPath": "m/44'/60'/0'/0/3",
			"L2OutputHDPath": "",
			"PrivateKey": "89310b99c43b06741c142fe0e8f122dff6b15e4223aeb183b1e87c9a700f95b9",
			"SignerCLIConfig": {
				"Endpoint": "",
				"Address": "",
				"TLSConfig": {
					"TLSCaCert": "tls/ca.crt",
					"TLSCert": "tls/tls.crt",
					"TLSKey": "tls/tls.key"
				}
			},
			"NumConfirmations": 1,
			"SafeAbortNonceTooLowCount": 3,
			"FeeLimitMultiplier": 5,
			"FeeLimitThresholdGwei": 100,
			"ResubmissionTimeout": 48000000000,
			"ReceiptQueryInterval": 12000000000,
			"NetworkTimeout": 2000000000,
			"TxSendTimeout": 120000000000,
			"TxNotInMempoolTimeout": 120000000000,
			"EnableHsm": false,
			"HsmCreden": "7b0a20202274797065223a2022736572766963655f6163636f756e74222c0a20202270726f6a6563745f6964223a20226d616e746c652d333831333032222c0a202022707269766174655f6b65795f6964223a202239633661386163613562353730663236643932613332633431626663363162633865363632373833222c0a202022707269766174655f6b6579223a20222d2d2d2d2d424547494e2050524956415445204b45592d2d2d2d2d5c6e4d494945766749424144414e42676b71686b6947397730424151454641415343424b67776767536b41674541416f49424151445a6644614c6275327835774f445c6e484a546e2f7176462f774e786e477875307045537577785854716d4f3748626f73663047484b43394c39566b6c2b626a544455354e546359676569636d63384e5c6e496276343854523945644a544133794c456d4a7a6d66674973512f714f54557263454d516c2b552f646b4a41796d416c782b435734424a61334b397a31444a425c6e3456374b39574c61507a376a6e6f625669686656795751506a32345035395452366b31506b53776752445147704f79513632457439376a5347555478494f316f5c6e394e76517a75562f4535437a4346682f372b6251464a3232487939346274344f424b6d6c776947755769614a4575674b4d6c6931454e74736c5869447466726f5c6e6d637861455373395044705068485866346e5644477265484f4d4c4b4346686641423030775a685461575a4a6f474a39566c514e4574315136346b32547a4d375c6e5a4a754d7861636e41674d4241414543676745414c62574c69546168516e69354a6a39466c4a545436574d3169425647504f79496a52552b2f4d4b4e704870535c6e61346d74456a484748723045376f563267324d713949455975552f6b59625635374e71674e53774d7968534b7a654f33733073443469514547312b4c5a7344725c6e5364766f584835774d697861744639555964786d655939536a454a42706568395035647359742b384450367036784551616155436157355156327a66787675685c6e5030543071686b38563273534c64784a454144325a32614431494748746b417943384c696f31386c457463555a636233347470387a556f6f794d6a4d475667725c6e493037714a52494a635a4e53315a4d48416979476e2f75463745324d6f52517457325373476a484f7171636f78717074776657453831627964344b635069476c5c6e31646c5573676e61344d66653831344f3778573977467a585736307a4f72364d2b4b6275327a7a6a30514b4267514437555952646a3653594256716a662f504a5c6e4b613372635a7139494262434e526b4c6234692b7a506b784e6251502f793733765a4a3443587a455368543637334242427a6b35744e386164533152554c436d5c6e4f7163516a4b79327949657366352b45744d62632b4331734c6c48705873737a5733794f5056536a6850557444587050504f42577a3537775a3146556e3644775c6e357074534e5171414377424b6c424c643534504f654c706849774b42675144646956704e6c6f336266396747372b7077683173543939494d374965556e6a69535c6e5a425a5161786e6e312f3948716d48584f6d4a2b6b2b6b61303468313138474947395576745876426d37315a466f5875545234307453676d62615548787170625c6e354a787039544f76715353636c6d6c6f576230572f72614c534a774f774d7465686f6f76307a6f686e422f4f50386533306c644d4f4e7032647a7776383767375c6e4c48706e485235634c514b42675144446a61755562626769506c42483173456f4c314651576561513851346b63644b61446d42324c754a63417a436f486555375c6e436e7956414c546675394656624d694a49516a4c4f553038746837634868424757473830746e4753444c6c645a5455487575376564424a4d456b4c5564316c675c6e44666a2b6151535a39465165695655356f4f486a53737965765a59515a65474363623438476c2b67514738716d4d7552645a73664a74764878774b42674857735c6e4c70354632535839653062384375416f315a542b72734555706c4f6e303037584151394952474e6b31514642484756526174336e505174317a756368616e67635c6e714a6d463461324f527635614f31752f394d7030613159324b56472f45654272787a56302f445a544e744a434273315a31566d7768452f7069704d2f6a7761765c6e6d686b624c71614a6f6b395169346f316e513873703859445161494b364248755a7a6e384f704d6c416f4742414c6a355932496a552f534e416c694e456748665c6e58533551726434583672635139777a52373861457251424d636e364e5659432f424a534a4d6f344832666b5a385736617454544b75616d4a4e657276754a70645c6e77686d4a475a4e4e67634372334559634d366565476c455045616d6f36374d464b6d4b6945484257714d7037522f6f395061597a624b667774554c64546d71345c6e4749336250676d4954346a72442f3932452f414c427544475c6e2d2d2d2d2d454e442050524956415445204b45592d2d2d2d2d5c6e222c0a202022636c69656e745f656d61696c223a2022636c6f756468736d2d7161406d616e746c652d3338313330322e69616d2e67736572766963656163636f756e742e636f6d222c0a202022636c69656e745f6964223a2022313035393734343937353938323137383437323831222c0a202022617574685f757269223a202268747470733a2f2f6163636f756e74732e676f6f676c652e636f6d2f6f2f6f61757468322f61757468222c0a202022746f6b656e5f757269223a202268747470733a2f2f6f61757468322e676f6f676c65617069732e636f6d2f746f6b656e222c0a202022617574685f70726f76696465725f783530395f636572745f75726c223a202268747470733a2f2f7777772e676f6f676c65617069732e636f6d2f6f61757468322f76312f6365727473222c0a202022636c69656e745f783530395f636572745f75726c223a202268747470733a2f2f7777772e676f6f676c65617069732e636f6d2f726f626f742f76312f6d657461646174612f783530392f636c6f756468736d2d71612534306d616e746c652d3338313330322e69616d2e67736572766963656163636f756e742e636f6d222c0a202022756e6976657273655f646f6d61696e223a2022676f6f676c65617069732e636f6d220a7d0a",
			"HsmAddress": "0x14E4FF2909EEB2bc7379bcfF8263f04671B0afDA",
			"HsmAPIName": "projects/mantle-381302/locations/global/keyRings/qa/cryptoKeys/sequencer-qa/cryptoKeyVersions/1"
		},
		"RPCConfig": {
			"ListenAddr": "0.0.0.0",
			"ListenPort": 6545,
			"EnableAdmin": true
		},
		"LogConfig": {
			"Level": "info",
			"Color": false,
			"Format": "text"
		},
		"MetricsConfig": {
			"Enabled": true,
			"ListenAddr": "0.0.0.0",
			"ListenPort": 7302
		},
		"PprofConfig": {
			"Enabled": true,
			"ListenAddr": "0.0.0.0",
			"ListenPort": 6065
		},
		"CompressorConfig": {
			"TargetL1TxSizeBytes": 100000,
			"TargetNumFrames": 1,
			"ApproxComprRatio": 0.4,
			"Kind": "ratio"
		},
		"EigenLogConfig": {
			"Path": "",
			"Prefix": "",
			"FileLevel": "",
			"StdLevel": ""
		},
		"EigenDAConfig": {
			"RPC": "disperser-holesky.eigenda.xyz:443",
			"StatusQueryRetryInterval": 5000000000,
			"StatusQueryTimeout": 1200000000000
		}
	}`
	cfg := CLIConfig{}
	_ = json.Unmarshal([]byte(cfgData), &cfg)

	batchSubmitter, err := NewBatchSubmitterFromCLIConfig(cfg, log.New(), metrics.NewMetrics("test"))

	ctx := context.Background()
	tx, pd, err := batchSubmitter.L1Client.TransactionByHash(ctx, common.HexToHash("0xfae251d4d80abdf7e6ee1bb2f230169e335b7dd02112f66398cd92be1f1a85bc"))
	if pd || err != nil {
		t.Error(err)
		return
	}

	rawData := tx.Data()
	//rawData, _ := hexutil.Decode("0xed12f0010a2098efdcc50cd8f59297d7688adb274d02a4eae30c8f199be6c9dd44571c44a06210661886a2662202000128a066a206bd01633836363438666337623636656239323934343938383865343634623134636434353133393331336439323465386631313864356330666566633431316662392d33313337333133373335333733383337333833303330333833373336333133323337333833323266333032663333333332663331326633333333326665336230633434323938666331633134396166626634633839393666623932343237616534316534363439623933346361343935393931623738353262383535")
	calldataFrame := &op_service.CalldataFrame{}
	fmt.Println(rawData[0])
	err = proto.Unmarshal(rawData[1:], calldataFrame)
	if err != nil {
		t.Error(err)
	} else {
		t.Log(calldataFrame)
	}

	switch calldataFrame.Value.(type) {
	case *op_service.CalldataFrame_FrameRef:
		frameRef := calldataFrame.GetFrameRef()
		if len(frameRef.QuorumIds) == 0 {
			t.Error("decoded frame ref contains no quorum IDs", "err", err)
			return
		}

		log.Info("requesting data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber,
			"batchHeaderHash", base64.StdEncoding.EncodeToString(frameRef.BatchHeaderHash), "blobIndex", frameRef.BlobIndex, "blobLength", frameRef.BlobLength)
		data, err := batchSubmitter.eigenDA.RetrieveBlob(context.Background(), frameRef.BatchHeaderHash, frameRef.BlobIndex)

		out := []eth.Data{}

		data = data[:frameRef.BlobLength]
		err = rlp.DecodeBytes(data, &out)
		if err != nil {
			log.Error("Decode retrieval frames in error,skip wrong data", "err", err, "blobInfo", fmt.Sprintf("%x:%d", frameRef.BatchHeaderHash, frameRef.BlobIndex))
			return
		}

		for _, d := range out {
			fs, err := ParseFrames(d)
			if err != nil {
				t.Error(err)
				return
			}
			for _, f := range fs {
				fmt.Println("frame:", hexutil.Encode(f.ID[:]), f.FrameNumber, len(f.Data))
			}
		}
	}

}

func TestCallEigenDA2(t *testing.T) {
	cfgData := `{
		"L1EthRpc": "http://localhost:8545",
		"L2EthRpc": "http://127.0.0.1:9545",
		"RollupRpc": "http://127.0.0.1:7545",
		"DisperserSocket": "127.0.0.1:31011",
		"DisperserTimeout": 60000000000,
		"DataStoreDuration": 1,
		"GraphPollingDuration": 60000000000,
		"GraphProvider": "http://127.0.0.1:8000/subgraphs/name/datalayr",
		"RollupMaxSize": 128000,
		"MantleDaNodes": 3,
		"MaxChannelDuration": 10,
		"SubSafetyMargin": 4,
		"PollInterval": 1000000000,
		"MaxPendingTransactions": 1,
		"MaxL1TxSize": 10000,
		"Stopped": false,
		"TxMgrConfig": {
			"L1RPCURL": "http://localhost:8545",
			"Mnemonic": "",
			"HDPath": "",
			"SequencerHDPath": "m/44'/60'/0'/0/3",
			"L2OutputHDPath": "",
			"PrivateKey": "89310b99c43b06741c142fe0e8f122dff6b15e4223aeb183b1e87c9a700f95b9",
			"SignerCLIConfig": {
				"Endpoint": "",
				"Address": "",
				"TLSConfig": {
					"TLSCaCert": "tls/ca.crt",
					"TLSCert": "tls/tls.crt",
					"TLSKey": "tls/tls.key"
				}
			},
			"NumConfirmations": 1,
			"SafeAbortNonceTooLowCount": 3,
			"FeeLimitMultiplier": 5,
			"FeeLimitThresholdGwei": 100,
			"ResubmissionTimeout": 48000000000,
			"ReceiptQueryInterval": 12000000000,
			"NetworkTimeout": 2000000000,
			"TxSendTimeout": 120000000000,
			"TxNotInMempoolTimeout": 120000000000,
			"EnableHsm": false,
			"HsmCreden": "7b0a20202274797065223a2022736572766963655f6163636f756e74222c0a20202270726f6a6563745f6964223a20226d616e746c652d333831333032222c0a202022707269766174655f6b65795f6964223a202239633661386163613562353730663236643932613332633431626663363162633865363632373833222c0a202022707269766174655f6b6579223a20222d2d2d2d2d424547494e2050524956415445204b45592d2d2d2d2d5c6e4d494945766749424144414e42676b71686b6947397730424151454641415343424b67776767536b41674541416f49424151445a6644614c6275327835774f445c6e484a546e2f7176462f774e786e477875307045537577785854716d4f3748626f73663047484b43394c39566b6c2b626a544455354e546359676569636d63384e5c6e496276343854523945644a544133794c456d4a7a6d66674973512f714f54557263454d516c2b552f646b4a41796d416c782b435734424a61334b397a31444a425c6e3456374b39574c61507a376a6e6f625669686656795751506a32345035395452366b31506b53776752445147704f79513632457439376a5347555478494f316f5c6e394e76517a75562f4535437a4346682f372b6251464a3232487939346274344f424b6d6c776947755769614a4575674b4d6c6931454e74736c5869447466726f5c6e6d637861455373395044705068485866346e5644477265484f4d4c4b4346686641423030775a685461575a4a6f474a39566c514e4574315136346b32547a4d375c6e5a4a754d7861636e41674d4241414543676745414c62574c69546168516e69354a6a39466c4a545436574d3169425647504f79496a52552b2f4d4b4e704870535c6e61346d74456a484748723045376f563267324d713949455975552f6b59625635374e71674e53774d7968534b7a654f33733073443469514547312b4c5a7344725c6e5364766f584835774d697861744639555964786d655939536a454a42706568395035647359742b384450367036784551616155436157355156327a66787675685c6e5030543071686b38563273534c64784a454144325a32614431494748746b417943384c696f31386c457463555a636233347470387a556f6f794d6a4d475667725c6e493037714a52494a635a4e53315a4d48416979476e2f75463745324d6f52517457325373476a484f7171636f78717074776657453831627964344b635069476c5c6e31646c5573676e61344d66653831344f3778573977467a585736307a4f72364d2b4b6275327a7a6a30514b4267514437555952646a3653594256716a662f504a5c6e4b613372635a7139494262434e526b4c6234692b7a506b784e6251502f793733765a4a3443587a455368543637334242427a6b35744e386164533152554c436d5c6e4f7163516a4b79327949657366352b45744d62632b4331734c6c48705873737a5733794f5056536a6850557444587050504f42577a3537775a3146556e3644775c6e357074534e5171414377424b6c424c643534504f654c706849774b42675144646956704e6c6f336266396747372b7077683173543939494d374965556e6a69535c6e5a425a5161786e6e312f3948716d48584f6d4a2b6b2b6b61303468313138474947395576745876426d37315a466f5875545234307453676d62615548787170625c6e354a787039544f76715353636c6d6c6f576230572f72614c534a774f774d7465686f6f76307a6f686e422f4f50386533306c644d4f4e7032647a7776383767375c6e4c48706e485235634c514b42675144446a61755562626769506c42483173456f4c314651576561513851346b63644b61446d42324c754a63417a436f486555375c6e436e7956414c546675394656624d694a49516a4c4f553038746837634868424757473830746e4753444c6c645a5455487575376564424a4d456b4c5564316c675c6e44666a2b6151535a39465165695655356f4f486a53737965765a59515a65474363623438476c2b67514738716d4d7552645a73664a74764878774b42674857735c6e4c70354632535839653062384375416f315a542b72734555706c4f6e303037584151394952474e6b31514642484756526174336e505174317a756368616e67635c6e714a6d463461324f527635614f31752f394d7030613159324b56472f45654272787a56302f445a544e744a434273315a31566d7768452f7069704d2f6a7761765c6e6d686b624c71614a6f6b395169346f316e513873703859445161494b364248755a7a6e384f704d6c416f4742414c6a355932496a552f534e416c694e456748665c6e58533551726434583672635139777a52373861457251424d636e364e5659432f424a534a4d6f344832666b5a385736617454544b75616d4a4e657276754a70645c6e77686d4a475a4e4e67634372334559634d366565476c455045616d6f36374d464b6d4b6945484257714d7037522f6f395061597a624b667774554c64546d71345c6e4749336250676d4954346a72442f3932452f414c427544475c6e2d2d2d2d2d454e442050524956415445204b45592d2d2d2d2d5c6e222c0a202022636c69656e745f656d61696c223a2022636c6f756468736d2d7161406d616e746c652d3338313330322e69616d2e67736572766963656163636f756e742e636f6d222c0a202022636c69656e745f6964223a2022313035393734343937353938323137383437323831222c0a202022617574685f757269223a202268747470733a2f2f6163636f756e74732e676f6f676c652e636f6d2f6f2f6f61757468322f61757468222c0a202022746f6b656e5f757269223a202268747470733a2f2f6f61757468322e676f6f676c65617069732e636f6d2f746f6b656e222c0a202022617574685f70726f76696465725f783530395f636572745f75726c223a202268747470733a2f2f7777772e676f6f676c65617069732e636f6d2f6f61757468322f76312f6365727473222c0a202022636c69656e745f783530395f636572745f75726c223a202268747470733a2f2f7777772e676f6f676c65617069732e636f6d2f726f626f742f76312f6d657461646174612f783530392f636c6f756468736d2d71612534306d616e746c652d3338313330322e69616d2e67736572766963656163636f756e742e636f6d222c0a202022756e6976657273655f646f6d61696e223a2022676f6f676c65617069732e636f6d220a7d0a",
			"HsmAddress": "0x14E4FF2909EEB2bc7379bcfF8263f04671B0afDA",
			"HsmAPIName": "projects/mantle-381302/locations/global/keyRings/qa/cryptoKeys/sequencer-qa/cryptoKeyVersions/1"
		},
		"RPCConfig": {
			"ListenAddr": "0.0.0.0",
			"ListenPort": 6545,
			"EnableAdmin": true
		},
		"LogConfig": {
			"Level": "info",
			"Color": false,
			"Format": "text"
		},
		"MetricsConfig": {
			"Enabled": true,
			"ListenAddr": "0.0.0.0",
			"ListenPort": 7302
		},
		"PprofConfig": {
			"Enabled": true,
			"ListenAddr": "0.0.0.0",
			"ListenPort": 6065
		},
		"CompressorConfig": {
			"TargetL1TxSizeBytes": 100000,
			"TargetNumFrames": 1,
			"ApproxComprRatio": 0.4,
			"Kind": "ratio"
		},
		"EigenLogConfig": {
			"Path": "",
			"Prefix": "",
			"FileLevel": "",
			"StdLevel": ""
		},
		"EigenDAConfig": {
			"RPC": "disperser-holesky.eigenda.xyz:443",
			"StatusQueryRetryInterval": 5000000000,
			"StatusQueryTimeout": 1200000000000
		}
	}`
	cfg := CLIConfig{}
	_ = json.Unmarshal([]byte(cfgData), &cfg)

	batchSubmitter, err := NewBatchSubmitterFromCLIConfig(cfg, log.New(), metrics.NewMetrics("test"))

	hash, _ := hexutil.Decode("0x4d175ff8a702034856bc7afb5adbd4317abfe64b18101f68d6b94570ce94d93c")
	hash2 := "TRdf+KcCA0hWvHr7WtvUMXq/5ksYEB9o1rlFcM6U2Tw="
	fmt.Println("hash == hash2", base64.StdEncoding.EncodeToString(hash) == hash2)
	data, err := batchSubmitter.eigenDA.RetrieveBlob(context.Background(), hash, 398)

	out := []eth.Data{}

	data = data[:]
	err = rlp.DecodeBytes(data, &out)
	if err != nil {
		log.Error("Decode retrieval frames in error,skip wrong data", "err", err)
		return
	}

	for _, d := range out {
		fs, err := ParseFrames(d)
		if err != nil {
			t.Error(err)
			return
		}
		for _, f := range fs {
			fmt.Println("frame:", hexutil.Encode(f.ID[:]), f.FrameNumber, len(f.Data))
		}
	}

}

func TestDecode(t *testing.T) {
	data, _ := base64.StdEncoding.DecodeString("fItdiqlPZyEjwyqJvLwlvuzkjYQ7qcnjI6DXooHht5U=")
	fmt.Println(hexutil.Encode(data))
	//data, _ = base64.StdEncoding.DecodeString("Yzg2NjQ4ZmM3YjY2ZWI5Mjk0NDk4ODhlNDY0YjE0Y2Q0NTEzOTMxM2Q5MjRlOGYxMThkNWMwZmVmYzQxMWZiOS0zMTM3MzEzNzM1MzczODM3MzgzMDMwMzgzNzM2MzEzMjM3MzgzMjJmMzAyZjMzMzMyZjMxMmYzMzMzMmZlM2IwYzQ0Mjk4ZmMxYzE0OWFmYmY0Yzg5OTZmYjkyNDI3YWU0MWU0NjQ5YjkzNGNhNDk1OTkxYjc4NTJiODU1")
	//fmt.Println(string(data))
}
