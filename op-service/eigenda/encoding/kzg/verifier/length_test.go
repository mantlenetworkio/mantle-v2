package verifier_test

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/eigenda/encoding"
	"github.com/ethereum-optimism/optimism/op-service/eigenda/encoding/kzg/prover"
	"github.com/ethereum-optimism/optimism/op-service/eigenda/encoding/kzg/verifier"
	"github.com/ethereum-optimism/optimism/op-service/eigenda/encoding/rs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLengthProof(t *testing.T) {

	group, _ := prover.NewProver(kzgConfig, true)
	v, _ := verifier.NewVerifier(kzgConfig, true)
	params := encoding.ParamsFromSysPar(numSys, numPar, uint64(len(gettysburgAddressBytes)))
	enc, err := group.GetKzgEncoder(params)
	require.Nil(t, err)

	numBlob := 5
	for z := 0; z < numBlob; z++ {
		extra := make([]byte, z*32*2)
		inputBytes := append(gettysburgAddressBytes, extra...)
		inputFr, err := rs.ToFrArray(inputBytes)
		require.Nil(t, err)

		_, lengthCommitment, lengthProof, _, _, err := enc.Encode(inputFr)
		require.Nil(t, err)

		length := len(inputFr)
		assert.NoError(t, v.VerifyCommit(lengthCommitment, lengthProof, uint64(length)), "low degree verification failed\n")

		length = len(inputFr) - 10
		assert.Error(t, v.VerifyCommit(lengthCommitment, lengthProof, uint64(length)), "low degree verification failed\n")
	}
}
