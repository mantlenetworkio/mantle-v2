package kzgEncoder

import (
	"context"
	"errors"
	"log"
	"math"
	"time"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
	kzg "github.com/Layr-Labs/datalayr/lib/encoding/kzg"
	"github.com/Layr-Labs/datalayr/lib/encoding/utils"
	rb "github.com/Layr-Labs/datalayr/lib/encoding/utils/reverseBits"
	"github.com/ethereum/go-ethereum/crypto"

	bls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
)

type KzgConfig struct {
	G1Path    string
	G2Path    string
	CacheDir  string
	NumWorker uint64
	SRSOrder  uint64 // Order is the total size of SRS
	Verbose   bool
}

type KzgEncoderGroup struct {
	*KzgConfig
	Srs *kzg.SRS

	Encoders  map[rs.EncodingParams]*KzgEncoder
	Verifiers map[rs.EncodingParams]*KzgVerifier
}

type KzgEncoder struct {
	*rs.Encoder

	*KzgConfig
	Srs *kzg.SRS

	Fs        *kzg.FFTSettings
	Ks        *kzg.KZGSettings
	SFs       *kzg.FFTSettings // fft used for submatrix product helper
	FFTPoints [][]bls.G1Point
}

func NewKzgEncoderGroup(config *KzgConfig) (*KzgEncoderGroup, error) {
	// read the whole order, and treat it as entire SRS for low degree proof
	s1, err := utils.ReadG1Points(config.G1Path, config.SRSOrder, config.NumWorker)
	if err != nil {
		log.Println("failed to read G1 points", err)
		return nil, err
	}
	s2, err := utils.ReadG2Points(config.G2Path, config.SRSOrder, config.NumWorker)
	if err != nil {
		log.Println("failed to read G2 points", err)
		return nil, err
	}

	srs, err := kzg.NewSrs(s1, s2)
	if err != nil {
		log.Println("Could not create srs", err)
		return nil, err
	}

	return &KzgEncoderGroup{
		KzgConfig: config,
		Srs:       srs,
		Encoders:  make(map[rs.EncodingParams]*KzgEncoder),
		Verifiers: make(map[rs.EncodingParams]*KzgVerifier),
	}, nil

}

func (g *KzgEncoderGroup) GetKzgEncoder(numSys, numPar, dataByteLen uint64) (*KzgEncoder, error) {
	params := rs.GetEncodingParams(numSys, numPar, dataByteLen)
	enc, ok := g.Encoders[params]
	if ok {
		return enc, nil
	}

	enc, err := g.NewKzgEncoder(numSys, numPar, dataByteLen)
	if err == nil {
		g.Encoders[params] = enc
	}

	return enc, err
}

func (g *KzgEncoderGroup) NewKzgEncoder(numSys, numPar, dataByteLen uint64) (*KzgEncoder, error) {
	encoder, err := rs.NewEncoder(numSys, numPar, dataByteLen, g.Verbose)
	if err != nil {
		log.Println("Could not create encoder", err)
		return nil, err
	}

	subTable, err := NewSRSTable(g.CacheDir, g.Srs.G1, g.NumWorker)
	if err != nil {
		log.Println("Could not create srs table", err)
		return nil, err
	}

	fftPoints, err := subTable.GetSubTables(encoder.NumNodeE, encoder.ChunkLen)
	if err != nil {
		log.Println("could not get sub tables", err)
		return nil, err
	}

	n := uint8(math.Log2(float64(encoder.PaddedNodeGroupSize)))
	fs := kzg.NewFFTSettings(n)

	ks, err := kzg.NewKZGSettings(fs, g.Srs)
	if err != nil {
		return nil, err
	}

	t := uint8(math.Log2(float64(2 * encoder.NumNodeE)))
	sfs := kzg.NewFFTSettings(t)

	return &KzgEncoder{
		Encoder:   encoder,
		KzgConfig: g.KzgConfig,
		Srs:       g.Srs,
		Fs:        fs,
		Ks:        ks,
		SFs:       sfs,
		FFTPoints: fftPoints,
	}, nil
}

