package genesis

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

var (
	// BedrockTransitionBlockExtraData represents the extradata
	// set in the very first bedrock block. This value must be
	// less than 32 bytes long or it will create an invalid block.
	BedrockTransitionBlockExtraData = []byte("BEDROCK")
)

const (
	// defaultGasLimit represents the default gas limit for a genesis block.
	defaultGasLimit  = 30_000_000
	defaultL2BaseFee = 1_000_000_000
)

// NewL2Genesis will create a new L2 genesis
func NewL2Genesis(config *DeployConfig, block *types.Block) (*core.Genesis, error) {
	if config.L2ChainID == 0 {
		return nil, errors.New("must define L2 ChainID")
	}

	eip1559Denom := config.EIP1559Denominator
	if eip1559Denom == 0 {
		eip1559Denom = 50
	}
	eip1559Elasticity := config.EIP1559Elasticity
	if eip1559Elasticity == 0 {
		eip1559Elasticity = 10
	}

	optimismChainConfig := params.ChainConfig{
		ChainID:                 new(big.Int).SetUint64(config.L2ChainID),
		HomesteadBlock:          big.NewInt(0),
		DAOForkBlock:            nil,
		DAOForkSupport:          false,
		EIP150Block:             big.NewInt(0),
		EIP155Block:             big.NewInt(0),
		EIP158Block:             big.NewInt(0),
		ByzantiumBlock:          big.NewInt(0),
		ConstantinopleBlock:     big.NewInt(0),
		PetersburgBlock:         big.NewInt(0),
		IstanbulBlock:           big.NewInt(0),
		MuirGlacierBlock:        big.NewInt(0),
		BerlinBlock:             big.NewInt(0),
		LondonBlock:             big.NewInt(0),
		ArrowGlacierBlock:       big.NewInt(0),
		GrayGlacierBlock:        big.NewInt(0),
		MergeNetsplitBlock:      big.NewInt(0),
		TerminalTotalDifficulty: big.NewInt(0),
		BedrockBlock:            new(big.Int).SetUint64(uint64(config.L2GenesisBlockNumber)),
		RegolithTime:            config.RegolithTime(block.Time()),
		Optimism: &params.OptimismConfig{
			EIP1559Denominator: eip1559Denom,
			EIP1559Elasticity:  eip1559Elasticity,
		},
	}

	mantleUpgradeChainConfig := params.GetUpgradeConfigForMantle(optimismChainConfig.ChainID)
	if mantleUpgradeChainConfig != nil {
		optimismChainConfig.BaseFeeTime = mantleUpgradeChainConfig.BaseFeeTime
		optimismChainConfig.BVMETHMintUpgradeTime = mantleUpgradeChainConfig.BVMETHMintUpgradeTime
		optimismChainConfig.MetaTxV2UpgradeTime = mantleUpgradeChainConfig.MetaTxV2UpgradeTime
		optimismChainConfig.MetaTxV3UpgradeTime = mantleUpgradeChainConfig.MetaTxV3UpgradeTime
		optimismChainConfig.ProxyOwnerUpgradeTime = mantleUpgradeChainConfig.ProxyOwnerUpgradeTime
		optimismChainConfig.MantleEverestTime = mantleUpgradeChainConfig.MantleEverestTime
		optimismChainConfig.MantleSkadiTime = mantleUpgradeChainConfig.MantleSkadiTime

		// active standard EVM version (shanghai/cancun/prague)  in mantle skadi time
		optimismChainConfig.ShanghaiTime = mantleUpgradeChainConfig.MantleSkadiTime
		optimismChainConfig.CancunTime = mantleUpgradeChainConfig.MantleSkadiTime
		optimismChainConfig.PragueTime = mantleUpgradeChainConfig.MantleSkadiTime
	}

	gasLimit := config.L2GenesisBlockGasLimit
	if gasLimit == 0 {
		gasLimit = defaultGasLimit
	}
	baseFee := config.L2GenesisBlockBaseFeePerGas
	if baseFee == nil {
		baseFee = newHexBig(params.InitialBaseFee)
	}
	difficulty := config.L2GenesisBlockDifficulty
	if difficulty == nil {
		difficulty = newHexBig(0)
	}

	// Ensure that the extradata is valid
	if size := len(BedrockTransitionBlockExtraData); size > 32 {
		return nil, fmt.Errorf("transition block extradata too long: %d", size)
	}

	genesis := &core.Genesis{
		Config:     &optimismChainConfig,
		Nonce:      uint64(config.L2GenesisBlockNonce),
		Timestamp:  block.Time(),
		ExtraData:  BedrockTransitionBlockExtraData,
		GasLimit:   uint64(gasLimit),
		Difficulty: difficulty.ToInt(),
		Mixhash:    config.L2GenesisBlockMixHash,
		Coinbase:   predeploys.SequencerFeeVaultAddr,
		Number:     uint64(config.L2GenesisBlockNumber),
		GasUsed:    uint64(config.L2GenesisBlockGasUsed),
		ParentHash: config.L2GenesisBlockParentHash,
		BaseFee:    baseFee.ToInt(),
		Alloc:      map[common.Address]types.Account{},
	}

	if optimismChainConfig.IsMantleSkadi(genesis.Timestamp) {
		genesis.BlobGasUsed = u64ptr(0)
		genesis.ExcessBlobGas = u64ptr(0)
		genesis.Alloc[params.BeaconRootsAddress] = types.Account{Nonce: 1, Code: params.BeaconRootsCode, Balance: common.Big0}
		genesis.Alloc[params.HistoryStorageAddress] = types.Account{Nonce: 1, Code: params.HistoryStorageCode, Balance: common.Big0}
	}

	return genesis, nil
}

