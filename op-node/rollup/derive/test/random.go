package test

import (
	"math/rand"

	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
)

// RandomL2Block returns a random block whose first transaction is a random
// L1 Info Deposit transaction.
func RandomL2Block(rng *rand.Rand, txCount int) (*types.Block, []*types.Receipt) {
	l1Block := types.NewBlock(testutils.RandomHeader(rng),
		&types.Body{}, nil, trie.NewStackTrie(nil), types.DefaultBlockConfig)
	l1InfoTx, err := derive.L1InfoDeposit(0, eth.BlockToInfo(l1Block), eth.SystemConfig{}, testutils.RandomBool(rng))
	if err != nil {
		panic("L1InfoDeposit: " + err.Error())
	}
	return testutils.RandomBlockPrependTxs(rng, txCount, types.NewTx(l1InfoTx))
}
