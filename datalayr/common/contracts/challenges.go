package contracts

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"math/big"

	wbls "github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/Layr-Labs/datalayr/common/header"

	kzg "github.com/Layr-Labs/datalayr/common/crypto/go-kzg-bn254"
	"github.com/Layr-Labs/datalayr/common/crypto/go-kzg-bn254/bn254"
	"github.com/Layr-Labs/datalayr/lib/merkzg"
)

type DisclosureProver struct {
	SecretG1 []bn254.G1Point
	SecretG2 []bn254.G2Point
}

type Frame struct {
	Proof  bn254.G1Point
	Coeffs []bn254.Fr
}

func (f *Frame) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(f)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func DecodeFrame(b []byte) (Frame, error) {
	var f Frame
	buf := bytes.NewBuffer(b)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&f)
	if err != nil {
		return Frame{}, err
	}
	return f, nil
}

const logMaxNodes = 8
const maxNodes = 1 << logMaxNodes
const logMaxInterpolationPolyDegree = 10
const maxInterpolationPolyDegree = 1 << logMaxInterpolationPolyDegree
const logMaxFullPolyDegree = logMaxNodes + logMaxInterpolationPolyDegree
const maxFullPolyDegree = 1 << logMaxFullPolyDegree

// RootOfUnityScaling = ceilNextPowerOf2(NumSysE + NumPar)
func NewDisclosureProver(secretG1 []bn254.G1Point, secretG2 []bn254.G2Point) *DisclosureProver {
	return &DisclosureProver{
		SecretG1: secretG1,
		SecretG2: secretG2,
	}
}

type MultiRevealProof struct {
	InterpolationPolyCommit [2]*big.Int
	RevealProof             [2]*big.Int
	ZeroPolyCommit          [4]*big.Int
	ZeroPolyProof           []byte
}

func (p *DisclosureProver) ProveBatchInterpolatingPolyDisclosure(
	frames []Frame,
	headerHash [32]byte,
	headerBytes []byte,
	startingIndex uint32,
) ([][]byte, []MultiRevealProof, [4]*big.Int, error) {
	header, err := header.DecodeDataStoreHeader(headerBytes)
	if err != nil {
		return nil, nil, [4]*big.Int{}, err
	}
	// numNode := ceilNextPowOf2(ceilNextPowOf2(header.NumSys) + header.NumPar)
	fs := kzg.NewFFTSettings(logMaxNodes + uint8(math.Log2(float64(header.Degree))))
	// fmt.Println("FS", logMaxNodes+uint8(math.Log2(float64(header.Degree))))
	ks := kzg.NewKZGSettings(fs, p.SecretG1, p.SecretG2)

	interpolationPolys := make([][]bn254.Fr, 0)
	interpolationPolyCommits := make([]bn254.G1Point, 0)
	multiRevealProofs := make([]MultiRevealProof, 0)

	for i := uint32(0); i < uint32(len(frames)); i++ {
		lcIndex, err := kzg.GetLeadingCosetIndex(uint64(startingIndex+i), uint64(header.NumSys), uint64(header.NumPar))
		if err != nil {
			return nil, nil, [4]*big.Int{}, err
		}
		interpolationPoly, interpolationPolyCommit, zeroPolyCommit, zeroPolyProof, err := p.ProveInterpolatingPolyReveal(
			&frames[i],
			header,
			lcIndex,
			*ks,
		)
		if err != nil {
			return nil, nil, [4]*big.Int{}, err
		}
		interpolationPolys = append(interpolationPolys, interpolationPoly)
		interpolationPolyCommits = append(interpolationPolyCommits, *interpolationPolyCommit)
		multiRevealProofs = append(multiRevealProofs, MultiRevealProof{
			InterpolationPolyCommit: [2]*big.Int{
				interpolationPolyCommit.X.ToBigIntRegular(new(big.Int)),
				interpolationPolyCommit.Y.ToBigIntRegular(new(big.Int)),
			},
			RevealProof: [2]*big.Int{
				frames[i].Proof.X.ToBigIntRegular(new(big.Int)),
				frames[i].Proof.Y.ToBigIntRegular(new(big.Int)),
			},
			ZeroPolyCommit: zeroPolyCommit,
			ZeroPolyProof:  zeroPolyProof,
		})
	}

	commitmentEquivalenceProof := ks.ComputeBatchPolynomialEquivalenceProofInG2(interpolationPolys, interpolationPolyCommits)

	commitmentEquivalenceProofBigInt := [4]*big.Int{
		commitmentEquivalenceProof.X.A0.ToBigIntRegular(new(big.Int)),
		commitmentEquivalenceProof.X.A1.ToBigIntRegular(new(big.Int)),
		commitmentEquivalenceProof.Y.A0.ToBigIntRegular(new(big.Int)),
		commitmentEquivalenceProof.Y.A1.ToBigIntRegular(new(big.Int)),
	}
	polys := make([][]byte, len(interpolationPolys))
	for i := 0; i < len(interpolationPolys); i++ {
		polys[i] = make([]byte, 32*len(interpolationPolys[i]))
		for j := 0; j < len(interpolationPolys[i]); j++ {
			frBytes := bn254.FrToBytes(&interpolationPolys[i][j])
			copy(polys[i][j*32:(j+1)*32], frBytes[:])
		}
	}

	// fmt.Println("PROOF", commitmentEquivalenceProof.String())

	return polys, multiRevealProofs, commitmentEquivalenceProofBigInt, nil
}

