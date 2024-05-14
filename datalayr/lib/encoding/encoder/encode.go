package encoder

import (
	"errors"
	"log"
	"time"

	bls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
	rb "github.com/Layr-Labs/datalayr/lib/encoding/utils/reverseBits"
)

type GlobalPoly struct {
	Coeffs []bls.Fr
	Values []bls.Fr
}

// just a wrapper to take bytes not Fr Element
func (g *Encoder) EncodeBytes(inputBytes []byte) (*GlobalPoly, []Frame, []uint32, error) {
	if g.verbose {
		log.Println("Entering EncodeBytes function")
		defer log.Println("Exiting EncodeBytes function")
	}

	inputFr := ToFrArray(inputBytes)
	return g.Encode(inputFr)
}

// Encode function takes input in unit of Fr Element, creates a kzg commit and a list of frames
// which contains a list of multireveal interpolating polynomial coefficients, a G1 proof and a
// low degree proof corresponding to the interpolating polynomial. Each frame is an independent
// group of data verifiable to the kzg commitment. The encoding functions ensures that in each
// frame, the multireveal interpolating coefficients are identical to the part of input bytes
// in the form of field element. The extra returned integer list corresponds to which leading
// coset root of unity, the frame is proving against, which can be deduced from a frame's index
func (g *Encoder) Encode(inputFr []bls.Fr) (*GlobalPoly, []Frame, []uint32, error) {
	start := time.Now()
	intermediate := time.Now()
	if g.verbose {
		log.Println("Entering Encode function")
		defer log.Println("Exiting Encode function")
	}

	// treating input as coeff of the interpolating poly
	interpolatingCoeffs := g.PadToRequiredSymbols(inputFr)

	if g.verbose {
		log.Printf("    Pad takes %v\n", time.Since(intermediate))
		intermediate = time.Now()
	}

	// divide the data into chunks, treat a chunk data as coeff of poly and get its eval
	evalChunks, err := g.CreateSysEvalChunks(interpolatingCoeffs)
	if err != nil {
		return nil, nil, nil, err
	}

	// butterfly reorder the chunks to make it digestible for erasure code (polynomial extension)
	splicedPolyEvals := g.SpliceEvalChunks(evalChunks)

	// compute polynomial coefficient. Note this interpolates the full poly, which is not
	// the original data. We want the interpolating poly in the frame coset to be the data
	fullCoeffsPoly, err := g.ConvertEvalsToCoeffs(splicedPolyEvals)
	if err != nil {
		return nil, nil, nil, err
	}

	if g.verbose {
		log.Printf("    Chunk and Splice take %v\n", time.Since(intermediate))
		log.Printf("    eval %v\n", time.Since(intermediate))
		log.Printf("    coeff %v\n", time.Since(intermediate))
		intermediate = time.Now()
	}

	// extend data based on Sys, Par ratio. The returned fullCoeffsPoly is padded with 0 to ease proof
	fullPolyEvals, fullCoeffsPoly, err := g.ExtendPolyEval(fullCoeffsPoly)
	if err != nil {
		return nil, nil, nil, err
	}

	poly := &GlobalPoly{
		Values: fullPolyEvals,
		Coeffs: fullCoeffsPoly,
	}

	if g.verbose {
		log.Printf("    Extending evaluation takes  %v\n", time.Since(intermediate))
	}

	// create frames to group relevant info
	frames, indices, err := g.MakeFrames(fullPolyEvals, interpolatingCoeffs)
	if err != nil {
		return nil, nil, nil, err
	}

	if g.verbose {
		log.Printf("  SUMMARY: Encode %v byte among %v numNode out of %v extended numNode takes %v\n",
			len(inputFr)*bls.BYTES_PER_COEFFICIENT, g.NumSys+g.NumPar, g.NumSysE, time.Since(start))
	}

	return poly, frames, indices, nil
}

