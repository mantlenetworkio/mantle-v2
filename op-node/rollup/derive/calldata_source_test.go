package derive

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/da"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
)

type testTx struct {
	to      *common.Address
	dataLen int
	author  *ecdsa.PrivateKey
	good    bool
	value   int
}

func (tx *testTx) Create(t *testing.T, signer types.Signer, rng *rand.Rand) *types.Transaction {
	t.Helper()
	out, err := types.SignNewTx(tx.author, signer, &types.DynamicFeeTx{
		ChainID:   signer.ChainID(),
		Nonce:     0,
		GasTipCap: big.NewInt(2 * params.GWei),
		GasFeeCap: big.NewInt(30 * params.GWei),
		Gas:       100_000,
		To:        tx.to,
		Value:     big.NewInt(int64(tx.value)),
		Data:      testutils.RandomData(rng, tx.dataLen),
	})
	require.NoError(t, err)
	return out
}

type calldataTest struct {
	name string
	txs  []testTx
}

// TestDataFromEVMTransactions creates some transactions from a specified template and asserts
// that DataFromEVMTransactions properly filters and returns the data from the authorized transactions
// inside the transaction set.
func TestDataFromEVMTransactions(t *testing.T) {
	inboxPriv := testutils.RandomKey()
	batcherPriv := testutils.RandomKey()
	cfg := &rollup.Config{
		L1ChainID:         big.NewInt(100),
		BatchInboxAddress: crypto.PubkeyToAddress(inboxPriv.PublicKey),
	}
	batcherAddr := crypto.PubkeyToAddress(batcherPriv.PublicKey)

	altInbox := testutils.RandomAddress(rand.New(rand.NewSource(1234)))
	altAuthor := testutils.RandomKey()

	testCases := []calldataTest{
		{
			name: "simple",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: batcherPriv, good: true}},
		},
		{
			name: "other inbox",
			txs:  []testTx{{to: &altInbox, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "other author",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: altAuthor, good: false}}},
		{
			name: "inbox is author",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: inboxPriv, good: false}}},
		{
			name: "author is inbox",
			txs:  []testTx{{to: &batcherAddr, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "unrelated",
			txs:  []testTx{{to: &altInbox, dataLen: 1234, author: altAuthor, good: false}}},
		{
			name: "contract creation",
			txs:  []testTx{{to: nil, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "empty tx",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 0, author: batcherPriv, good: true}}},
		{
			name: "value tx",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, value: 42, author: batcherPriv, good: true}}},
		{
			name: "empty block", txs: []testTx{},
		},
		{
			name: "mixed txs",
			txs: []testTx{
				{to: &cfg.BatchInboxAddress, dataLen: 1234, value: 42, author: batcherPriv, good: true},
				{to: &cfg.BatchInboxAddress, dataLen: 3333, value: 32, author: altAuthor, good: false},
				{to: &cfg.BatchInboxAddress, dataLen: 2000, value: 22, author: batcherPriv, good: true},
				{to: &altInbox, dataLen: 2020, value: 12, author: batcherPriv, good: false},
			},
		},
		// TODO: test with different batcher key, i.e. when it's changed from initial config value by L1 contract
	}

	for i, tc := range testCases {
		rng := rand.New(rand.NewSource(int64(i)))
		signer := cfg.L1Signer()

		var expectedData []eth.Data
		var txs []*types.Transaction
		for i, tx := range tc.txs {
			txs = append(txs, tx.Create(t, signer, rng))
			if tx.good {
				expectedData = append(expectedData, txs[i].Data())
			}
		}

		out := DataFromEVMTransactions(cfg, batcherAddr, txs, testlog.Logger(t, log.LvlCrit))
		require.ElementsMatch(t, expectedData, out)
	}

}

func TestRLPEncodeDecodeEthData(t *testing.T) {
	var dataS = make([]eth.Data, 0)
	dataS = append(dataS,
		eth.Data(common.Hex2Bytes("test1")),
		eth.Data(common.Hex2Bytes("test2")),
		eth.Data(common.Hex2Bytes("test3")),
	)
	// encode
	bz, err := rlp.EncodeToBytes(dataS)
	require.NoError(t, err)

	// decode
	var dataL = make([]eth.Data, 0)
	err = rlp.DecodeBytes(bz, &dataL)
	require.NoError(t, err)
}