func (p *DisclosureProver) ProveInterpolatingPolyDisclosure(
	frame *Frame,
	headerBytes []byte,
	index uint32,
	ks kzg.KZGSettings,
) (MultiRevealProof, [4]*big.Int, error) {
	header, err := header.DecodeDataStoreHeader(headerBytes)
	if err != nil {
		return MultiRevealProof{}, [4]*big.Int{}, err
	}
	//make it not work
	// poly[0] = poly[0] + 1
	lcIndex, err := kzg.GetLeadingCosetIndex(uint64(index), uint64(header.NumSys), uint64(header.NumPar))
	if err != nil {
		return MultiRevealProof{}, [4]*big.Int{}, err
	}
	interpolationPoly, interpolationPolyCommit, zeroPolyCommit, zeroPolyProof, err := p.ProveInterpolatingPolyReveal(
		frame,
		header,
		lcIndex,
		ks,
	)
	if err != nil {
		return MultiRevealProof{}, [4]*big.Int{}, err
	}

	multiRevealProof := MultiRevealProof{
		InterpolationPolyCommit: [2]*big.Int{
			interpolationPolyCommit.X.ToBigIntRegular(new(big.Int)),
			interpolationPolyCommit.Y.ToBigIntRegular(new(big.Int)),
		},
		RevealProof: [2]*big.Int{
			frame.Proof.X.ToBigIntRegular(new(big.Int)),
			frame.Proof.Y.ToBigIntRegular(new(big.Int)),
		},
		ZeroPolyCommit: zeroPolyCommit,
		ZeroPolyProof:  zeroPolyProof,
	}

	commitmentEquivalenceProof := ks.ComputePolynomialEquivalenceProofInG2(interpolationPoly, *interpolationPolyCommit)

	commitmentEquivalenceProofBigInt := [4]*big.Int{
		commitmentEquivalenceProof.X.A0.ToBigIntRegular(new(big.Int)),
		commitmentEquivalenceProof.X.A1.ToBigIntRegular(new(big.Int)),
		commitmentEquivalenceProof.Y.A0.ToBigIntRegular(new(big.Int)),
		commitmentEquivalenceProof.Y.A1.ToBigIntRegular(new(big.Int)),
	}

	return multiRevealProof, commitmentEquivalenceProofBigInt, nil
}

func (p *DisclosureProver) ProveInterpolatingPolyReveal(
	frame *Frame,
	header header.DataStoreHeader,
	index uint32,
	ks kzg.KZGSettings,
) ([]bn254.Fr, *bn254.G1Point, [4]*big.Int, []byte, error) {
	interpolationPoly := frame.Coeffs
	numNode := ceilNextPowOf2(ceilNextPowOf2(header.NumSys) + header.NumPar)
	interpolationPolyCommit := bn254.LinCombG1(ks.SecretG1[:len(interpolationPoly)], interpolationPoly)

	// c.Logger.Println("[DLN RESPONSE] poly", hex.EncodeToString(poly))
	// c.Logger.Println("MULTIREVEAL DEGREE", header.Degree)

	merkleProof, zeroPolyCommitment := p.PrepareZeroPolys(uint32(header.Degree), index, numNode, ks)

	zeroPolyCommitmentBigInt := [4]*big.Int{
		zeroPolyCommitment.X.A0.ToBigIntRegular(new(big.Int)),
		zeroPolyCommitment.X.A1.ToBigIntRegular(new(big.Int)),
		zeroPolyCommitment.Y.A0.ToBigIntRegular(new(big.Int)),
		zeroPolyCommitment.Y.A1.ToBigIntRegular(new(big.Int)),
	}

	// optionally verify
	var commitMinusInterpolation bn254.G1Point
	bn254.SubG1(&commitMinusInterpolation, (*bn254.G1Point)(wbls.DeserializeG1(header.KzgCommit[:])), interpolationPolyCommit)

	ok := bn254.PairingsVerify(&commitMinusInterpolation, &bn254.GenG2, &frame.Proof, &zeroPolyCommitment)
	if !ok {
		fmt.Println("kzg verification fails")
	} else {
		fmt.Println("kzg verification succeed")
	}

	// fmt.Println("[DLN RESPONSE] INTERPOLATING POLY COMMIT", interpolationPolyCommit.String(), zeroPolyCommitment.String(), frame.Proof.String())

	return interpolationPoly, interpolationPolyCommit, zeroPolyCommitmentBigInt, merkleProof, nil
}

