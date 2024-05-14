package kzg

import (
	"errors"
	"math"
)

// This function is used by user to get the leading coset for a frame, where i is frame index
func GetLeadingCosetIndex(i uint64, numSys, numPar uint64) (uint32, error) {
	numNode := numSys + numPar
	numSysE := uint64(CeilIntPowerOf2Num(numSys))
	ratio := RoundUpDivision(numNode, numSys)
	numNodeE := uint64(CeilIntPowerOf2Num(numSysE * ratio))

	if i < numSys {
		j := ReverseBitsLimited(uint32(numNodeE), uint32(i))
		return j, nil
	} else if i < numNodeE-(numSysE-numSys) {
		j := ReverseBitsLimited(uint32(numNodeE), uint32((i-numSys)+numSysE))
		return j, nil
	}

	return 0, errors.New("Cannot create number of frame higher than possible")
}

func CeilIntPowerOf2Num(d uint64) uint64 {
	nextPower := math.Ceil(math.Log2(float64(d)))
	return uint64(math.Pow(2.0, nextPower))
}

// helper function
func RoundUpDivision(a, b uint64) uint64 {
	return uint64(math.Ceil(float64(a) / float64(b)))
}
