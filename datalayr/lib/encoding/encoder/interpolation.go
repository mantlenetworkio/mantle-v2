package encoder

import (
	bls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
	rb "github.com/Layr-Labs/datalayr/lib/encoding/utils/reverseBits"
)

// Consider input data as the polynomial Coefficients, c
// This functions computes the evaluations of the such the interpolation polynomial
// Passing through input data, evaluated at series of root of unity.
// Consider the following points (w, d[0]), (wφ, d[1]), (wφ^2, d[2]), (wφ^3, d[3])
// Suppose F be the fft matrix, then the systamtic equation that going through those points is
// d = W F c, where each row corresponds to equation being evaluated at [1, φ, φ^2, φ^3]
// where W is a diagonal matrix with diagonal [1 w w^2 w^3] for shifting the evaluation points

// The index is transformed using FFT, for example 001 => 100, 110 => 011
// The reason behind is because Reed Solomon extension using FFT insert evaluation within original
// Data. i.e. [o_1, o_2, o_3..] with coding ratio 0.5 becomes [o_1, p_1, o_2, p_2...]

func (g *Encoder) GetInterpolationPolyEval(
	interpolationPoly []bls.Fr,
	j uint32,
) ([]bls.Fr, error) {
	evals := make([]bls.Fr, g.ChunkLen)
	w := g.Fs.ExpandedRootsOfUnity[uint64(j)]
	shiftedInterpolationPoly := make([]bls.Fr, len(interpolationPoly))

	//multiply each term of the polynomial by x^i so the fourier transform results in the desired evaluations
	//The fourier matrix looks like
	// ___                    ___
	// | 1  1   1    1  . . . . |
	// | 1  φ   φ^2 φ^3         |
	// | 1  φ^2 φ^4 φ^6         |
	// | 1  φ^3 φ^6 φ^9         |  = F
	// | .   .          .       |
	// | .   .            .     |
	// | .   .              .   |
	// |__                    __|

	//
	// F * p = [p(1), p(φ), p(φ^2), ...]
	//
	// but we want
	//
	// [p(w), p(wφ), p(wφ^2), ...]
	//
	// we can do this by computing shiftedInterpolationPoly = q = p(wx) and then doing
	//
	// F * q = [p(w), p(wφ), p(wφ^2), ...]
	//
	// to get our desired evaluations
	// cool idea protolambda :)
	var wPow bls.Fr
	bls.CopyFr(&wPow, &bls.ONE)
	var tmp, tmp2 bls.Fr
	for i := 0; i < len(interpolationPoly); i++ {
		bls.MulModFr(&tmp2, &interpolationPoly[i], &wPow)
		bls.CopyFr(&shiftedInterpolationPoly[i], &tmp2)
		bls.MulModFr(&tmp, &wPow, &w)
		bls.CopyFr(&wPow, &tmp)
	}

	err := g.Fs.InplaceFFT(shiftedInterpolationPoly, evals, false)
	return evals, err
}

// Since both F W are invertible, c = W^-1 F^-1 d, convert it back. F W W^-1 F^-1 d = c
func (g *Encoder) GetInterpolationPolyCoeff(chunk []bls.Fr, k uint32) ([]bls.Fr, error) {
	coeffs := make([]bls.Fr, g.ChunkLen)
	w := g.Fs.ExpandedRootsOfUnity[uint64(k)]
	shiftedInterpolationPoly := make([]bls.Fr, len(chunk))
	err := g.Fs.InplaceFFT(chunk, shiftedInterpolationPoly, true)
	if err != nil {
		return coeffs, err
	}
	var wPow bls.Fr
	bls.CopyFr(&wPow, &bls.ONE)
	var tmp, tmp2 bls.Fr
	for i := 0; i < len(chunk); i++ {
		bls.InvModFr(&tmp, &wPow)
		bls.MulModFr(&tmp2, &shiftedInterpolationPoly[i], &tmp)
		bls.CopyFr(&coeffs[i], &tmp2)
		bls.MulModFr(&tmp, &wPow, &w)
		bls.CopyFr(&wPow, &tmp)
	}
	return coeffs, nil
}

// Create evaluation using coeff based on systematic data chunk
func (g *Encoder) CreateSysEvalChunks(interpolatingCoeff []bls.Fr) ([][]bls.Fr, error) {
	chunks := make([][]bls.Fr, g.NumSysE)
	for i := uint64(0); i < g.NumSysE; i++ {
		chunks[i] = make([]bls.Fr, g.ChunkLen)
		// evaluate each ChunkLen of the data as a interpolating poly to get the evals
		// where to evaluate at depends on index
		j := rb.ReverseBitsLimited(uint32(g.NumNodeE), uint32(i))
		evals, err := g.GetInterpolationPolyEval(
			interpolatingCoeff[g.ChunkLen*i:g.ChunkLen*(i+1)],
			j,
		)
		chunks[i] = evals

		if err != nil {
			return nil, err
		}
	}
	return chunks, nil
}

// suppose the original data are [1,2,3,o,4,5..,n], represented as a matrix
// [1, 2, 3, o]
// [4, 5, 6, p]
// [7, 8, 9, m]
// [a, b, c, n]
// where each row is the coeff for the smaller poly.
// last GetInterpolationPolyEval step applied F W to every row of the matrix
// to transfer coeffs to evaluations,  such that we want every entry in the matrix
// corresponds to distinct evaluation index for the full polynomial,
// whose indices are [1, w, w^2, w^3, φ, wφ, w^2φ, ...]
// But since our data undergo a butterfly operation
// [1, 7, 4, a, 2, 8, 5, b, 3, 9, 6, c, o, m, p, n]
// those number are symbolic, but they should gives a assending order
// [1, w, w^2, w^3, φ, wφ, w^2φ, ...]
// hence chunk 0 has a coset of [1, 2, 3, o] <=> [1,      φ,    φ^2,    φ^3] <=> F W0 [1, 2, 3, o]
// hence chunk 2 has a coset of [7, 8, 9, m] <=> [w,     wφ,   wφ^2,   wφ^3] <=> F W2 [7, 8, 9, m]
// hence chunk 1 has a coset of [4, 5, 6, p] <=> [w^2, w^2φ, w^2φ^2, w^2φ^3] <=> F W1 [4, 5, 6, p]
// hence chunk 3 has a coset of [a, b, c, n] <=> [w^3, w^3φ, w^3φ^2, w^3φ^3] <=> F W3 [a, b, c, n]
// W_i are different diaganol matrices
func (g *Encoder) SpliceEvalChunks(chunks [][]bls.Fr) []bls.Fr {
	dataFr := make([]bls.Fr, g.PaddedSysGroupSize)
	indices := make([]int, 0)
	for i := 0; i < len(chunks); i++ {
		indices = append(indices, int(rb.ReverseBitsLimited(uint32(g.NumSysE), uint32(i))))
	}

	for i := 0; i < int(g.ChunkLen); i++ {
		for j, k := range indices {
			dataFr[i*len(chunks)+k] = chunks[j][i]
		}
	}
	return dataFr
}
