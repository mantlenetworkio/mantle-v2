package main

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (d *simChainContext) CurrentHeader() *types.Header {
	return d.head
}

func (d *simChainContext) GetHeaderByHash(hash common.Hash) *types.Header {
	if hash == d.head.Hash() {
		return d.head
	}
	panic(fmt.Errorf("GetHeaderByHash not supported, cannot fetch %s", hash))
}

func (d *simChainContext) GetHeaderByNumber(n uint64) *types.Header {
	if n == d.head.Number.Uint64() {
		return d.head
	}
	panic(fmt.Errorf("GetHeaderByNumber not supported, cannot fetch %d", n))
}
