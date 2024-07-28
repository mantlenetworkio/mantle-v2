package prover_test

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/eigenda/encoding"
	"github.com/ethereum-optimism/optimism/op-service/eigenda/encoding/kzg/prover"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeFrame_AreInverses(t *testing.T) {

	group, _ := prover.NewProver(kzgConfig, true)

	params := encoding.ParamsFromSysPar(numSys, numPar, uint64(len(gettysburgAddressBytes)))

	p, err := group.GetKzgEncoder(params)

	require.Nil(t, err)
	require.NotNil(t, p)

	_, _, _, frames, _, err := p.EncodeBytes(gettysburgAddressBytes)
	require.Nil(t, err)
	require.NotNil(t, frames, err)

	b, err := frames[0].Encode()
	require.Nil(t, err)
	require.NotNil(t, b)

	frame, err := encoding.Decode(b)
	require.Nil(t, err)
	require.NotNil(t, frame)

	assert.Equal(t, frame, frames[0])
}
