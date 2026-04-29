package apis

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/common"
)

type Faucet interface {
	RequestMNT(ctx context.Context, addr common.Address, amount eth.ETH) error
	Balance(ctx context.Context) (eth.ETH, error)
}
