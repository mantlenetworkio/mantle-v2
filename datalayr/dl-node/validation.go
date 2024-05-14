package dln

import (
	"context"
	"time"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/header"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/encoder"
	"github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
)

func (s *Server) ValidateDataStore(ctx context.Context, msgHash []byte, frameBytes [][]byte) error {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering ValidateDataStore function...")
	defer log.Trace().Msg("Exiting ValidateDataStore function...")

	if s.Dln.store.HasCommit(ctx, msgHash) {
		log.Warn().Err(pb.ErrSavedAlready).Msg("Store already commited")
		return pb.ErrSavedAlready
	}

	// search graphql db for InitDataStore entities with the msgHash provided in the request
	ds, ok := s.Dln.GraphClient.PollingInitDataStoreByMsgHash(
		ctx,
		msgHash,
		s.Config.Timeout/2, // 3 retries, should be less than disperser
	)

	// TODO: Do we need to check amount of funds
	if !ok {
		log.Error().Msg("No InitDataStore found")
		return pb.ErrInsufficientFund
	}

	// Get state view and extract relevant items
	stateView, err := s.Dln.GraphClient.GetStateView(ctx, ds.ReferenceBlockNumber)
	if err != nil {
		log.Error().Err(err).Msg("Could not get stateView")
		return pb.ErrLocalLostGraphState
	}
	advRatioBasisPoints, err := s.ChainClient.GetAdversaryThresholdBasisPoints()
	if err != nil {
		return pb.ErrLocalLostGraphState
	}
	liveRatioBasisPoints, err := s.ChainClient.GetQuorumThresholdBasisPoints()
	if err != nil {
		return pb.ErrLocalLostGraphState
	}

	addr := common.HexToAddress(s.Config.Address)
	reg, ok := stateView.RegistrantMap[addr]
	if !ok {
		log.Error().Err(err).Msg("Node is not a part of registry")
		return pb.ErrLocalNotInState
	}

	quorumIndex := 0
	params, assignment, err := encoding.GetOperatorAssignment(liveRatioBasisPoints, advRatioBasisPoints, stateView, ds.DataCommitment, quorumIndex, int(reg.Index))
	if err != nil {
		log.Error().Err(err).Msg("Could not get chunk assignment")
		return pb.ErrLocalLostGraphState
	}

	// decode the the raw bytes in the header posted on chain into a header object
	dsHeader, err := header.DecodeDataStoreHeader(ds.Header[:])
	if err != nil {
		log.Error().Err(err).Msg("Could not decode datastore header")
		return pb.ErrDecodeDataStore
	}

	// Decode the frames and ensure they are the same length

	if len(frameBytes) != int(assignment.NumChunks) {
		log.Error().Err(err).Msg("Incorrect number of chunks received")
		return pb.ErrIncorrectNumberOfChunks
	}

	frames := make([]kzgRs.Frame, len(frameBytes))
	frameLen := -1

	for ind, rawFrame := range frameBytes {

		frame, err := kzgRs.Decode(rawFrame)
		if err != nil {
			log.Error().Err(err).Msgf("Decoding error")
			return pb.ErrDecodeFrame
		}
		if frameLen == -1 {
			frameLen = len(frame.Coeffs)
		} else if frameLen != len(frame.Coeffs) {
			log.Error().Err(err).Msg("Inconsistent frame sizes")
			return pb.ErrDecodeFrame
		}
		frames[ind] = frame
	}

	// validate the header is consistent with the decoded frame
	err = s.validateDataStoreHeader(ctx, &dsHeader, params, uint32(frameLen))
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate store header")
		return err
	}

	// Verify low degree proof, it ensures that following header fields are consistent
	// 1. kzgCommit, 2. Low Degree Proof, 3. Degree, 4. NumSys
	// Verify low degree proof

	// Verify Multireveal proof, it ensures the followings are consistent
	// 1. kzgCommit, 2. Degree, 3. NumSys, 4. NumPar, 5. frame.Coeffs, 6. frame.Proof,
	// since 1, 2, 3 are verified in the last section, 4 is checked against total nodes
	// all that is left is to check frame.

	verifyStart := time.Now()

	log.Trace().Msgf("disperser %v. Headerhash %v", hexutil.Encode(dsHeader.Disperser[:]), hexutil.Encode(ds.DataCommitment[:]))
	log.Trace().Msgf("Registrant index %v get frame  %v", reg.Index, assignment.ChunkIndex)

	verifier, err := s.kzgEncoderGroup.GetKzgVerifier(uint64(dsHeader.NumSys), uint64(dsHeader.NumPar), uint64(dsHeader.OrigDataSize))
	if err != nil {
		log.Error().Err(err).Msg("Failed to get KZG verifier")
		return err
	}

	err = verifier.VerifyCommit((*bn254.G1Point)(bls.DeserializeG1(dsHeader.KzgCommit[:])), (*bn254.G1Point)(bls.DeserializeG1(dsHeader.LowDegreeProof[:])))
	if err != nil {
		log.Error().Err(err).Msg("Low degree verification failed")
		return err
	}

	indices := assignment.GetIndices()
	for ind, frame := range frames {
		err = verifier.VerifyFrame(
			(*bn254.G1Point)(bls.DeserializeG1(dsHeader.KzgCommit[:])),
			&frame,
			indices[ind],
		)

		if err != nil {
			log.Error().Err(err).Msg("Frame verification failed")
			return err
		}
		log.Trace().Msgf("Frame verification succeeded for chunk %v", ind)
	}

	verifyEnd := time.Now()
	log.Debug().Msgf("Verifying frame takes %v", verifyEnd.Sub(verifyStart))
	log.Info().Msg("Frame verification suceeded!")

	return nil
}

