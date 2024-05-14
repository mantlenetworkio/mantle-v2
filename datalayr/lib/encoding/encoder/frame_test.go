package encoder_test

import (
	"testing"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeFrame_AreInverses(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	enc, _ := rs.NewEncoder(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)), true)
	require.NotNil(t, enc)

	_, frames, _, err := enc.EncodeBytes(GETTYSBURG_ADDRESS_BYTES)
	require.Nil(t, err)
	require.NotNil(t, frames, err)

	b, err := frames[0].Encode()
	require.Nil(t, err)
	require.NotNil(t, b)

	frame, err := rs.Decode(b)
	require.Nil(t, err)
	require.NotNil(t, frame)

	assert.Equal(t, frame, frames[0])
}
