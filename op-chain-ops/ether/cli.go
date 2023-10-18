package ether

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/state"
)

// getBVMMNTTotalSupply returns BVM MNT's total supply by reading
// the appropriate storage slot.
func getBVMMNTTotalSupply(db *state.StateDB) *big.Int {
	key := getBVMMNTTotalSupplySlot()
	return db.GetState(OVMETHAddress, key).Big()
}

func getBVMMNTTotalSupplySlot() common.Hash {
	position := common.Big2
	key := common.BytesToHash(common.LeftPadBytes(position.Bytes(), 32))
	return key
}

func GetBVMMNTTotalSupplySlot() common.Hash {
	return getBVMMNTTotalSupplySlot()
}

// GetBVMMNTBalance gets a user's OVM ETH balance from state by querying the
// appropriate storage slot directly.
func GetBVMMNTBalance(db *state.StateDB, addr common.Address) *big.Int {
	return db.GetState(OVMETHAddress, CalcBVMETHStorageKey(addr)).Big()
}
