package bls

import (
	"math/big"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

type G1Point struct {
	X *big.Int
	Y *big.Int
}

type G2Point struct {
	X [2]*big.Int
	Y [2]*big.Int
}

func GetG1Generator() *bn254.G1Affine {
	g1Gen := new(bn254.G1Affine)
	g1Gen.X.SetString("1")
	g1Gen.Y.SetString("2")
	return g1Gen
}

func GetG2Generator() *bn254.G2Affine {
	g2Gen := new(bn254.G2Affine)
	g2Gen.X.SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781",
		"11559732032986387107991004021392285783925812861821192530917403151452391805634")
	g2Gen.Y.SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930",
		"4082367875863433681332203403145435568316851327593401208105741076214120093531")
	return g2Gen
}

func CheckG1AndG2DiscreteLogEquality(pointG1 *bn254.G1Affine, pointG2 *bn254.G2Affine) (bool, error) {
	negGenG1 := new(bn254.G1Affine).Neg(GetG1Generator())
	return bn254.PairingCheck([]bn254.G1Affine{*pointG1, *negGenG1}, []bn254.G2Affine{*GetG2Generator(), *pointG2})
}

func SerializeG1(p *bn254.G1Affine) []byte {
	b := make([]byte, 0)
	tmp := p.X.Bytes()
	for i := 0; i < 32; i++ {
		b = append(b, tmp[i])
	}
	tmp = p.Y.Bytes()
	for i := 0; i < 32; i++ {
		b = append(b, tmp[i])
	}
	return b
}

func DeserializeG1(b []byte) *bn254.G1Affine {
	p := new(bn254.G1Affine)
	p.X.SetBytes(b[0:32])
	p.Y.SetBytes(b[32:64])
	return p
}

func SerializeG2(p *bn254.G2Affine) []byte {
	b := make([]byte, 0)
	tmp := p.X.A0.Bytes()
	for i := 0; i < 32; i++ {
		b = append(b, tmp[i])
	}
	tmp = p.X.A1.Bytes()
	for i := 0; i < 32; i++ {
		b = append(b, tmp[i])
	}
	tmp = p.Y.A0.Bytes()
	for i := 0; i < 32; i++ {
		b = append(b, tmp[i])
	}
	tmp = p.Y.A1.Bytes()
	for i := 0; i < 32; i++ {
		b = append(b, tmp[i])
	}
	return b
}

func DeserializeG2(b []byte) *bn254.G2Affine {
	p := new(bn254.G2Affine)
	p.X.A0.SetBytes(b[0:32])
	p.X.A1.SetBytes(b[32:64])
	p.Y.A0.SetBytes(b[64:96])
	p.Y.A1.SetBytes(b[96:128])
	return p
}

func ConvertFrameBlsKzgToBytes(castedG1Point *bn254.G1Affine) [64]byte {
	castedG1LowDegreeProof := *(castedG1Point)
	lowDegreeProof := SerializeG1(&castedG1LowDegreeProof)
	var lowDegreeProof64 [64]byte
	copy(lowDegreeProof64[:], lowDegreeProof[:])
	return lowDegreeProof64
}

func mulByGeneratorG1(a *fr.Element) *bn254.G1Affine {
	var g1Gen bn254.G1Affine
	g1Gen.X.SetString("1")
	g1Gen.Y.SetString("2")

	return new(bn254.G1Affine).ScalarMultiplication(&g1Gen, a.ToBigIntRegular(new(big.Int)))
}

func mulByGeneratorG2(a *fr.Element) *bn254.G2Affine {
	var g2Gen bn254.G2Affine
	g2Gen.X.SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781",
		"11559732032986387107991004021392285783925812861821192530917403151452391805634")
	g2Gen.Y.SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930",
		"4082367875863433681332203403145435568316851327593401208105741076214120093531")

	return new(bn254.G2Affine).ScalarMultiplication(&g2Gen, a.ToBigIntRegular(new(big.Int)))
}

func ToG1Point(p *bn254.G1Affine) G1Point {
	return G1Point{X: p.X.ToBigIntRegular(new(big.Int)), Y: p.Y.ToBigIntRegular(new(big.Int))}
}

func ToG2Point(p *bn254.G2Affine) G2Point {
	return G2Point{X: [2]*big.Int{p.X.A1.ToBigIntRegular(new(big.Int)), p.X.A0.ToBigIntRegular(new(big.Int))}, Y: [2]*big.Int{p.Y.A1.ToBigIntRegular(new(big.Int)), p.Y.A0.ToBigIntRegular(new(big.Int))}}
}
