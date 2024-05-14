//go:build !bignum_pure && !bignum_hol256
// +build !bignum_pure,!bignum_hol256

package kzg

import (
	"testing"

	bls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKZGSettings_DAUsingFK20(t *testing.T) {
	fs := NewFFTSettings(5)
	s1, s2 := GenerateTestingSetup("1927409816240961209460912649124", 32+1)
	srs, _ := NewSrs(s1, s2)
	ks, _ := NewKZGSettings(fs, srs)
	fk := NewFK20SingleSettings(ks, 32)

	polynomial := testPoly(1, 2, 3, 4, 7, 7, 7, 7, 13, 13, 13, 13, 13, 13, 13, 13)

	commitment := ks.CommitToPoly(polynomial)
	t.Log("commitment\n", bls.StrG1(commitment))

	allProofs, err := fk.DAUsingFK20(polynomial)
	require.Nil(t, err, "could not compute proof")
	require.NotNil(t, allProofs)

	t.Log("All KZG proofs computed")
	for i := 0; i < len(allProofs); i++ {
		t.Logf("%d: %s", i, bls.StrG1(&allProofs[i]))
	}

	// Now check a random position
	pos := uint64(9)
	var posFr bls.Fr
	bls.AsFr(&posFr, pos)
	var x bls.Fr
	bls.CopyFr(&x, &ks.ExpandedRootsOfUnity[pos])
	t.Log("x:\n", bls.FrStr(&x))
	var y bls.Fr
	bls.EvalPolyAt(&y, polynomial, &x)
	t.Log("y:\n", bls.FrStr(&y))

	proof := &allProofs[reverseBitsLimited(uint32(2*16), uint32(pos))]
	assert.True(t, ks.CheckProofSingle(commitment, proof, &x, &y), "could not verify proof")
}
