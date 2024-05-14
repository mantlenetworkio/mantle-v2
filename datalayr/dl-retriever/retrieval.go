package retriever

import (
	"context"
	"io"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/Layr-Labs/datalayr/common/encoding"
	"github.com/Layr-Labs/datalayr/common/graphView"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/Layr-Labs/datalayr/common/middleware/correlation"
	"github.com/Layr-Labs/datalayr/common/middleware/logger"
	"github.com/Layr-Labs/datalayr/lib/encoding/kzg/bn254"
	rs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"

	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RetrievalResult struct {
	Err     error
	Frames  []rs.Frame
	Indices []uint64
}

func (v *Retriever) RetrieveOperatorFrames(ctx context.Context, ds *graphView.DataStoreRetrieve, socket string, indices []uint64, verifier *rs.KzgVerifier) ([]rs.Frame, []uint64, error) {
	log := v.Logger.SubloggerId(ctx)
	conn, err := grpc.Dial(
		socket,
		grpc.WithChainStreamInterceptor(
			correlation.StreamClientInterceptor(),
			logger.StreamClientInterceptor(*v.Logger.Logger),
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error().Err(err).Msgf("Retriever cannot connect to %v\n", socket)
		return nil, nil, err
	}
	defer conn.Close()
	c := pb.NewDataDispersalClient(conn)

	ctx, cancel := context.WithTimeout(ctx, v.Timeout)
	defer cancel()

	request := &pb.RetrieveFrameRequest{
		Commit: ds.MsgHash,
	}
	opt := grpc.MaxCallSendMsgSize(1024 * 1024 * 300)
	replyStream, err := c.RetrieveFrame(ctx, request, opt)
	if err != nil {
		return nil, nil, err
	}

	frames := make([]rs.Frame, 0)
	validIndices := make([]uint64, 0)

	var errList error

	count := 0
	for {
		reply, err := replyStream.Recv()
		if err == io.EOF {
			break
		} else if count >= len(indices) {
			log.Warn().Msg("Recieved more frames than expected")
			break
		}

		frame, err := v.validateFrame(
			ctx,
			reply.GetFrame(),
			indices[count],
			ds,
			verifier,
		)

		if err == nil {
			frames = append(frames, frame)
			validIndices = append(validIndices, indices[count])
		}
		errList = multierr.Append(errList, err)

		count += 1
	}

	if count < len(indices) {
		log.Warn().Msg("Received fewer frames than expected")
	}

	return frames, validIndices, errList
}

func (v *Retriever) validateFrame(
	ctx context.Context,
	frameBytes []byte,
	index uint64,
	ds *graphView.DataStoreRetrieve,
	verifier *rs.KzgVerifier,
) (rs.Frame, error) {
	log := v.Logger.SubloggerId(ctx)
	frame, err := rs.Decode(frameBytes)
	if err != nil {
		log.Error().Msgf("decode frame bytes error %v", err)
		return rs.Frame{}, err
	}

	err = verifier.VerifyFrame(
		(*bn254.G1Point)(bls.DeserializeG1(ds.Header.KzgCommit[:])),
		&frame,
		index,
	)

	if err != nil {
		log.Error().Msgf("decode verify error %v", err)
		return rs.Frame{}, ErrRetrieveReply_DECODEERR
	}

	return frame, nil
}
func (v *Retriever) RetrieveAllFrames(ctx context.Context, dataStoreId uint32) ([]rs.Frame, []uint64, *graphView.DataStoreRetrieve, error) {
	log := v.Logger.SubloggerId(ctx)
	ds, err := v.GraphClient.QueryDataStoreInitBlockNumber(dataStoreId)
	if err != nil {
		return nil, nil, nil, err
	}
	stateView, err := v.GraphClient.GetStateView(ctx, ds.InitBlockNumber)
	if err != nil {
		return nil, nil, nil, err
	}
	advRatioBasisPoints, err := v.ChainClient.GetAdversaryThresholdBasisPoints()
	if err != nil {
		return nil, nil, nil, err
	}
	liveRatioBasisPoints, err := v.ChainClient.GetQuorumThresholdBasisPoints()
	if err != nil {
		return nil, nil, nil, err
	}

	params, err := encoding.GetQuorumParams(liveRatioBasisPoints, advRatioBasisPoints, stateView, 0)
	if err != nil {
		return nil, nil, nil, err
	}

	headerHash := [32]byte{}
	copy(headerHash[:], ds.HeaderHash)
	assignments := encoding.GetOperatorAssignments(params, headerHash)

	// get total Nodes
	totalOperators := len(stateView.Registrants)
	log.Debug().Msgf("Total Operators %v\n", totalOperators)

	verifier, err := v.KzgGroup.NewKzgVerifier(uint64(ds.Header.NumSys), uint64(ds.Header.NumPar), uint64(ds.Header.OrigDataSize))
	if err != nil {
		return nil, nil, nil, err
	}

	update := make(chan RetrievalResult, totalOperators)

	for _, reg := range stateView.Registrants {

		go func(reg *graphView.RegistrantView) {

			indices := assignments[reg.Index].GetIndices()
			log.Debug().Msgf("node reg socket %v\n", reg.Socket)
			frames, indices, errs := v.RetrieveOperatorFrames(ctx, ds, reg.Socket, indices, verifier)

			if err != nil {
				log.Debug().Err(err).Msg("Errors returned during retrieval")
			}

			update <- RetrievalResult{Err: errs, Frames: frames, Indices: indices}
		}(reg)
	}

	rr, ir, err := v.collectFrames(ctx, update, totalOperators)

	if err != nil {
		log.Warn().Err(err).Msg("Errors returned during retrieval")
	}

	return rr, ir, ds, err
}

func (v *Retriever) collectFrames(ctx context.Context, update chan RetrievalResult, numRegistrant int) ([]rs.Frame, []uint64, error) {
	numReply := 0
	rr := make([]rs.Frame, 0)
	var er error
	ir := make([]uint64, 0)

	for {
		select {
		case r := <-update:
			numReply += 1

			rr = append(rr, r.Frames...)
			ir = append(ir, r.Indices...)
			if r.Err != nil {
				er = multierr.Append(er, r.Err)
			}

			// since grpc timeout, we always collect all response
			if numReply == numRegistrant {
				return rr, ir, er
			}
		}
	}
}
