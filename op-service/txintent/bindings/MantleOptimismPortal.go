package bindings

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
)

type MantleProvenWithdrawal struct {
	OutputRoot    [32]byte
	Timestamp     *big.Int
	L2OutputIndex *big.Int
}

type MantleWithdrawalTransaction struct {
	Nonce    *big.Int
	Sender   common.Address
	Target   common.Address
	MNTValue *big.Int
	ETHValue *big.Int
	GasLimit *big.Int
	Data     []byte
}

type MantleOutputRootProof struct {
	Version                  [32]byte
	StateRoot                [32]byte
	MessagePasserStorageRoot [32]byte
	LatestBlockhash          [32]byte
}

// MantleOptimismPortal matches Mantle's OptimismPortal ABI.
type MantleOptimismPortal struct {
	// Read-only functions
	GUARDIAN             func() TypedCall[common.Address]                                `sol:"GUARDIAN"`
	L1MNTAddress         func() TypedCall[common.Address]                                `sol:"L1_MNT_ADDRESS"`
	L2Oracle             func() TypedCall[common.Address]                                `sol:"L2_ORACLE"`
	SystemConfig         func() TypedCall[common.Address]                                `sol:"SYSTEM_CONFIG"`
	FinalizedWithdrawals func(withdrawalHash [32]byte) TypedCall[bool]                   `sol:"finalizedWithdrawals"`
	L2Sender             func() TypedCall[common.Address]                                `sol:"l2Sender"`
	MinimumGasLimit      func(byteCount uint64) TypedCall[uint64]                        `sol:"minimumGasLimit"`
	Paused               func() TypedCall[bool]                                          `sol:"paused"`
	ProvenWithdrawals    func(withdrawalHash [32]byte) TypedCall[MantleProvenWithdrawal] `sol:"provenWithdrawals"`
	Version              func() TypedCall[string]                                        `sol:"version"`
	IsOutputFinalized    func(l2OutputIndex *big.Int) TypedCall[bool]                    `sol:"isOutputFinalized"`

	// Write functions
	DepositTransaction            func(ethTxValue eth.ETH, mntValue eth.ETH, to common.Address, mntTxValue eth.ETH, gasLimit uint64, isCreation bool, data []byte) TypedCall[any] `sol:"depositTransaction"`
	DonateETH                     func() TypedCall[any]                                                                                                                           `sol:"donateETH"`
	FinalizeWithdrawalTransaction func(tx MantleWithdrawalTransaction) TypedCall[any]                                                                                             `sol:"finalizeWithdrawalTransaction"`
	ProveWithdrawalTransaction    func(tx MantleWithdrawalTransaction, l2OutputIndex *big.Int, outputRootProof MantleOutputRootProof, withdrawalProof [][]byte) TypedCall[any]    `sol:"proveWithdrawalTransaction"`
	Pause                         func() TypedCall[any]                                                                                                                           `sol:"pause"`
	Unpause                       func() TypedCall[any]                                                                                                                           `sol:"unpause"`
	Receive                       func() TypedCall[any]                                                                                                                           `sol:"receive"`
}
