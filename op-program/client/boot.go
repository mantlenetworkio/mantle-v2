package client

import (
	"encoding/binary"
	"encoding/json"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/da"
	"github.com/ethereum-optimism/optimism/op-program/preimage"
)

const (
	L1HeadLocalIndex preimage.LocalIndexKey = iota + 1
	L2HeadLocalIndex
	L2ClaimLocalIndex
	L2ClaimBlockNumberLocalIndex
	L2ChainConfigLocalIndex
	RollupConfigLocalIndex
	RetrieverTimeout         = 60 * time.Second
	DataStorePollingDuration = 1 * time.Second
	// RetrieverSocket It will be modified into a configuration item when fraud proof was migrated later.,
	RetrieverSocket = ""
	// GraphProvider It will be modified into a configuration item when fraud proof was migrated later.,
	GraphProvider = ""
)

type BootInfo struct {
	L1Head             common.Hash
	L2Head             common.Hash
	L2Claim            common.Hash
	L2ClaimBlockNumber uint64
	L2ChainConfig      *params.ChainConfig
	RollupConfig       *rollup.Config
	DatastoreConfig    *da.MantleDataStoreConfig
}

type oracleClient interface {
	Get(key preimage.Key) []byte
}

type BootstrapClient struct {
	r oracleClient
}

func NewBootstrapClient(r oracleClient) *BootstrapClient {
	return &BootstrapClient{r: r}
}

func (br *BootstrapClient) BootInfo() *BootInfo {
	l1Head := common.BytesToHash(br.r.Get(L1HeadLocalIndex))
	l2Head := common.BytesToHash(br.r.Get(L2HeadLocalIndex))
	l2Claim := common.BytesToHash(br.r.Get(L2ClaimLocalIndex))
	l2ClaimBlockNumber := binary.BigEndian.Uint64(br.r.Get(L2ClaimBlockNumberLocalIndex))
	l2ChainConfig := new(params.ChainConfig)
	err := json.Unmarshal(br.r.Get(L2ChainConfigLocalIndex), &l2ChainConfig)
	if err != nil {
		panic("failed to bootstrap l2ChainConfig")
	}
	rollupConfig := new(rollup.Config)
	err = json.Unmarshal(br.r.Get(RollupConfigLocalIndex), rollupConfig)
	if err != nil {
		panic("failed to bootstrap rollup config")
	}
	daCfg := &da.MantleDataStoreConfig{
		RetrieverSocket:          RetrieverSocket,
		GraphProvider:            GraphProvider,
		RetrieverTimeout:         RetrieverTimeout,
		DataStorePollingDuration: DataStorePollingDuration,
	}
	return &BootInfo{
		L1Head:             l1Head,
		L2Head:             l2Head,
		L2Claim:            l2Claim,
		L2ClaimBlockNumber: l2ClaimBlockNumber,
		L2ChainConfig:      l2ChainConfig,
		RollupConfig:       rollupConfig,
		DatastoreConfig:    daCfg,
	}
}
