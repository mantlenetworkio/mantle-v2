package bls

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"

	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/consensys/gnark-crypto/ecc/bn254/fp"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"

	"github.com/ethereum/go-ethereum/crypto"
)

type BlsKeyPair struct {
	secretKey *fr.Element
	PublicKey *bn254.G1Affine // G1 public key = secretKey * g1
}

func VerifyBlsSig(sig *bn254.G1Affine, pubkey *bn254.G2Affine, msgBytes []byte) bool {
	var g2Gen bn254.G2Affine
	g2Gen.X.SetString("10857046999023057135944570762232829481370756359578518086990519993285655852781",
		"11559732032986387107991004021392285783925812861821192530917403151452391805634")
	g2Gen.Y.SetString("8495653923123431417604973247489272438418190587263600148770280649306958101930",
		"4082367875863433681332203403145435568316851327593401208105741076214120093531")

	msgPoint := MapToCurve(msgBytes)

	var negSig bn254.G1Affine
	negSig.Neg((*bn254.G1Affine)(sig))

	P := [2]bn254.G1Affine{*msgPoint, negSig}
	Q := [2]bn254.G2Affine{*pubkey, g2Gen}

	ok, err := bn254.PairingCheck(P[:], Q[:])
	if err != nil {
		fmt.Println("[Bls] Unable to do pairing check.", err)
		return false
	}
	return ok

}

func (k *BlsKeyPair) SignMessage(headerHash []byte) *bn254.G1Affine {
	if len(headerHash) != 32 {
		fmt.Println("SignMessage only on header hash")
		os.Exit(1)
	}

	H := MapToCurve(headerHash)
	sig := new(bn254.G1Affine).ScalarMultiplication(H, k.secretKey.ToBigIntRegular(new(big.Int)))
	return sig
}

func (k *BlsKeyPair) GetPubKeyG1Bytes() [32]byte {
	return k.PublicKey.Bytes()
}

func (k *BlsKeyPair) GetPubKeyPointG2() *bn254.G2Affine {
	return mulByGeneratorG2(k.secretKey)
}

func BlsKeysFromSecretKey(sk *fr.Element) (*BlsKeyPair, error) {
	pk := mulByGeneratorG1(sk)
	return &BlsKeyPair{sk, pk}, nil
}

func BlsKeysFromString(sk string) (*BlsKeyPair, error) {
	_sk, err := new(fr.Element).SetString(sk)
	if err != nil {
		return nil, err
	}

	return BlsKeysFromSecretKey(_sk)
}

func GenRandomBlsKeys() (*BlsKeyPair, error) {

	//Max random value, a 381-bits integer, i.e 2^381 - 1
	max := new(big.Int)
	max.SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

	//Generate cryptographically strong pseudo-random between 0 - max
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, err
	}

	sk := new(fr.Element).SetBigInt(n)
	return BlsKeysFromSecretKey(sk)
}

func HashToCurve(data []byte) *bn254.G1Affine {
	digest := crypto.Keccak256(data)
	return MapToCurve(digest[:])
}

func MapToCurve(digest []byte) *bn254.G1Affine {
	if len(digest) != 32 {
		fmt.Println("only map 32 bytes")
		os.Exit(1)
	}

	//fmt.Println("fp.Modulus", fp.Modulus())

	one := new(big.Int).SetUint64(1)
	three := new(big.Int).SetUint64(3)
	x := new(big.Int)
	x.SetBytes(digest[:])
	for true {
		// y = x^3 + 3
		xP3 := new(big.Int).Exp(x, big.NewInt(3), fp.Modulus())
		y := new(big.Int).Add(xP3, three)
		y.Mod(y, fp.Modulus())
		//fmt.Println("x", x)
		//fmt.Println("y", y)

		if y.ModSqrt(y, fp.Modulus()) == nil {
			x.Add(x, one).Mod(x, fp.Modulus())
		} else {
			var fpX, fpY fp.Element
			fpX.SetBigInt(x)
			fpY.SetBigInt(y)
			return &bn254.G1Affine{
				X: fpX,
				Y: fpY,
			}
		}
	}
	return new(bn254.G1Affine)
}

// ToDo handle errro
func ConvertStringsToG2Point(data [4]string) *bn254.G2Affine {
	public := new(bn254.G2Affine)

	_, err := public.X.A0.SetString(data[0])
	public.X.A1.SetString(data[1])
	public.Y.A0.SetString(data[2])
	public.Y.A1.SetString(data[3])
	_ = err

	return public
}

func ConvertStringsToG1Point(data [2]string) *bn254.G1Affine {
	public := new(bn254.G1Affine)

	_, err := public.X.SetString(data[0])
	public.Y.SetString(data[1])
	_ = err

	return public
}
