package sync_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/ethereum-optimism/optimism/op-node/metrics"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
	smetrics "github.com/ethereum-optimism/optimism/op-service/metrics"

	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/require"
)

// generateFakeL2 creates a fake L2 chain with the following conditions:
// - The L2 chain is based off of the L1 chain
// - The actual L1 chain is the New L1 chain
// - Both heads are at the tip of their respective chains
func (c *syncStartTestCase) generateFakeL2(t *testing.T) (*testutils.FakeChainSource, rollup.Genesis) {
	t.Helper()
	log := testlog.Logger(t, log.LvlError)
	chain := testutils.NewFakeChainSource([]string{c.L1, c.NewL1}, []string{c.L2}, int(c.GenesisL1Num), log)
	chain.SetL2Head(len(c.L2) - 1)
	genesis := testutils.FakeGenesis(c.GenesisL1, c.GenesisL2, int(c.GenesisL1Num))
	chain.ReorgL1()
	for i := 0; i < len(c.NewL1)-1; i++ {
		chain.AdvanceL1()
	}
	return chain, genesis

}

func runeToHash(id rune) common.Hash {
	var h common.Hash
	copy(h[:], string(id))
	return h
}

type syncStartTestCase struct {
	Name string

	L1    string // L1 Chain prior to a re-org or other change
	L2    string // L2 Chain that follows from L1Chain
	NewL1 string // New L1 chain

	PreFinalizedL2 rune
	PreSafeL2      rune

	GenesisL1    rune
	GenesisL1Num uint64
	GenesisL2    rune

	SeqWindowSize uint64
	SafeL2Head    rune
	UnsafeL2Head  rune
	ExpectedErr   error
}

func refToRune(r eth.BlockID) rune {
	return rune(r.Hash.Bytes()[0])
}

func (c *syncStartTestCase) Run(t *testing.T) {
	chain, genesis := c.generateFakeL2(t)
	chain.SetL2Finalized(runeToHash(c.PreFinalizedL2))
	chain.SetL2Safe(runeToHash(c.PreSafeL2))

	cfg := &rollup.Config{
		Genesis:       genesis,
		SeqWindowSize: c.SeqWindowSize,
	}
	lgr := log.New()
	lgr.SetHandler(log.DiscardHandler())
	result, err := sync.FindL2Heads(context.Background(), cfg, chain, chain, lgr, &sync.Config{})
	if c.ExpectedErr != nil {
		require.ErrorIs(t, err, c.ExpectedErr, "expected error")
		return
	} else {
		require.NoError(t, err, "expected no error")
	}

	gotUnsafeHead := refToRune(result.Unsafe.ID())
	require.Equal(t, string(c.UnsafeL2Head), string(gotUnsafeHead), "Unsafe L2 Head not equal")

	gotSafeHead := refToRune(result.Safe.ID())
	require.Equal(t, string(c.SafeL2Head), string(gotSafeHead), "Safe L2 Head not equal")
}