// This Function takes extended evaluation data and bundles relevant information into Frame.
// Every frame is verifiable to the commitment. It returns only first numSys frame, and the
// first numSysE to numSysE + numPar chunks
func (g *Encoder) MakeFrames(
	encodedEvalFr []bls.Fr,
	coeffs []bls.Fr,
) ([]Frame, []uint32, error) {
	if g.verbose {
		log.Println("Entering MakeFrames function")
		defer log.Println("Exiting MakeFrames function")
	}

	numFrame := g.NumSys + g.NumPar
	numPadSys := g.NumSysE - g.NumSys
	if numFrame > g.NumNodeE-numPadSys {
		return nil, nil, errors.New("cannot create number of frame higher than possible")
	}

	// reverse dataFr making easier to sample points
	err := rb.ReverseBitOrderFr(encodedEvalFr)
	if err != nil {
		return nil, nil, err
	}
	k := uint64(0)

	indices := make([]uint32, 0)
	frames := make([]Frame, numFrame)

	for i := uint64(0); i < uint64(g.NumNodeE); i++ {

		// skip padded chunk and its coding
		if i >= g.NumSys && i < g.NumSysE {
			continue
		}
		// if collect sufficient chunks
		if k == numFrame {
			return frames, indices, nil
		}

		// finds out which coset leader i-th node is having
		j := rb.ReverseBitsLimited(uint32(g.NumNodeE), uint32(i))

		// mutltiprover return proof in butterfly order
		frame := Frame{}
		indices = append(indices, j)

		// since coeff of interpolating poly for the coset is data itself, avoid taking FFT
		if i < g.NumSys {
			frame.Coeffs = coeffs[g.ChunkLen*i : g.ChunkLen*(i+1)]
		} else {
			ys := encodedEvalFr[g.ChunkLen*i : g.ChunkLen*(i+1)]
			err := rb.ReverseBitOrderFr(ys)
			if err != nil {
				return nil, nil, err
			}
			coeffs, err := g.GetInterpolationPolyCoeff(ys, uint32(j))
			if err != nil {
				return nil, nil, err
			}

			frame.Coeffs = coeffs

		}

		frames[k] = frame
		k++
	}

	return frames, indices, nil
}

// Encoding Reed Solomon using FFT
func (g *Encoder) ExtendPolyEval(coeffs []bls.Fr) ([]bls.Fr, []bls.Fr, error) {
	if g.verbose {
		log.Println("Entering ExtendPolyEval function")
		defer log.Println("Exiting ExtendPolyEval function")
	}

	pdCoeffs := make([]bls.Fr, g.PaddedNodeGroupSize)
	for i := 0; i < len(coeffs); i++ {
		bls.CopyFr(&pdCoeffs[i], &coeffs[i])
	}
	for i := len(coeffs); i < len(pdCoeffs); i++ {
		bls.CopyFr(&pdCoeffs[i], &bls.ZERO)
	}

	evals, err := g.Fs.FFT(pdCoeffs, false)
	if err != nil {
		return nil, nil, err
	}

	return evals, pdCoeffs, nil
}

// get coeff from data
func (g *Encoder) ConvertEvalsToCoeffs(coeffs []bls.Fr) ([]bls.Fr, error) {
	if g.verbose {
		log.Println("Entering ConvertEvalsToCoeffs function")
		defer log.Println("Exiting ConvertEvalsToCoeffs function")
	}

	evals, err := g.Fs.FFT(coeffs, true)
	if err != nil {
		return nil, err
	}
	return evals, nil
}

// Pad 0 to input to reach the closest power of 2
func (g *Encoder) PadToRequiredSymbols(dataFr []bls.Fr) []bls.Fr {
	if g.verbose {
		log.Println("Entering PadToRequiredSymbols function")
		defer log.Println("Exiting PadToRequiredSymbols function")
	}

	outFr := make([]bls.Fr, g.PaddedSysGroupSize)
	copy(outFr, dataFr)
	for i := len(dataFr); i < len(outFr); i++ {
		bls.CopyFr(&outFr[i], &bls.ZERO)
	}
	return outFr
}
