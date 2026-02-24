package bindings

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// MantleL2ToL1MessagePasser matches Mantle's L2ToL1MessagePasser ABI (includes ethValue).
type MantleL2ToL1MessagePasser struct {
	// Read-only functions
	MESSAGEVERSION func() TypedCall[uint16]                   `sol:"MESSAGE_VERSION"`
	MessageNonce   func() TypedCall[*big.Int]                 `sol:"messageNonce"`
	SentMessages   func(messageHash [32]byte) TypedCall[bool] `sol:"sentMessages"`
	Version        func() TypedCall[string]                   `sol:"version"`

	// Write functions
	Burn               func() TypedCall[any]                                                                         `sol:"burn"`
	InitiateWithdrawal func(ethValue *big.Int, target common.Address, gasLimit *big.Int, data []byte) TypedCall[any] `sol:"initiateWithdrawal"`
	Receive            func() TypedCall[any]                                                                         `sol:"receive"`
}
