package oracle

import (
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
)

func TestIsDifferenceSignificant(t *testing.T) {
	tests := []struct {
		name   string
		a      uint64
		b      uint64
		sig    float64
		expect bool
	}{
		{name: "test 1", a: 1, b: 1, sig: 0.05, expect: false},
		{name: "test 2", a: 4, b: 1, sig: 0.25, expect: true},
		{name: "test 3", a: 3, b: 1, sig: 0.1, expect: true},
		{name: "test 4", a: 4, b: 1, sig: 0.9, expect: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isDifferenceSignificant(tc.a, tc.b, tc.sig)
			if result != tc.expect {
				t.Fatalf("mismatch %s", tc.name)
			}
		})
	}
}

func newSimulatedBackend(key *ecdsa.PrivateKey) (*backends.SimulatedBackend, ethdb.Database) {
	var gasLimit uint64 = 9_000_000
	auth, _ := bind.NewKeyedTransactorWithChainID(key, big.NewInt(1337))
	genAlloc := make(core.GenesisAlloc)
	genAlloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(9223372036854775807)}
	db := rawdb.NewMemoryDatabase()
	sim := backends.NewSimulatedBackendWithDatabase(db, genAlloc, gasLimit)
	return sim, db
}
