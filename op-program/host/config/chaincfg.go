package config

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

var OPGoerliChainConfig = &params.ChainConfig{
	ChainID:                 big.NewInt(420),
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
	LondonBlock:             big.NewInt(4061224),
	ArrowGlacierBlock:       big.NewInt(4061224),
	GrayGlacierBlock:        big.NewInt(4061224),
	MergeNetsplitBlock:      big.NewInt(4061224),
	BedrockBlock:            big.NewInt(4061224),
	RegolithTime:            u64Ptr(0),
	TerminalTotalDifficulty: big.NewInt(0),
	Optimism: &params.OptimismConfig{
		EIP1559Elasticity:  10,
		EIP1559Denominator: 50,
	},
}

var L2ChainConfigsByName = map[string]*params.ChainConfig{
	"goerli": OPGoerliChainConfig,
}

func u64Ptr(v uint64) *uint64 {
	return &v
}