func TestFindSyncStart(t *testing.T) {
	testCases := []syncStartTestCase{
		{
			Name:           "already synced",
			GenesisL1Num:   0,
			L1:             "ab",
			L2:             "AB",
			NewL1:          "ab",
			PreFinalizedL2: 'A',
			PreSafeL2:      'A',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'B',
			SeqWindowSize:  2,
			SafeL2Head:     'A',
			ExpectedErr:    nil,
		},
		{
			Name:           "small reorg long chain",
			GenesisL1Num:   0,
			L1:             "abcdefgh",
			L2:             "ABCDEFGH",
			NewL1:          "abcdefgx",
			PreFinalizedL2: 'B',
			PreSafeL2:      'H',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'G',
			SeqWindowSize:  2,
			SafeL2Head:     'C',
			ExpectedErr:    nil,
		},
		{
			Name:           "L1 Chain ahead",
			GenesisL1Num:   0,
			L1:             "abcdef",
			L2:             "ABCDE",
			NewL1:          "abcdef",
			PreFinalizedL2: 'A',
			PreSafeL2:      'D',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'E',
			SeqWindowSize:  2,
			SafeL2Head:     'A',
			ExpectedErr:    nil,
		},
		{
			Name:           "L2 Chain ahead after reorg",
			GenesisL1Num:   0,
			L1:             "abcxyz",
			L2:             "ABCXYZ",
			NewL1:          "abcx",
			PreFinalizedL2: 'B',
			PreSafeL2:      'X',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'Z',
			SeqWindowSize:  2,
			SafeL2Head:     'B',
			ExpectedErr:    nil,
		},
		{
			Name:           "genesis",
			GenesisL1Num:   0,
			L1:             "a",
			L2:             "A",
			NewL1:          "a",
			PreFinalizedL2: 'A',
			PreSafeL2:      'A',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'A',
			SeqWindowSize:  2,
			SafeL2Head:     'A',
			ExpectedErr:    nil,
		},
		{
			Name:           "reorg one step back",
			GenesisL1Num:   0,
			L1:             "abcdefg",
			L2:             "ABCDEFG",
			NewL1:          "abcdefx",
			PreFinalizedL2: 'A',
			PreSafeL2:      'E',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'F',
			SeqWindowSize:  3,
			SafeL2Head:     'A',
			ExpectedErr:    nil,
		},
		{
			Name:           "reorg two steps back, clip genesis and finalized",
			GenesisL1Num:   0,
			L1:             "abc",
			L2:             "ABC",
			PreFinalizedL2: 'A',
			PreSafeL2:      'B',
			NewL1:          "axy",
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'A',
			SeqWindowSize:  2,
			SafeL2Head:     'A',
			ExpectedErr:    nil,
		},
		{
			Name:           "reorg three steps back",
			GenesisL1Num:   0,
			L1:             "abcdefgh",
			L2:             "ABCDEFGH",
			NewL1:          "abcdexyz",
			PreFinalizedL2: 'A',
			PreSafeL2:      'D',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'E',
			SeqWindowSize:  2,
			SafeL2Head:     'A',
			ExpectedErr:    nil,
		},
		{
			Name:           "unexpected L1 chain",
			GenesisL1Num:   0,
			L1:             "abcdef",
			L2:             "ABCDEF",
			NewL1:          "xyzwio",
			PreFinalizedL2: 'A',
			PreSafeL2:      'B',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   0,
			SeqWindowSize:  2,
			ExpectedErr:    sync.WrongChainErr,
		},
		{
			Name:           "unexpected L2 chain",
			GenesisL1Num:   0,
			L1:             "abcdef",
			L2:             "ABCDEF",
			NewL1:          "xyzwio",
			PreFinalizedL2: 'A',
			PreSafeL2:      'B',
			GenesisL1:      'a',
			GenesisL2:      'X',
			UnsafeL2Head:   0,
			SeqWindowSize:  2,
			ExpectedErr:    sync.WrongChainErr,
		},
		{
			Name:           "offset L2 genesis",
			GenesisL1Num:   3,
			L1:             "abcdefghi",
			L2:             "DEFGHI",
			NewL1:          "abcdefghi",
			PreFinalizedL2: 'E',
			PreSafeL2:      'H',
			GenesisL1:      'd',
			GenesisL2:      'D',
			UnsafeL2Head:   'I',
			SeqWindowSize:  2,
			SafeL2Head:     'E',
			ExpectedErr:    nil,
		},
		{
			Name:           "offset L2 genesis reorg",
			GenesisL1Num:   3,
			L1:             "abcdefgh",
			L2:             "DEFGH",
			NewL1:          "abcdxyzw",
			PreFinalizedL2: 'D',
			PreSafeL2:      'D',
			GenesisL1:      'd',
			GenesisL2:      'D',
			UnsafeL2Head:   'D',
			SeqWindowSize:  2,
			SafeL2Head:     'D',
			ExpectedErr:    nil,
		},
		{
			Name:           "reorg past offset genesis",
			GenesisL1Num:   3,
			L1:             "abcdefgh",
			L2:             "DEFGH",
			NewL1:          "abxyzwio",
			PreFinalizedL2: 'D',
			PreSafeL2:      'D',
			GenesisL1:      'd',
			GenesisL2:      'D',
			UnsafeL2Head:   0,
			SeqWindowSize:  2,
			SafeL2Head:     'D',
			ExpectedErr:    sync.WrongChainErr,
		},
		{
			// FindL2Heads() keeps walking back to safe head after finding canonical unsafe head
			// TooDeepReorgErr must not be raised
			Name:           "long traverse to safe head",
			GenesisL1Num:   0,
			L1:             "abcdefgh",
			L2:             "ABCDEFGH",
			NewL1:          "abcdefgx",
			PreFinalizedL2: 'B',
			PreSafeL2:      'B',
			GenesisL1:      'a',
			GenesisL2:      'A',
			UnsafeL2Head:   'G',
			SeqWindowSize:  1,
			SafeL2Head:     'B',
			ExpectedErr:    nil,
		},
		{
			// L2 reorg is too deep
			Name:           "reorg too deep",
			GenesisL1Num:   0,
			L1:             "abcdefgh",
			L2:             "ABCDEFGH",
			NewL1:          "abijklmn",
			PreFinalizedL2: 'B',
			PreSafeL2:      'B',
			GenesisL1:      'a',
			GenesisL2:      'A',
			SeqWindowSize:  1,
			ExpectedErr:    sync.TooDeepReorgErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, testCase.Run)
	}
}

