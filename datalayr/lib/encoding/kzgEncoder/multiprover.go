package kzgEncoder

import (
	"math"

	bls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
	"github.com/Layr-Labs/datalayr/lib/encoding/utils/toeplitz"
)

type WorkerResult struct {
	points []bls.G1Point
	err    error
}

func (p *KzgEncoder) ProveAllCosetThreads(polyFr []bls.Fr, numChunks, chunkLen, numWorker uint64) ([]bls.G1Point, error) {

	// Robert: Standardizing this to use the same math used in precomputeSRS
	dimE := numChunks
	l := chunkLen

	// nc = CeilIntPowerOf2Num(nc)
	// coset size = order of root of unity / number cosets
	// l := p.Fs.MaxWidth / nc
	//dim := (len(polyFr) - 1) / l
	// dimE := CeilIntPowerOf2Num(uint64(len(polyFr)) / l)

	//fmt.Println("l", l)
	//fmt.Println("nc", nc)
	//fmt.Println("dim", dim)
	//fmt.Println("dimE", dimE)
	//fmt.Println("polyFr len", len(polyFr))

	sumVec := make([]bls.G1Point, dimE*2)
	// Robert: This code doesn't appear to be needed
	// for i := uint64(0); i < numChunks; i++ {
	// 	bls.CopyG1(&sumVec[i], &bls.ZeroG1)
	// }

	numJob := dimE * 2

	jobChan := make(chan uint64, numJob)
	results := make(chan WorkerResult, l)

	for w := uint64(0); w < numWorker; w++ {
		go p.proofWorker(polyFr, jobChan, l, results)
	}

	for j := uint64(0); j < l; j++ {
		jobChan <- j
	}
	close(jobChan)

	// return only first error
	var err error
	for w := uint64(0); w < l; w++ {
		wr := <-results
		if wr.err == nil {
			for i := 0; i < len(wr.points); i++ {
				bls.AddG1(&sumVec[i], &sumVec[i], &wr.points[i])
			}
		} else {
			err = wr.err
		}
	}

	if err != nil {
		return nil, err
	}

	// only 1 ifft is needed
	sumVecInv, err := p.Fs.FFTG1(sumVec, true)
	if err != nil {
		return nil, err
	}

	// outputs is out of order - buttefly
	proofs, err := p.Fs.FFTG1(sumVecInv[:dimE], false)
	if err != nil {
		return nil, err
	}

	//rb.ReverseBitOrderG1Point(proofs)
	return proofs, nil
}

func (p *KzgEncoder) proofWorker(
	polyFr []bls.Fr,
	jobChan <-chan uint64,
	l uint64,
	results chan<- WorkerResult,
) {
	for j := range jobChan {
		points, err := p.ProveSlices(polyFr, j, l)
		results <- WorkerResult{
			points: points,
			err:    err,
		}
	}
}

// output is in the form see primeField toeplitz
//
// phi ^ (coset size ) = 1
//
// implicitly pad slices to power of 2
func (p *KzgEncoder) ProveSlices(polyFr []bls.Fr, j, l uint64) ([]bls.G1Point, error) {
	// there is a constant term
	m := uint64(len(polyFr)) - 1
	dim := (m - j) / l
	dimE := CeilIntPowerOf2Num(dim)

	toeV := make([]bls.Fr, 2*dimE-1)
	for i := uint64(0); i < dim; i++ {
		bls.CopyFr(&toeV[i], &polyFr[m-(j+i*l)])
	}
	// pad the rest of mat to 0
	for i := dim; i < 2*dimE-1; i++ {
		bls.CopyFr(&toeV[i], &bls.ZERO)
	}

	// use precompute table
	tm, err := toeplitz.NewToeplitz(toeV, p.SFs)
	if err != nil {
		return nil, err
	}

	return tm.MultiplyPoints(p.FFTPoints[j], false, true)
}

/*
returns the power of 2 which is immediately bigger than the input
*/
func CeilIntPowerOf2Num(d uint64) uint64 {
	nextPower := math.Ceil(math.Log2(float64(d)))
	return uint64(math.Pow(2.0, nextPower))
}
