package encoder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
)

func TestEncodeDecode_InvertsWhenSamplingAllFrames(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	enc, _ := rs.NewEncoder(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)), true)
	require.NotNil(t, enc)

	inputFr := rs.ToFrArray(GETTYSBURG_ADDRESS_BYTES)
	_, frames, _, err := enc.Encode(inputFr)

	// sample some frames
	samples, indices := sampleFrames(frames, uint64(len(frames)))
	data, err := enc.DecodeSafe(samples, indices, uint64(len(GETTYSBURG_ADDRESS_BYTES)))

	require.Nil(t, err)
	require.NotNil(t, data)

	assert.Equal(t, data, GETTYSBURG_ADDRESS_BYTES)
}

func TestEncodeDecode_InvertsWhenSamplingMissingFrame(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	enc, _ := rs.NewEncoder(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)), true)
	require.NotNil(t, enc)

	inputFr := rs.ToFrArray(GETTYSBURG_ADDRESS_BYTES)
	_, frames, _, err := enc.Encode(inputFr)

	// sample some frames
	samples, indices := sampleFrames(frames, uint64(len(frames)-1))
	data, err := enc.DecodeSafe(samples, indices, uint64(len(GETTYSBURG_ADDRESS_BYTES)))

	require.Nil(t, err)
	require.NotNil(t, data)

	assert.Equal(t, data, GETTYSBURG_ADDRESS_BYTES)
}

func TestEncodeDecode_ErrorsWhenNotEnoughSampledFrames(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	enc, _ := rs.NewEncoder(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)), true)
	require.NotNil(t, enc)

	inputFr := rs.ToFrArray(GETTYSBURG_ADDRESS_BYTES)
	_, frames, _, err := enc.Encode(inputFr)

	// sample some frames
	samples, indices := sampleFrames(frames, uint64(len(frames)-2))
	data, err := enc.DecodeSafe(samples, indices, uint64(len(GETTYSBURG_ADDRESS_BYTES)))

	require.Nil(t, data)
	require.NotNil(t, err)

	assert.EqualError(t, err, "number of frame must be sufficient")
}