func ceilNextPowOf2(n uint32) uint32 {
	k := uint32(1)
	for k < n {
		k = k << 1
	}
	return k
}

// index in butterfly order
func (p *DisclosureProver) PrepareZeroPolys(
	degree uint32,
	index uint32,
	numNode uint32,
	ks kzg.KZGSettings,
) ([]byte, bn254.G2Point) {
	zeroPolyCommitments := make([]bn254.G2Point, maxNodes)
	// in natural order
	for i := uint32(0); i < maxNodes; i++ {
		//nodes get smaller leading cosets (aka mor tightly packed on the unit circle)
		//when the degree of the interpolation poly is high
		//when degree is the maxInterpolationPolyDegree
		//nodes leading cosets will be at indexes [0], [1], [2], ...
		//if it is half of max, the cosets will be at [0], [2], ...
		//this is for when there are 256 nodes
		zeroPolyCommitments[i] = p.GetZeroPolyEval(i, degree, ks)
	}
	//scale index by what multiple it is away from maxNodes
	//for example, maxNodes is 256, and if there 256 nodes
	//then node 0 will get [0], 1 will get [1], ...
	//but if there are only 128 nodes
	//then node 0 will get [0], 1 will get [2], ...
	scaledLcIndex := index * maxNodes / numNode
	out := zeroPolyCommitments[scaledLcIndex]

	tree := merkzg.NewG2MerkleTree(zeroPolyCommitments, len(zeroPolyCommitments))

	zeroPolyCommitmentProof := tree.ProveIndex(int(scaledLcIndex))

	// fmt.Println(merkzg.VerifyProof(int(scaledLcIndex), tree.Tree[len(tree.Tree)-1][scaledLcIndex], zeroPolyCommitmentProof, tree.Tree[0][0]))

	flattenedProof := make([]byte, 0)
	for i := 0; i < len(zeroPolyCommitmentProof); i++ {
		flattenedProof = append(flattenedProof, zeroPolyCommitmentProof[i]...)
	}

	//fmt.Println("[FDCDLN] ZERO POLYS", degree, index, scaledLcIndex, maxNodes/numNode, hex.EncodeToString(tree.Tree[0][0]), hex.EncodeToString(tree.Tree[len(tree.Tree)-1][scaledLcIndex]), hex.EncodeToString(flattenedProof), zeroPolyCommitments[scaledLcIndex].String())

	return flattenedProof, out
}

// if 1024th root of unity w (meaning w^1024 = 1)
// then this returns kzg.commit(x^degree - (w^index)^degree)
func (p *DisclosureProver) GetZeroPolyEval(index, degree uint32, ks kzg.KZGSettings) bn254.G2Point {
	//REQUIRES degree is a power of 2
	// fmt.Println(len(p.KS.ExpandedRootsOfUnity))
	x := ks.ExpandedRootsOfUnity[index]
	// c.Logger.Println("x", x)

	var xPow bn254.Fr
	if degree < 1 {
		bn254.CopyFr(&xPow, &bn254.ONE)
	} else {
		xPow = x
		for i := uint32(1); i < degree; i *= 2 {
			bn254.MulModFr(&xPow, &xPow, &xPow)
		}
	}

	var xPowPoint bn254.G2Point
	bn254.MulG2(&xPowPoint, &bn254.GenG2, &xPow)

	var xnMinusYn bn254.G2Point
	bn254.SubG2(&xnMinusYn, &ks.SecretG2[degree], &xPowPoint)
	return xnMinusYn
}
