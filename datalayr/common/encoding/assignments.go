package encoding

import (
	"math/big"

	"github.com/Layr-Labs/datalayr/common/graphView"
)

type ChunkAssignment struct {
	ChunkIndex uint64
	NumChunks  uint64
}

type QuorumParams struct {
	StakeThreshold   *big.Int
	NumSys           uint64
	NumPar           uint64
	ChunksByOperator []uint64
}

const BIP_MULTIPLIER = 10000

func (c *ChunkAssignment) GetIndices() []uint64 {
	indices := make([]uint64, c.NumChunks)
	for ind := range indices {
		indices[ind] = c.ChunkIndex + uint64(ind)
	}
	return indices
}

func roundUpDivide(a, b *big.Int) *big.Int {

	one := new(big.Int).SetUint64(1)
	res := new(big.Int)
	res.Sub(a, one)
	res.Div(res, b)
	res.Add(res, one)
	return res

}

func GetQuorumParams(liveRatioBasisPoints, advRatioBasisPoints *big.Int, stateView *graphView.StateView, quorumIndex int) (QuorumParams, error) {

	numOperators := len(stateView.Registrants)
	numOperatorsBig := new(big.Int).SetUint64(uint64(numOperators))

	// Get stake threshold
	stakeThreshold := new(big.Int)
	stakeThreshold.Mul(liveRatioBasisPoints, stateView.TotalStake.QuorumStakes[quorumIndex])
	stakeThreshold = roundUpDivide(stakeThreshold, new(big.Int).SetUint64(BIP_MULTIPLIER))

	// Get NumSys
	numSys := new(big.Int).Sub(liveRatioBasisPoints, advRatioBasisPoints)  // 9000 - 4000 = 5000
	numSys.Mul(numSys, numOperatorsBig)                                    // 5000 * 6 = 30000
	numSys = roundUpDivide(numSys, new(big.Int).SetUint64(BIP_MULTIPLIER)) // 30000 / 10000 = 3
	numChunks := make([]uint64, numOperators)

	// Get NumPar
	totalChunks := new(big.Int)
	totalStakes := stateView.TotalStake.QuorumStakes[quorumIndex]
	for ind, r := range stateView.Registrants {

		m := new(big.Int).Mul(numOperatorsBig, r.QuorumStakes[quorumIndex])
		m = roundUpDivide(m, totalStakes)

		totalChunks.Add(totalChunks, m)

		numChunks[ind] = m.Uint64()

	}
	numPar := new(big.Int).Sub(totalChunks, numSys) // 9 - 3 = 6
	return QuorumParams{
		StakeThreshold:   stakeThreshold,
		NumSys:           numSys.Uint64(),
		NumPar:           numPar.Uint64(),
		ChunksByOperator: numChunks,
	}, nil

}

func GetOperatorAssignments(params QuorumParams, headerHash [32]byte) []ChunkAssignment {

	numOperators := len(params.ChunksByOperator)
	currentIndex := uint64(0)
	assignments := make([]ChunkAssignment, numOperators)

	for orderedInd := range params.ChunksByOperator {

		// Find the operator that should be at index currentIndex
		operatorInd := GetOperatorAtIndex(headerHash, orderedInd, numOperators)
		m := params.ChunksByOperator[operatorInd]

		assignments[operatorInd] = ChunkAssignment{
			ChunkIndex: currentIndex,
			NumChunks:  m,
		}

		currentIndex += m

	}

	return assignments

}

// Returns the operator at a given index within the reordered sequence. We reorder the sequence
// by letting the reordered_index = operator_index + headerHash.
// Thus, get get the operator at a given reordered_index, we simply reverse:
// operator_index = reordered_index - headerHash
func GetOperatorAtIndex(headerHash [32]byte, index, numOperators int) int {

	indexBig := new(big.Int).SetUint64(uint64(index))
	offset := new(big.Int).SetBytes(headerHash[:])

	operatorIndex := new(big.Int).Sub(indexBig, offset)

	operatorIndex.Mod(operatorIndex, new(big.Int).SetUint64(uint64(numOperators)))

	return int(operatorIndex.Uint64())
}

func GetOperatorAssignment(liveRatioBasisPoints, advRatioBasisPoints *big.Int, stateView *graphView.StateView, headerHash [32]byte, quorumIndex, operatorIndex int) (QuorumParams, ChunkAssignment, error) {

	params, err := GetQuorumParams(liveRatioBasisPoints, advRatioBasisPoints, stateView, quorumIndex)
	if err != nil {
		return QuorumParams{}, ChunkAssignment{}, err
	}
	assignments := GetOperatorAssignments(params, headerHash)

	return params, assignments[operatorIndex], nil

}
