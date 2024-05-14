package encoder

import (
	"errors"
	"math"

	kzg "github.com/Layr-Labs/datalayr/lib/encoding/kzg"
)

type Encoder struct {
	EncodingParams

	Fs *kzg.FFTSettings

	verbose bool
}

// The function creates a high level struct that determines the encoding the a data of a
// specific length under (num systematic node, num parity node) setup. A systematic node
// stores a systematic data chunk that contains part of the original data. A parity node
// stores a parity data chunk which is an encoding of the original data. A receiver that
// collects all systematic chunks can simply stitch data together to reconstruct the
// original data. When some systematic chunks are missing but identical parity chunk are
// available, the receive can go through a Reed Solomon decoding to reconstruct the
// original data.
func NewEncoder(numSys, numPar, dataByteLen uint64, verbose bool) (*Encoder, error) {
	if !CheckPreconditions(numSys, numPar) {
		return nil, errors.New("kzgFFT input precondition not satisfied")
	}

	params := GetEncodingParams(numSys, numPar, dataByteLen)

	n := uint8(math.Log2(float64(params.PaddedNodeGroupSize)))
	fs := kzg.NewFFTSettings(n)

	return &Encoder{
		EncodingParams: params,
		Fs:             fs,
		verbose:        verbose,
	}, nil

}

// This function checks preconditions for kzgFFT library
// 1. numPar must be non-zero
// 2. numSys must be non-zero
func CheckPreconditions(numSys, numPar uint64) bool {
	if numSys == 0 {
		return false
	}

	if numPar == 0 {
		return false
	}
	return true
}
