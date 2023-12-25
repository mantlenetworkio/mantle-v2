package crossdomain

import (
	"math/big"
)

// Params contains the configuration parameters used for verifying
// the integrity of the migration.
type Params struct {
	// ExpectedSupplyDelta is the expected delta between the total supply of OVM ETH,
	// and ETH we were able to migrate. This is used to account for supply bugs in
	//previous regenesis events.
	ExpectedSupplyDelta *big.Int
}

var ParamsByChainID = map[int]*Params{
	1: { // mainnet
		new(big.Int).SetUint64(0),
	},
	5: { // Goerli
		new(big.Int).SetUint64(0),
	},
	11155111: { // Sepolia
		new(big.Int).SetUint64(0),
	},
	31337: { //Devnet
		new(big.Int).Mul(new(big.Int).SetInt64(1e18), new(big.Int).SetInt64(-600000)),
	},
}
