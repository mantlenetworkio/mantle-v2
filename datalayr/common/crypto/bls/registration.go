package bls

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (k *BlsKeyPair) MakeRegistrationData(operatorAddress common.Address) (*big.Int, G1Point, G2Point) {
	secretKeyBytes := k.secretKey.Bytes()
	nonceDiscreteLog := make([]byte, 0)

	for i := 0; i < len(secretKeyBytes); i++ {
		nonceDiscreteLog = append(nonceDiscreteLog, secretKeyBytes[i])
	}

	r := new(fr.Element).SetBytes(crypto.Keccak256(nonceDiscreteLog, []byte("registration with bls pubkey registry")))

	c := generateSchnorrChallenge(operatorAddress, k.PublicKey, mulByGeneratorG1(r))

	//s = cx
	cx := new(fr.Element).Mul(c, k.secretKey)
	//s = r + cx
	s := new(fr.Element).Add(r, cx)

	return s.ToBigIntRegular(new(big.Int)), ToG1Point(mulByGeneratorG1(r)), ToG2Point(mulByGeneratorG2(k.secretKey))
}

func generateSchnorrChallenge(operatorAddress common.Address, pubkey, R *bn254.G1Affine) *fr.Element {
	toHash := operatorAddress.Bytes()
	tmp := pubkey.X.Bytes()
	for i := 0; i < 32; i++ {
		toHash = append(toHash, tmp[i])
	}
	tmp = pubkey.Y.Bytes()
	for i := 0; i < 32; i++ {
		toHash = append(toHash, tmp[i])
	}
	tmp = R.X.Bytes()
	for i := 0; i < 32; i++ {
		toHash = append(toHash, tmp[i])
	}
	tmp = R.Y.Bytes()
	for i := 0; i < 32; i++ {
		toHash = append(toHash, tmp[i])
	}
	return new(fr.Element).SetBytes(crypto.Keccak256(toHash))
}
