package common

import (
	"errors"
	"github.com/Layr-Labs/datalayr/common/header"
	"math/big"
)

func CreateUploadHeader(params *StoreParams) ([]byte, error) {
	var kzgCommitArray [64]byte
	copy(kzgCommitArray[:], params.KzgCommit)
	var lowDegreeProof [64]byte
	copy(lowDegreeProof[:], params.LowDegreeProof)
	var disperserArray [20]byte
	copy(disperserArray[:], params.Disperser)

	h := header.DataStoreHeader{
		KzgCommit:      kzgCommitArray,
		LowDegreeProof: lowDegreeProof,
		Degree:         params.Degree,
		NumSys:         params.NumSys,
		NumPar:         params.NumPar,
		OrigDataSize:   params.OrigDataSize,
		Disperser:      disperserArray,
	}
	uploadHeader, _, err := header.CreateUploadHeader(h)
	if err != nil {
		return nil, err
	}
	return uploadHeader, nil
}

func MakeCalldata(
	params *StoreParams,
	meta DisperseMeta,
	storeNumber uint32,
	msgHash [32]byte,
) ([]byte, error) {

	totalStakeIndexBytes, err := bigIntToBytes(
		new(big.Int).SetUint64(meta.TotalStakeIndex),
		6,
	)
	if err != nil {
		return nil, err
	}

	storeNumberBytes, err := bigIntToBytes(
		new(big.Int).SetUint64(uint64(storeNumber)),
		4,
	)
	if err != nil {
		return nil, err
	}

	referenceBlockNumberBytes, err := bigIntToBytes(
		new(big.Int).SetUint64(uint64(params.ReferenceBlockNumber)),
		4,
	)
	if err != nil {
		return nil, err
	}

	numNonPubKeysBytes, err := bigIntToBytes(
		new(big.Int).SetUint64(uint64(len(meta.Sigs.NonSignerPubkeys))),
		4,
	)
	if err != nil {
		return nil, err
	}

	flattenedNonPubKeysBytes := make([]byte, 0)
	for i := 0; i < len(meta.Sigs.NonSignerPubkeys); i++ {
		flattenedNonPubKeysBytes = append(
			flattenedNonPubKeysBytes,
			meta.Sigs.NonSignerPubkeys[i]...,
		)
	}

	apkIndexBytes, err := bigIntToBytes(
		new(big.Int).SetUint64(uint64(meta.ApkIndex)),
		4,
	)
	if err != nil {
		return nil, err
	}

	var calldata []byte
	calldata = append(calldata, msgHash[:]...)
	calldata = append(calldata, totalStakeIndexBytes...)
	calldata = append(calldata, referenceBlockNumberBytes...)
	calldata = append(calldata, storeNumberBytes...)
	calldata = append(calldata, numNonPubKeysBytes...)
	calldata = append(calldata, flattenedNonPubKeysBytes...)
	calldata = append(calldata, apkIndexBytes...)
	calldata = append(calldata, meta.Sigs.StoredAggPubkeyG1...)
	calldata = append(calldata, meta.Sigs.UsedAggPubkeyG2...)
	calldata = append(calldata, meta.Sigs.AggSig...)
	return calldata, nil

}

func bigIntToBytes(n *big.Int, packTo int) ([]byte, error) {
	bigIntBytes := n.Bytes()
	bigIntLen := len(bigIntBytes)
	intBytes := make([]byte, packTo)

	if bigIntLen > packTo {
		return nil, errors.New("cannot pad bytes: Desired length is less than existing length")
	}

	for i := 0; i < bigIntLen; i++ {
		intBytes[packTo-1-i] = bigIntBytes[bigIntLen-1-i]
	}
	return intBytes, nil
}
