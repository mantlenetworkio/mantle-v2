package client

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
)

// DialRollupClientWithTimeout attempts to dial the RPC provider using the provided
// URL. If the dial doesn't complete within defaultDialTimeout seconds, this
// method will return an error.
func DialRollupClientWithTimeout(ctx context.Context, url string, timeout time.Duration) (*RollupClient, error) {
	ctxt, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	rpcCl, err := rpc.DialContext(ctxt, url)
	if err != nil {
		return nil, err
	}

	return NewRollupClient(NewBaseRPCClient(rpcCl)), nil
}
