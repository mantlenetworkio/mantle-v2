package encoder

import (
	bls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
)

type EncodingParams struct {
	// num chunk meta
	NumSys              uint64 // number of systematic nodes that are storing data
	NumPar              uint64 // number of parity nodes that are storing data
	NumSysE             uint64 // number of systematic chunks that are padded to power of 2
	NumNodeE            uint64 // number of total chunks that are padded to power of 2
	ChunkLen            uint64 // number of Fr symbol stored inside a chunk
	ChunkDegree         uint64 // degree of the polynomial interpolating a chunk. ChunkDegree = ChunkLen-1
	PaddedSysGroupSize  uint64 // the size of the group (power of 2) used to construct the global polynomial which interpolates numSysE padded systematic chunks. PaddedSysGroupSize = NumSysE*ChunkLen
	PaddedNodeGroupSize uint64 // the size of the group (power of 2) on which the global polynomial is evaluted in order to extend to NumPar chunks PaddedNodeGroupSize = NumNodeE*ChunkLen
	GlobalPolyDegree    uint64 // degree of the PaddedSymPoly. GlobalPolyDegree = PaddedSysGroupSize - 1
}

// Used to save only the Encoding Parameters which are key degrees of freedom which which other params can be reconstructed
type EncodingKey struct {
	NumSys   uint64
	NumPar   uint64
	ChunkLen uint64
}

func GetEncodingParams(numSys, numPar, dataByteLen uint64) EncodingParams {

	// Extended number of systematic symbols
	numSysE := NextPowerOf2(numSys)

	// Extended number of nodes
	numNode := numSys + numPar
	ratio := RoundUpDivision(numNode, numSys)
	numNodeE := NextPowerOf2(numSysE * ratio)

	// chunk/coset size to fit into FFT for multi-reveal
	dataLen := RoundUpDivision(dataByteLen, bls.BYTES_PER_COEFFICIENT)
	chunkLen := NextPowerOf2(RoundUpDivision(dataLen, numSys))

	// Order of the global polynomial + 1
	paddedSysGroupSize := numSysE * chunkLen   // This will always be a power of 2, since both factors are powers of 2.
	paddedNodeGroupSize := numNodeE * chunkLen // This will always be a power of 2, since both factors are powers of 2.

	params := EncodingParams{
		NumSys:              numSys,
		NumPar:              numPar,
		NumSysE:             numSysE,
		NumNodeE:            numNodeE,
		ChunkLen:            chunkLen,
		ChunkDegree:         chunkLen - 1,
		PaddedSysGroupSize:  paddedSysGroupSize,
		PaddedNodeGroupSize: paddedNodeGroupSize,
		GlobalPolyDegree:    paddedSysGroupSize - 1,
	}

	return params
}
