package kzgEncoder

import (
	"errors"
	"math"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
	kzg "github.com/Layr-Labs/datalayr/lib/encoding/kzg"
	wbls "github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
)

type KzgVerifier struct {
	*KzgConfig
	Srs *kzg.SRS

	rs.EncodingParams

	Fs *kzg.FFTSettings
	Ks *kzg.KZGSettings
}

func (g *KzgEncoderGroup) GetKzgVerifier(numSys, numPar, dataByteLen uint64) (*KzgVerifier, error) {

	// key := EncoderKey{numSys, numPar, dataByteLen}

	params := rs.GetEncodingParams(numSys, numPar, dataByteLen)
	ver, ok := g.Verifiers[params]
	if ok {
		return ver, nil
	}

	ver, err := g.NewKzgVerifier(numSys, numPar, dataByteLen)
	if err == nil {
		g.Verifiers[params] = ver
	}

	return ver, err
}

func (g *KzgEncoderGroup) NewKzgVerifier(numSys, numPar, dataByteLen uint64) (*KzgVerifier, error) {

	params := rs.GetEncodingParams(numSys, numPar, dataByteLen)

	n := uint8(math.Log2(float64(params.PaddedNodeGroupSize)))
	fs := kzg.NewFFTSettings(n)
	ks, err := kzg.NewKZGSettings(fs, g.Srs)

	if err != nil {
		return nil, err
	}

	return &KzgVerifier{
		KzgConfig:      g.KzgConfig,
		Srs:            g.Srs,
		EncodingParams: params,
		Fs:             fs,
		Ks:             ks,
	}, nil
}

func (v *KzgVerifier) VerifyCommit(commit, lowDegreeProof *wbls.G1Point) error {

	if !VerifyLowDegreeProof(commit, lowDegreeProof, v.GlobalPolyDegree, v.SRSOrder, v.Srs.G2) {
		return errors.New("low degree proof fails")
	}
	return nil

}

func (v *KzgVerifier) VerifyFrame(commit *wbls.G1Point, f *Frame, index uint64) error {

	j, err := rs.GetLeadingCosetIndex(
		uint64(index),
		v.NumSys,
		v.NumPar,
	)
	if err != nil {
		return err
	}

	if !f.Verify(v.Ks, commit, &v.Ks.ExpandedRootsOfUnity[j]) {
		return errors.New("multireveal proof fails")
	}

	return nil

}

// A single thread verifier
func (v *KzgVerifier) Verify(commit, lowDegreeProof *wbls.G1Point, f *Frame, index uint64) error {

	if err := v.VerifyCommit(commit, lowDegreeProof); err != nil {
		return err
	}

	return v.VerifyFrame(commit, f, index)

}
