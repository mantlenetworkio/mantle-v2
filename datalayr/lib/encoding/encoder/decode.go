package encoder

import (
	"errors"
	"log"

	bls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
	rb "github.com/Layr-Labs/datalayr/lib/encoding/utils/reverseBits"
)

// Decode data when some chunks from systematic nodes are lost. It first uses FFT to recover
// the whole polynomial. Then it extracts only the systematic chunks
func (g *Encoder) Decode(samples []Frame, indices []uint64, inputSize uint64) ([]bls.Fr, error) {
	if g.verbose {
		log.Println("Entering Decode function")
		defer log.Println("Exiting Decode function")
	}

	reconSysEval, err := g.RecoverPolyEval(samples, indices)
	if err != nil {
		return nil, err
	}

	concatFr := make([]bls.Fr, 0)
	for j := 0; j < int(g.NumSysE); j++ {
		dataFr := make([]bls.Fr, g.ChunkLen)
		for i := 0; i < int(g.ChunkLen); i++ {
			dataFr[i] = reconSysEval[i*int(g.NumSysE)+j]
		}

		z := rb.ReverseBitsLimited(uint32(g.NumNodeE), uint32(j))
		coeffs, err := g.GetInterpolationPolyCoeff(dataFr, z)
		if err != nil {
			return nil, err
		}
		concatFr = append(concatFr, coeffs...)
	}

	return concatFr[:inputSize], nil
}

// Decode Original data if chunks from all systematic node are received. No need to do erasure recovery
func (g *Encoder) DecodeSys(frames []Frame, indices []uint64, inputSize uint64) ([]bls.Fr, error) {
	if g.verbose {
		log.Println("Entering DecodeSys function")
		defer log.Println("Exiting DecodeSys function")
	}

	codedDataFr := make([]bls.Fr, g.PaddedSysGroupSize)

	num := 0
	numFrame := uint64(0)
	for i, d := range indices {
		if uint64(d) < g.NumSys {
			f := frames[i]
			for j := uint64(0); j < g.ChunkLen; j++ {
				p := j*g.NumSysE + d
				bls.CopyFr(&codedDataFr[p], &f.Coeffs[j])
				num += 1
			}
			numFrame += 1
		}
	}
	if numFrame != g.NumSys {
		return nil, errors.New("does not contain sufficient chunks from systematic nodes")
	}

	concatFr := make([]bls.Fr, 0)
	for j := 0; j < int(g.NumSysE); j++ {
		dataFr := make([]bls.Fr, g.ChunkLen)
		for i := 0; i < int(g.ChunkLen); i++ {
			dataFr[i] = codedDataFr[i*int(g.NumSysE)+j]
		}
		concatFr = append(concatFr, dataFr...)
	}

	return concatFr[:inputSize], nil
}

// This function takes a list of available frame, and return the original encoded data
// storing the evaluation points, since it is where RS is applied. The input frame contains
// the coefficient of the interpolating polynomina, hence interpolation is needed before
// recovery. The code autmomatic padded the systematic chunk (numSys <= chunk id < numSysE)
// which consists of zeros bls.Fr
func (g *Encoder) RecoverPolyEval(frames []Frame, indices []uint64) ([]bls.Fr, error) {
	if g.verbose {
		log.Println("Entering RecoverPolyEval function")
		defer log.Println("Exiting RecoverPolyEval function")
	}

	if uint64(len(frames)) < g.NumSys {
		return nil, errors.New("number of frame must be sufficient")
	}

	samples := make([]*bls.Fr, g.NumNodeE*g.ChunkLen)
	// copy evals based on frame coeffs into samples
	for i, d := range indices {
		f := frames[i]
		e, err := GetLeadingCosetIndex(d, g.NumSys, g.NumPar)
		if err != nil {
			return nil, err
		}

		evals, err := g.GetInterpolationPolyEval(f.Coeffs, uint32(e))
		if err != nil {
			return nil, err
		}

		// Some pattern i butterfly swap. Find the leading coset, then increment by number of coset
		for j := uint64(0); j < g.ChunkLen; j++ {
			p := j*g.NumNodeE + uint64(e)
			samples[p] = new(bls.Fr)
			bls.CopyFr(samples[p], &evals[j])
		}
	}

	// padded zero chunks
	zeroChunk := make([]bls.Fr, g.ChunkLen)
	for i := uint64(0); i < g.ChunkLen; i++ {
		bls.CopyFr(&zeroChunk[i], &bls.ZERO)
	}
	// copy evals based on frame zero coeffs padded NumSys dimension
	for d := g.NumSys; d < g.NumSysE; d++ {
		e := rb.ReverseBitsLimited(uint32(g.NumNodeE), uint32(d))
		evals, err := g.GetInterpolationPolyEval(zeroChunk, uint32(e))
		if err != nil {
			return nil, err
		}

		for j := uint64(0); j < g.ChunkLen; j++ {
			p := j*g.NumNodeE + uint64(e)
			samples[p] = new(bls.Fr)
			bls.CopyFr(samples[p], &evals[j])
		}
	}

	recovered, err := g.Fs.RecoverPolyFromSamples(
		samples,
		g.Fs.ZeroPolyViaMultiplication,
	)
	if err != nil {
		return nil, err
	}

	// extract only systematic data evals from full recovered eval poly
	polyOrder := g.NumSysE * g.ChunkLen
	orderRatio := uint64(len(recovered)) / (polyOrder)
	dataFr := make([]bls.Fr, polyOrder)

	k := uint64(0)
	for j := uint64(0); j < polyOrder; j++ {
		bls.CopyFr(&dataFr[k], &recovered[j*orderRatio])
		k++
	}

	for j := uint64(0); j < g.ChunkLen; j++ {
		err = rb.ReverseBitOrderFr(dataFr[j*g.NumSysE : (j+1)*g.NumSysE])
		if err != nil {
			return nil, err
		}
	}

	return dataFr, nil
}

func (g *Encoder) DecodeSafe(frames []Frame, indices []uint64, inputSize uint64) ([]byte, error) {
	if g.verbose {
		log.Println("Entering DecodeSafe function")
		defer log.Println("Exiting DecodeSafe function")
	}

	numFr := GetNumElement(inputSize, bls.BYTES_PER_COEFFICIENT)
	mpFrames := make([]Frame, len(frames))

	copy(mpFrames, frames)

	var data []byte

	// TODO optimize that if num sys is sufficient
	if g.EncodingParams.NumSys+g.EncodingParams.NumPar > uint64(len(frames)) {
		dataFr, err := g.Decode(mpFrames[:], indices[:], numFr)
		if err != nil {
			return nil, err
		}

		data = ToByteArray(dataFr, inputSize)
	} else {

		dataFr, err := g.DecodeSys(mpFrames, indices, numFr)
		if err != nil {
			return nil, err
		}
		data = ToByteArray(dataFr, inputSize)
	}
	return data, nil
}