var rcfgData = `
{
	"genesis": {
	  "l1": {
		"hash": "0x4b77c36e36d655483cdbf0c7119dfff3ff21e415b93831184f264189d34c4413",
		"number": 5485455
	  },
	  "l2": {
		"hash": "0xf658f235e93299bb391c179ac2c62fcbda1b91f087034ee4757c29c7c2b70929",
		"number": 4983
	  },
	  "l2_time": 1710436356,
	  "system_config": {
		"batcherAddr": "0x10eb1276305f45ff0e89f81d0d304eb2a64b82ee",
		"overhead": "0x0000000000000000000000000000000000000000000000000000000000000834",
		"scalar": "0x00000000000000000000000000000000000000000000000000000000000f4240",
		"gasLimit": 1125899906842624,
		"baseFee": 1000000000
	  }
	},
	"block_time": 2,
	"max_sequencer_drift": 600,
	"seq_window_size": 3600,
	"channel_timeout": 300,
	"l1_chain_id": 11155111,
	"l2_chain_id": 5003003,
	"regolith_time": 0,
	"batch_inbox_address": "0x16fcf349b60262C4A87350757085784E39804810",
	"deposit_contract_address": "0xc54a00a4abeba64e6fdbea4b6521e79a4ae5722a",
	"l1_system_config_address": "0xd5e98bb1c7df7515c73dacb293cc1fcb219a66f4",
	"mantle_da_switch": true,
	"datalayr_service_manager_addr": "0x0185eBB3Fc4c101eCdDB753aB6Eb38252882E62a"
  }
`

func TestFindL2Heads(t *testing.T) {
	type args struct {
		ctx     context.Context
		cfg     *rollup.Config
		l1      sync.L1Chain
		l2      sync.L2Chain
		lgr     log.Logger
		syncCfg *sync.Config
	}

	rcfg := &rollup.Config{}
	ctx := context.Background()
	log := log.New("t1")
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	factory := smetrics.With(registry)
	m := metrics.NewMetrics("default")

	L1SourceCache := metrics.NewCacheMetrics(factory, "t1", "l1_source_cache", "L1 Source cache")
	L2SourceCache := metrics.NewCacheMetrics(factory, "t1", "l2_source_cache", "L2 Source cache")
	l1ClientCfg := &sources.L1ClientConfig{
		EthClientConfig: sources.EthClientConfig{
			MaxRequestsPerBatch:   1000,
			MaxConcurrentRequests: 100,
			ReceiptsCacheSize:     1000,
			TransactionsCacheSize: 1000,
			HeadersCacheSize:      1000,
			PayloadsCacheSize:     1000,
			TrustRPC:              true,
			RPCProviderKind:       sources.RPCKindAny,
		},
		L1BlockRefsCacheSize: 1000,
	}
	l2ClientCfg := &sources.EngineClientConfig{
		L2ClientConfig: sources.L2ClientConfig{
			EthClientConfig: sources.EthClientConfig{
				MaxRequestsPerBatch:   1000,
				MaxConcurrentRequests: 100,
				ReceiptsCacheSize:     1000,
				TransactionsCacheSize: 1000,
				HeadersCacheSize:      1000,
				PayloadsCacheSize:     1000,
				TrustRPC:              true,
				RPCProviderKind:       sources.RPCKindAny,
			},
			L2BlockRefsCacheSize: 1000,
			L1ConfigsCacheSize:   1000,
			RollupCfg:            rcfg,
		},
	}
	fmt.Println("rollup config", json.Unmarshal([]byte(rcfgData), rcfg))

	l1Rpc, err := client.NewRPC(ctx, log, "http://127.0.0.1:9875")
	fmt.Println("new rpc", l1Rpc, err)
	l2Rpc, err := client.NewRPC(ctx, log, "http://127.0.0.1:9874")
	fmt.Println("new rpc", l2Rpc, err)

	l1, err := sources.NewL1Client(
		client.NewInstrumentedRPC(l1Rpc, m), log, L1SourceCache, l1ClientCfg)
	l2, err := sources.NewEngineClient(
		client.NewInstrumentedRPC(l2Rpc, m), log, L2SourceCache, l2ClientCfg)

	tests := []struct {
		name       string
		args       args
		wantResult *sync.FindHeadsResult
		wantErr    bool
	}{
		{
			name: "test",
			args: args{
				ctx:     ctx,
				cfg:     rcfg,
				l1:      l1,
				l2:      l2,
				lgr:     log,
				syncCfg: &sync.Config{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := sync.FindL2Heads(tt.args.ctx, tt.args.cfg, tt.args.l1, tt.args.l2, tt.args.lgr, tt.args.syncCfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindL2Heads() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			fmt.Println(gotResult)
		})
	}
}
