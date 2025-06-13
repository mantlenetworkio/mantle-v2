package genesis

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/tracing"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/triedb"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
)

func TestWipePredeployStorage(t *testing.T) {
	sdb := state.NewDatabase(triedb.NewDatabase(rawdb.NewMemoryDatabase(), &triedb.Config{Preimages: true}), nil)
	stateDB, err := state.New(types.EmptyRootHash, sdb)
	require.NoError(t, err)

	storeVal := common.Hash{31: 0xff}

	for _, addr := range predeploys.Predeploys {
		a := *addr
		stateDB.SetState(a, storeVal, storeVal)
		stateDB.SetBalance(a, uint256.NewInt(99), tracing.BalanceMint)
		stateDB.SetNonce(a, 99, tracing.NonceChangeUnspecified)
	}

	root, err := stateDB.Commit(0, false, false)
	require.NoError(t, err)

	err = stateDB.Database().TrieDB().Commit(root, true)
	require.NoError(t, err)

	require.NoError(t, WipePredeployStorage(stateDB))

	for _, addr := range predeploys.Predeploys {
		a := *addr
		if FrozenStoragePredeploys[a] {
			require.Equal(t, storeVal, stateDB.GetState(a, storeVal))
		} else {
			require.Equal(t, common.Hash{}, stateDB.GetState(a, storeVal))
		}
		require.Equal(t, big.NewInt(99), stateDB.GetBalance(a))
		require.Equal(t, uint64(99), stateDB.GetNonce(a))
	}
}