// ValidateDataStoreHeader makes sure header is consistent
// to the state on chain, and data
// user should check numPar to ensure data is coded in the right ratio
func (s *Server) validateDataStoreHeader(
	ctx context.Context,
	dsHeader *header.DataStoreHeader,
	chainParams encoding.QuorumParams,
	frameLen uint32,
) error {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering validateDataStoreHeader function...")
	defer log.Trace().Msg("Exiting validateDataStoreHeader function...")

	// Make Sure NumSys, NumPar in the header is consistent with the values that we would calculate using the parameters on chain
	if dsHeader.NumSys != uint32(chainParams.NumSys) {
		log.Warn().Uint32("header.NumSys", dsHeader.NumSys).Uint64("chainParams.NumSys", chainParams.NumSys).
			Msgf("NumSys from header does not match calculated value from state")
		return pb.ErrInconsistantTotalNodes
	}

	if dsHeader.NumPar != uint32(chainParams.NumPar) {
		log.Warn().Uint32("header.NumPar", dsHeader.NumPar).Uint64("chainParams.NumPar", chainParams.NumPar).
			Msgf("NumPar from header does not match calculated value from state")
		return pb.ErrInconsistantTotalNodes
	}

	// check that the number degree from the header is consistent with the
	// claimed original size of the data and the number of systematic chunks
	headerParams := rs.GetEncodingParams(uint64(dsHeader.NumSys), uint64(dsHeader.NumSys), uint64(dsHeader.OrigDataSize))
	if dsHeader.Degree != uint32(headerParams.ChunkLen) {
		log.Warn().Msgf(
			"origDataSize %v, numSys %v, degree %v, calculated %v",
			dsHeader.OrigDataSize,
			dsHeader.NumSys,
			dsHeader.Degree,
			headerParams.ChunkLen,
		)
		return pb.ErrInconsistantDegreeAndOrigDataSize
	}

	// check that the degree of the polynomial in the header is equal to the
	// length of the data being sent
	if dsHeader.Degree != frameLen {
		log.Warn().Msgf(
			"Frame size received %v != Degree claimed in the header %v",
			dsHeader.Degree,
			frameLen,
		)
		return pb.ErrInconsistantTotalNodes
	}

	// ToDo: Add Transaction sender Address to graphView, InitDataStoreEvent, and
	// compare against header

	return nil
}
