package disperser

import (
	"errors"
)

// aggregator
var ErrInconsistentChainStateFrame = errors.New("Expect number of frames is different from number of registrant from chain state")
var ErrInsufficientEigenSigs = errors.New("Cannot collect sufficient Eigen signatures from data layr nodes")
var ErrInsufficientEthSigs = errors.New("Cannot collect sufficient Eth signatures from data layr nodes")

var ErrRegistrantNonActiveOnChain = errors.New("ErrRegistrantNonActiveOnChain")
var ErrAggregateUnknown = errors.New("ErrAggregateUnknown")

var ErrCouldntFindRegistrant = errors.New("ErrCouldntFindRegistrant")

var ErrDisperseReplyNotOK = errors.New("ErrDisperseReplyNotOK")
var ErrPrecommitTimeout = errors.New("ErrPrecommitTimeout")
var ErrPrecommitQuorumInconsistent = errors.New("ErrPrecommitQuorumInconsistent")
var ErrInvalidInputDuration = errors.New("ErrInvalidInputDuration. Must be smaller than 15 hours")
var ErrInvalidInputLength = errors.New("ErrInvalidInputLength. Must be greater than 31 * numDatalayrNode")

var ErrInvalidAggSig = errors.New("ErrInvalidAggSig")
var ErrInconsistentAggPub = errors.New("ErrInconsistentAggPub")
var ErrNotEnoughParticipants = errors.New("Not enough participants")