func TestRetrieveBlob(t *testing.T) {
	// da := eigenda.NewEigenDAClient(
	cfg := eigenda.Config{
		DisperserUrl:        "disperser-holesky.eigenda.xyz:443",
		ProxyUrl:            "http://127.0.0.1:3100",
		DisperseBlobTimeout: 20 * time.Minute,
		RetrieveBlobTimeout: 20 * time.Minute,
	}
	// 	log.New(context.Background()),
	// 	nil,
	// )

	calldata, _ := hex.DecodeString("ed1293040a20313594efb8d9721c7f000ba5dcc3e51a717e33662d318bb88c24896fda6473ff102218978a9a012202000128f2a442aa06de03010000f901d8f854f842a01805f5ef4df0dfc3d4ef78648e59b9888345e90ea6e8cdd5d662856782260517a000e4879b77e0e9542db302a559ef63944dcef02f5d1660bfb8f1793ba4b2c3938288dcccc480213740c6012137820400f9017f82b28d22f873eba0e6c0d26a4c0f30c76df3730f82e2f26d9015c43b90a61f260d0b0a8cb261c39d82000182416483268517a0c6cce1fadc0eaea692ce44586249c41b42befd001eb44363e7a7e130cafbe4bd008326856da0313594efb8d9721c7f000ba5dcc3e51a717e33662d318bb88c24896fda6473ffb901009bc91a9ff8246b9a3fd2efd49a48fa41add3eb78a208988f1497baae5837909faa4a8f58ed9cb9bd5bac32b7d2dd9483764d2af98b7752846f5677d4012864bcbbaa78187c8d4a7be2354515e8e25b4d7d9ee75b12581ca995437d6ca9e959d7767ef5a372b4587a4288aa16242217e1ab6252ac6193a88add75b6f7d9ca6f2f5125d73c4ebcadf3d82ab1f54d8fe2e91d4c193905b91fbbd9dd0fd30bd3fd2f5f2d5d7ff44d4baaabcb0f5e60d4217eceba6e3e4c2d160d9c7e78718d2018f26abb188a83feada8847ba19ad49eb97979a6e147eabf6a729b56903eda3722b574890eca80948cc8eff9eb9bd4037cc38b00f20136e66b756692734f403c337c820001")
	calldataFrame := &op_service.CalldataFrame{}
	err := proto.Unmarshal(calldata[1:], calldataFrame)
	if err != nil {
		return
	}
	frame := calldataFrame.Value.(*op_service.CalldataFrame_FrameRef)
	da := da.NewEigenDADataStore(context.Background(), log.New("t1"), &cfg, nil, nil)
	fmt.Printf("%x\n%x\n", frame.FrameRef.BatchHeaderHash, frame.FrameRef.Commitment)
	data, err := da.RetrieveBlob(frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex, nil)
	if err != nil {
		t.Errorf("RetrieveBlob err:%v", err)
		return
	}
	fmt.Printf("RetrieveBlob data:%x\n", data)

	data, err = da.RetrieveBlob(frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex, frame.FrameRef.Commitment)
	if err != nil {
		t.Errorf("RetrieveBlob err:%v", err)
		return
	}

	fmt.Printf("RetrieveBlob data:%x\n", data)

	data = data[:frame.FrameRef.BlobLength]
	outData := []eth.Data{}
	err = rlp.DecodeBytes(data, &outData)
	if err != nil {
		log.Error("Decode retrieval frames in error,skip wrong data", "err", err, "blobInfo", fmt.Sprintf("%x:%d", frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex))
		return
	}

	fmt.Printf("RetrieveBlob %d\n", len(outData))
}

func TestRetrieveBlobTx(t *testing.T) {
}

func TestRetrieveFromDaIndexer(t *testing.T) {
	eigenDA := eigenda.Config{
		ProxyUrl: "disperser-holesky.eigenda.xyz:443",
	}

	eigenDaSyncer := da.NewEigenDADataStore(context.Background(), log.New("t1"), &eigenDA, &da.MantleDataStoreConfig{
		MantleDaIndexerSocket: "127.0.0.1:32111",
		MantleDAIndexerEnable: true,
	}, nil)

	out := []eth.Data{}

	if eigenDaSyncer.IsDaIndexer() {
		data, err := eigenDaSyncer.RetrievalFramesFromDaIndexer("0x8494e3e2c70933fc69b82bc0a851f77716d385b52fa8f386df29b819c717be9b")
		if err != nil {
			fmt.Println("Retrieval frames from eigenDa indexer error", "err", err)
			return
		}
		outData := []eth.Data{}
		err = rlp.DecodeBytes(data, &outData)
		if err != nil {
			fmt.Println("Decode retrieval frames in error,skip wrong data", "err", err)
			return
		}
		out = append(out, outData...)

		fmt.Println(len(out))
		return
	}

}
