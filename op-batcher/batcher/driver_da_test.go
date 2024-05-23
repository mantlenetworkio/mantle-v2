package batcher

import (
	"math/big"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/Layr-Labs/datalayr/common/graphView"
	"github.com/Layr-Labs/datalayr/common/logging"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-batcher/metrics"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
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
