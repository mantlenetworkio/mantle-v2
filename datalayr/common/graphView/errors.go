package graphView

import "errors"

var (
	ErrSubgraphTimeout                      = errors.New("timed out while waiting for subgraph")
	ErrOperatorNotFound                     = errors.New("operator does not exist")
	ErrEmptyTotalStakes                     = errors.New("empty response about TotalStake")
	ErrEmptyTotalOperators                  = errors.New("empty response about TotalOperators")
	ErrFeeStringNotParseable                = errors.New("fee string not ok")
	ErrEthSignedStringNotParseable          = errors.New("eth signed string not ok")
	ErrEigenSignedStringNotParseable        = errors.New("eigen signed string not ok")
	ErrMantleFirstStakedStringNotParseable  = errors.New("eth staked string not ok")
	ErrMantleSencodStakedStringNotParseable = errors.New("eigen staked string not ok")
	ErrInvalidG1PubkeyLength                = errors.New("invalid G1 pk length")
	ErrInvalidG2PubkeyLength                = errors.New("invalid G2 pk length")
)
