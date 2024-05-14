package bls_test

import (
	"math/big"
	"math/rand"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bls", func() {
	_, _, g1Aff, g2Aff := bn254.Generators()
	var randomG1 bn254.G1Affine
	var randomG2 bn254.G2Affine

	BeforeEach(func() {
		//randomG1 and randomG2 are random scalar multiples of the generator, with the multiple known
		//not the same at hashing to curve
		randInt := new(big.Int)
		randInt.Rand(rand.New(rand.NewSource(1)), fp.Modulus())
		randomG1.ScalarMultiplication(&g1Aff, randInt)
		randInt.Rand(rand.New(rand.NewSource(1)), fp.Modulus())
		randomG2.ScalarMultiplication(&g2Aff, randInt)
	})

	Describe("Serializing points", func() {
		Context("in G1", func() {
			It("should work", func() {
				randomG1Bytes := bls.SerializeG1(&randomG1)
				resRandomG1 := bls.DeserializeG1(randomG1Bytes)
				Expect(randomG1.Equal(resRandomG1)).To(Equal(true))
			})
		})

		Context("in G2", func() {
			It("should work", func() {
				randomG2Bytes := bls.SerializeG2(&randomG2)
				resRandomG2 := bls.DeserializeG2(randomG2Bytes)
				Expect(randomG2.Equal(resRandomG2)).To(Equal(true))
			})
		})
	})
})
