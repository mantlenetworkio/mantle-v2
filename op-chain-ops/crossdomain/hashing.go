package crossdomain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// HashCrossDomainMessageV0 computes the pre bedrock cross domain messaging
// hashing scheme.
func HashCrossDomainMessageV0(
	target common.Address,
	sender common.Address,
	data []byte,
	nonce *big.Int,
) (common.Hash, error) {
	encoded, err := EncodeCrossDomainMessageV0(target, sender, data, nonce)
	if err != nil {
		return common.Hash{}, err
	}
	hash := crypto.Keccak256(encoded)
	return common.BytesToHash(hash), nil
}

// HashCrossDomainMessageV1 computes the first post bedrock cross domain
// messaging hashing scheme.
func HashCrossDomainMessageV1(
	nonce *big.Int,
	sender common.Address,
	target common.Address,
	mntValue *big.Int,
	ethValue *big.Int,
	gasLimit *big.Int,
	data []byte,
) (common.Hash, error) {
	encoded, err := EncodeCrossDomainMessageV1(nonce, sender, target, mntValue, ethValue, gasLimit, data)
	if err != nil {
		return common.Hash{}, err
	}
	hash := crypto.Keccak256(encoded)
	return common.BytesToHash(hash), nil
}
