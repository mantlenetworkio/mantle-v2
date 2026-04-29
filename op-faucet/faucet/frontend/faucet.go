package frontend

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	ftypes "github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/types"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// SharedBackend is the backend for shared (non-chain-specific) operations.
type SharedBackend interface {
	Register(addr common.Address) (bool, error)
	Eligibility(addr common.Address) (*ftypes.EligibilityResult, error)
}

// ChainBackend is the backend for chain-specific operations.
type ChainBackend interface {
	RequestMNT(ctx context.Context, request *ftypes.FaucetRequest) error
	Balance() (eth.ETH, error)
}

// SharedFrontend exposes faucet_register and faucet_eligibility on the root path.
type SharedFrontend struct {
	b SharedBackend
}

func NewSharedFrontend(b SharedBackend) *SharedFrontend {
	return &SharedFrontend{b: b}
}

func (f *SharedFrontend) Register(ctx context.Context, addr common.Address) (bool, error) {
	return f.b.Register(addr)
}

func (f *SharedFrontend) Eligibility(ctx context.Context, addr common.Address) (*ftypes.EligibilityResult, error) {
	return f.b.Eligibility(addr)
}

// ChainFrontend exposes faucet_balance and faucet_requestMNT on /chain/{chainId}.
type ChainFrontend struct {
	b ChainBackend
}

func NewChainFrontend(b ChainBackend) *ChainFrontend {
	return &ChainFrontend{b: b}
}

func (f *ChainFrontend) RequestMNT(ctx context.Context, addr common.Address, amount eth.ETH) error {
	info := rpc.PeerInfoFromContext(ctx)
	request := &ftypes.FaucetRequest{
		RpcUser: &info,
		Target:  addr,
		Amount:  amount,
	}
	return f.b.RequestMNT(ctx, request)
}

func (f *ChainFrontend) Balance(ctx context.Context) (eth.ETH, error) {
	return f.b.Balance()
}