// NewL1Genesis will create a new L1 genesis config
func NewL1Genesis(config *DeployConfig) (*core.Genesis, error) {
	if config.L1ChainID == 0 {
		return nil, errors.New("must define L1 ChainID")
	}

	chainConfig := params.ChainConfig{
		ChainID:             uint642Big(config.L1ChainID),
		HomesteadBlock:      big.NewInt(0),
		DAOForkBlock:        nil,
		DAOForkSupport:      false,
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		LondonBlock:         big.NewInt(0),
		ArrowGlacierBlock:   big.NewInt(0),
		GrayGlacierBlock:    big.NewInt(0),
	}

	if config.CliqueSignerAddress != (common.Address{}) {
		// warning: clique has an overly strict block header timestamp check against the system wallclock,
		// causing blocks to get scheduled as "future block" and not get mined instantly when produced.
		chainConfig.Clique = &params.CliqueConfig{
			Period: config.L1BlockTime,
			Epoch:  30000,
		}
	} else {
		chainConfig.MergeNetsplitBlock = big.NewInt(0)
		chainConfig.TerminalTotalDifficulty = big.NewInt(0)
	}

	gasLimit := config.L1GenesisBlockGasLimit
	if gasLimit == 0 {
		gasLimit = defaultGasLimit
	}
	baseFee := config.L1GenesisBlockBaseFeePerGas
	if baseFee == nil {
		baseFee = newHexBig(params.InitialBaseFee)
	}
	difficulty := config.L1GenesisBlockDifficulty
	if difficulty == nil {
		difficulty = newHexBig(1)
	}
	timestamp := config.L1GenesisBlockTimestamp
	if timestamp == 0 {
		timestamp = hexutil.Uint64(time.Now().Unix())
	}

	extraData := make([]byte, 0)
	if config.CliqueSignerAddress != (common.Address{}) {
		extraData = append(append(make([]byte, 32), config.CliqueSignerAddress[:]...), make([]byte, crypto.SignatureLength)...)
	}

	return &core.Genesis{
		Config:     &chainConfig,
		Nonce:      uint64(config.L1GenesisBlockNonce),
		Timestamp:  uint64(timestamp),
		ExtraData:  extraData,
		GasLimit:   uint64(gasLimit),
		Difficulty: difficulty.ToInt(),
		Mixhash:    config.L1GenesisBlockMixHash,
		Coinbase:   config.L1GenesisBlockCoinbase,
		Number:     uint64(config.L1GenesisBlockNumber),
		GasUsed:    uint64(config.L1GenesisBlockGasUsed),
		ParentHash: config.L1GenesisBlockParentHash,
		BaseFee:    baseFee.ToInt(),
		Alloc:      map[common.Address]core.GenesisAccount{},
	}, nil
}

func u64ptr(n uint64) *uint64 {
	return &n
}
