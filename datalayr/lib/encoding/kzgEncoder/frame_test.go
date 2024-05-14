package kzgEncoder_test

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
	kzg "github.com/Layr-Labs/datalayr/lib/encoding/kzg"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
)

func TestEncodeDecodeFrame_AreInverses(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgRs.NewKzgEncoderGroup(kzgConfig)
	enc, err := group.NewKzgEncoder(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)))
	require.Nil(t, err)
	require.NotNil(t, enc)

	_, _, frames, _, err := enc.EncodeBytes(context.Background(), GETTYSBURG_ADDRESS_BYTES)
	require.Nil(t, err)
	require.NotNil(t, frames, err)

	b, err := frames[0].Encode()
	require.Nil(t, err)
	require.NotNil(t, b)

	frame, err := kzgRs.Decode(b)
	require.Nil(t, err)
	require.NotNil(t, frame)

	assert.Equal(t, frame, frames[0])
}

func TestVerify(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	group, _ := kzgRs.NewKzgEncoderGroup(kzgConfig)
	enc, err := group.NewKzgEncoder(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)))
	require.Nil(t, err)
	require.NotNil(t, enc)

	commit, _, frames, _, err := enc.EncodeBytes(context.Background(), GETTYSBURG_ADDRESS_BYTES)
	require.Nil(t, err)
	require.NotNil(t, commit)
	require.NotNil(t, frames)

	params := rs.GetEncodingParams(numSys, numPar, uint64(len(GETTYSBURG_ADDRESS_BYTES)))
	require.NotNil(t, params)

	n := uint8(math.Log2(float64(params.PaddedNodeGroupSize)))
	fs := kzg.NewFFTSettings(n)
	require.NotNil(t, fs)

	lc := enc.Fs.ExpandedRootsOfUnity[uint64(0)]
	require.NotNil(t, lc)

	assert.True(t, frames[0].Verify(enc.Ks, commit, &lc))
}
