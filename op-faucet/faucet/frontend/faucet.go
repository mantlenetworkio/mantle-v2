package frontend

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	ftypes "github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/types"
	"github.com/ethereum-optimism/optimism/op-service/apis"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type FaucetBackend interface {
	ChainID() eth.ChainID
	RequestMNT(ctx context.Context, request *ftypes.FaucetRequest) error
	Balance() (eth.ETH, error)
	Register(addr common.Address) (bool, error)
	Eligibility(addr common.Address) (*ftypes.EligibilityResult, error)
}

type FaucetFrontend struct {
	b FaucetBackend
}

var _ apis.Faucet = (*FaucetFrontend)(nil)

func NewFaucetFrontend(b FaucetBackend) *FaucetFrontend {
	return &FaucetFrontend{b: b}
}

func (f *FaucetFrontend) ChainID(ctx context.Context) (eth.ChainID, error) {
	return f.b.ChainID(), nil
}

func (f *FaucetFrontend) RequestMNT(ctx context.Context, addr common.Address, amount eth.ETH) error {
	info := rpc.PeerInfoFromContext(ctx)
	request := &ftypes.FaucetRequest{
		RpcUser: &info,
		Target:  addr,
		Amount:  amount,
	}
	return f.b.RequestMNT(ctx, request)
}

func (f *FaucetFrontend) Balance(ctx context.Context) (eth.ETH, error) {
	return f.b.Balance()
}

func (f *FaucetFrontend) Register(ctx context.Context, addr common.Address) (bool, error) {
	return f.b.Register(addr)
}

func (f *FaucetFrontend) Eligibility(ctx context.Context, addr common.Address) (*ftypes.EligibilityResult, error) {
	return f.b.Eligibility(addr)
}