// just a wrapper to take bytes not Fr Element
func (g *KzgEncoder) EncodeBytes(ctx context.Context, inputBytes []byte) (*bls.G1Point, *bls.G1Point, []Frame, []uint32, error) {
	if g.Verbose {
		log.Println("Entering EncodeBytes function")
		defer log.Println("Exiting EncodeBytes function")
	}

	inputFr := rs.ToFrArray(inputBytes)
	return g.Encode(inputFr)
}

func (g *KzgEncoder) Encode(inputFr []bls.Fr) (*bls.G1Point, *bls.G1Point, []Frame, []uint32, error) {
	if g.Verbose {
		log.Println("Entering Encode function")
		defer log.Println("Exiting Encode function")
	}

	poly, frames, indices, err := g.Encoder.Encode(inputFr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// compute commit for the full poly
	commit := g.Commit(poly.Coeffs)

	intermediate := time.Now()

	if g.Verbose {
		log.Printf("    Commiting takes  %v\n", time.Since(intermediate))
		intermediate = time.Now()

		log.Printf("shift %v\n", g.SRSOrder-g.PaddedSysGroupSize)
		log.Printf("order %v\n", len(g.Srs.G2))
		log.Println("low degree verification info")
	}

	shiftedSecret := g.Srs.G1[g.SRSOrder-g.PaddedSysGroupSize:]

	//The proof of low degree is commitment of the polynomial shifted to the largest srs degree
	lowDegreeProof := bls.LinCombG1(shiftedSecret, poly.Coeffs[:g.PaddedSysGroupSize])
	//fmt.Println("kzgFFT lowDegreeProof", lowDegreeProof, "poly len ", len(fullCoeffsPoly), "order", len(g.Ks.SecretG2) )
	ok := VerifyLowDegreeProof(&commit, lowDegreeProof, g.GlobalPolyDegree, g.SRSOrder, g.Srs.G2)
	if !ok {
		log.Printf("Kzg FFT Cannot Verify low degree proof %v", lowDegreeProof)
		return nil, nil, nil, nil, errors.New("cannot verify low degree proof")
	} else {
		log.Printf("Kzg FFT Verify low degree proof  PPPASSS %v", lowDegreeProof)
	}

	if g.Verbose {
		log.Printf("    Generating Low Degree Proof takes  %v\n", time.Since(intermediate))
		intermediate = time.Now()
	}

	// compute proofs
	proofs, err := g.ProveAllCosetThreads(poly.Coeffs, g.NumNodeE, g.ChunkLen, g.NumWorker)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if g.Verbose {
		log.Printf("    Proving takes    %v\n", time.Since(intermediate))
	}

	kzgFrames := make([]Frame, len(frames))
	for i, index := range indices {
		kzgFrames[i] = Frame{
			Proof:  proofs[index],
			Coeffs: frames[i].Coeffs,
		}
	}

	// Perform zero padding only if NumSys is not a power of 2
	// TODO: Disabling zero padding proof due to CPU issues
// 	if g.NumSys != g.NumSysE {
// 		paddingproof, paddingQuotientPolyCommit, _ := g.ProveZeroPadding(poly.Coeffs, commit)
// 		verificationFlag := g.VerifyZeroPadding(paddingproof, paddingQuotientPolyCommit, &commit)
// 		if !verificationFlag {
// 			log.Printf("Kzg FFT Cannot Verify zero padding proof %v\n", paddingproof)
// 			return nil, nil, nil, nil, errors.New("cannot verify zero padding proof")
// 		} else {
// 			log.Printf("Kzg FFT can verify zero padding proof  PPPASSS %v\n", lowDegreeProof)
// 			log.Printf("Verification flag is %v\n", verificationFlag)
// 		}
// 	}

	return &commit, lowDegreeProof, kzgFrames, indices, nil

}

func (g *KzgEncoder) Commit(polyFr []bls.Fr) bls.G1Point {
	if g.Verbose {
		log.Println("Entering Commit function")
		defer log.Println("Exiting Commit function")
	}

	commit := g.Ks.CommitToPoly(polyFr)
	return *commit
}

// This function is called for generating the proof for zero padding.
func (g *KzgEncoder) ProveZeroPadding(metaPoly []bls.Fr, metaPolyCommit bls.G1Point) (*bls.G1Point, *bls.G1Point, error) {
	if g.Verbose {
		log.Println("Entering ProveZeroPadding function")
		defer log.Println("Exiting ProveZeroPadding function")
	}

	// construct the polynomial vanishingPoly(x)
	// start by initalizing it first to  a constant poly
	vanishingPoly := make([]bls.Fr, g.ChunkLen*(g.NumSysE-g.NumSys)+1)
	vanishingPoly[0] = bls.ToFr("1")
	for i := g.NumSys; i < g.NumSysE; i++ {

		// multiplying the zero poly with the running product
		prod := g.ZeroPolyMul(vanishingPoly[:int(g.ChunkLen*(i-g.NumSys)+1)], i)

		copy(vanishingPoly[:len(prod)], prod)
	}

	// getting the polynomial: quotientPoly(x)
	quotientPoly := kzg.PolyLongDiv(metaPoly, vanishingPoly[:])

	// getting commitment at G1: [quotientPoly(x)]_1
	quotientPolyCommit := g.Commit(quotientPoly)

	// getting random challenge as part of Fiat-Shamir heurestic
	byteArray := [][32]byte{quotientPolyCommit.X.Bytes(),
		quotientPolyCommit.Y.Bytes(),
		metaPolyCommit.X.Bytes(),
		metaPolyCommit.Y.Bytes(),
	}
	alpha := createFiatShamirChallenge(byteArray)

	// evaluate vanishing polynomial vanishingPoly(x) at the challenge alpha
	var vanishingPolyEval bls.Fr
	bls.EvalPolyAt(&vanishingPolyEval, vanishingPoly, alpha)

	// construct the numerator polynomial metaPoly(x) - vanishingPoly(alpha)*quotientPoly(x)
	// we are storing the numerator polynomial in metaPoly itself
	var tmp2 bls.Fr
	for i := 0; i < len(quotientPoly); i++ {
		bls.MulModFr(&tmp2, &vanishingPolyEval, &quotientPoly[i])
		bls.SubModFr(&metaPoly[i], &metaPoly[i], &tmp2)
	}

	// evaluate proof for zero padding
	zeroPaddingProof := g.Ks.ComputeProofSingleAtFr(metaPoly, *alpha)
	return zeroPaddingProof, &quotientPolyCommit, nil
}

// This function verifies that the proof generated for zero padding is correct.
func (g *KzgEncoder) VerifyZeroPadding(zeroPaddingProof *bls.G1Point, quotientPolyCommit *bls.G1Point, metaPolyCommit *bls.G1Point) bool {
	if g.Verbose {
		log.Println("Entering VerifyZeroPadding function")
		defer log.Println("Exiting VerifyZeroPadding function")
	}

	// getting random challenge alpha as part of Fiat-Shamir heurestic
	byteArray := [][32]byte{quotientPolyCommit.X.Bytes(),
		quotientPolyCommit.Y.Bytes(),
		metaPolyCommit.X.Bytes(),
		metaPolyCommit.Y.Bytes(),
	}
	alpha := createFiatShamirChallenge(byteArray)

	// construct the big vanishing poly vanishingPoly(x) and initalizing it to first zero poly
	vanishingPoly := make([]bls.Fr, g.ChunkLen*(g.NumSysE-g.NumSys)+1)
	vanishingPoly[0] = bls.ToFr("1")
	for i := g.NumSys; i < g.NumSysE; i++ {
		// multiplying the zero poly with the running product
		prod := g.ZeroPolyMul(vanishingPoly[:int(g.ChunkLen*(i-g.NumSys)+1)], i)
		copy(vanishingPoly[:len(prod)], prod)

	}

	// evaluate vanishing polynomial at alpha: vanishingPoly(alpha)
	var vanishingPolyEval bls.Fr
	bls.EvalPolyAt(&vanishingPolyEval, vanishingPoly, alpha)

	// pairing check
	// compute alpha[zeroPaddingProof]_1 + [metaPoly(X)]_1 - vanishingPolyEval[quotientPolyCommit]_1
	var summand bls.G1Point
	bls.MulG1(&summand, zeroPaddingProof, alpha)
	bls.AddG1(&summand, metaPolyCommit, &summand)
	var zMulPiG1 bls.G1Point
	bls.MulG1(&zMulPiG1, quotientPolyCommit, &vanishingPolyEval)
	bls.SubG1(&summand, &summand, &zMulPiG1)

	return bls.PairingsVerify(&summand, &bls.GenG2, zeroPaddingProof, &g.Srs.G2[1])

}

// The function verify low degree proof against a poly commitment
// We wish to show x^shift poly = shiftedPoly, with
// With shift = SRSOrder-1 - claimedDegree and
// proof = commit(shiftedPoly) on G1
// so we can verify by checking
// e( commit_1, [x^shift]_2) = e( proof_1, G_2 )
func VerifyLowDegreeProof(poly, proof *bls.G1Point, claimedDegree, SRSOrder uint64, srsG2 []bls.G2Point) bool {
	return bls.PairingsVerify(poly, &srsG2[SRSOrder-1-claimedDegree], proof, &bls.GenG2)
}

// get Fiat-Shamir challenge
func createFiatShamirChallenge(byteArray [][32]byte) *bls.Fr {
	alphaBytesTmp := make([]byte, 0)
	for i := 0; i < len(byteArray); i++ {
		for j := 0; j < len(byteArray[i]); j++ {
			alphaBytesTmp = append(alphaBytesTmp, byteArray[i][j])
		}
	}
	alphaBytes := crypto.Keccak256(alphaBytesTmp)
	alpha := new(bls.Fr)
	bls.FrSetBytes(alpha, alphaBytes)

	return alpha
}

// invert the divisor, then multiply
func polyFactorDiv(dst *bls.Fr, a *bls.Fr, b *bls.Fr) {
	// TODO: use divmod instead.
	var tmp bls.Fr
	bls.InvModFr(&tmp, b)
	bls.MulModFr(dst, &tmp, a)
}

// Multiplying the zero polynomial
func (g *KzgEncoder) ZeroPolyMul(f []bls.Fr, index uint64) []bls.Fr {
	prod := make([]bls.Fr, len(f)+int(g.ChunkLen))
	copy(prod[int(g.ChunkLen):], f)

	// Observe that we can write (⍵^i * φ)^{ChunkLenE} = (⍵^i * ⍵^{NumSysE})^{ChunkLenE} = ⍵^{(i+NumSysE)*ChunkLenE}
	// ATTENTION: Due to butterfly algorithm used in FFT, we need to put
	// 			  i = rb.ReverseBitsLimited(uint32(g.NumNodeE), uint32(index)) instead of index to evaluate the
	// 			  corresponding power of ⍵.
	// TRIVIA: The root of this zero polynomial is given by g.Fs.ExpandedRootsOfUnity[(rb.ReverseBitsLimited(uint32(g.NumNodeE), uint32(index)))]
	constcoeff := g.Ks.ExpandedRootsOfUnity[uint32(g.ChunkLen)*(rb.ReverseBitsLimited(uint32(g.NumNodeE), uint32(index)))]

	for i := 0; i < len(f); i++ {
		bls.MulModFr(&f[i], &f[i], &constcoeff)
		bls.SubModFr(&prod[i], &prod[i], &f[i])
	}

	return prod

}
