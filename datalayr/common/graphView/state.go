package graphView

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common"
)

type Registrant struct {
	Address         common.Address
	Socket          string
	PubkeyG1        *bn254.G1Affine
	PubkeyG2        *bn254.G2Affine
	FromBlockNumber uint32
	ToBlockNumber   uint32
}

type TotalStake struct {
	ToBlockNumber uint64
	QuorumStakes  []*big.Int
	Index         uint32
}

type TotalOperator struct {
	ToBlockNumber uint64
	Count         uint64
	AggPubKeyHash [32]byte
	AggPubKey     *bn254.G1Affine
	Index         uint32
}
