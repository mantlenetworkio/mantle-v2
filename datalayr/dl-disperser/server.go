package disperser

import (
	"context"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"reflect"
	"time"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/common/middleware/correlation"
	"github.com/Layr-Labs/datalayr/common/middleware/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const localhost = "0.0.0.0"

type Server struct {
	pb.UnimplementedDataDispersalServer

	*Disperser
	*Config
	logger *logging.Logger
}

func NewServer(config *Config, disperser *Disperser, logger *logging.Logger) *Server {
	return &Server{
		Config:    config,
		Disperser: disperser,
		logger:    logger,
	}
}

func (s *Server) Start() {
	s.logger.Trace().Msg("Entering Start function...")
	defer s.logger.Trace().Msg("Exiting Start function...")
	// Serve grpc requests
	//todo: resolve fatal
	s.CodedDataCache.StartExpireLoop()
	go func(s *Server) {
		addr := fmt.Sprintf("%s:%s", localhost, s.Config.GrpcPort)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			s.logger.Fatal().Err(err).Msg("Could not start tcp listener. ")
		}

		opt := grpc.MaxRecvMsgSize(1024 * 1024 * 300)
		gs := grpc.NewServer(
			opt,
			grpc.ChainUnaryInterceptor(
				correlation.UnaryServerInterceptor(),
				logger.UnaryServerInterceptor(*s.logger.Logger),
			),
		)
		reflection.Register(gs)
		pb.RegisterDataDispersalServer(gs, s)
		go func() {
			http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		}()

		s.logger.Info().Str("port", s.Config.GrpcPort).Str("address", listener.Addr().String()).Msg("GRPC Listening")
		if err := gs.Serve(listener); err != nil {
			s.logger.Fatal().Err(err).Msgf("Cannot start GRPC server")
		}
	}(s)

}

func (s *Server) EncodeStore(ctx context.Context, in *pb.EncodeStoreRequest) (*pb.EncodeStoreReply, error) {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering EncodeStore function...")
	defer log.Trace().Msg("Exiting EncodeStore function...")

	// s.metrics.Encode.WithLabelValues("total", "number").Inc()
	//get ratio basis points from contract
	advRatioBasisPoints, err := s.ChainClient.GetAdversaryThresholdBasisPoints() // 4000
	if err != nil {
		return nil, err
	}

	liveRatioBasisPoints, err := s.ChainClient.GetQuorumThresholdBasisPoints() // 9000
	if err != nil {
		return nil, err
	}

	store, _, err := s.Disperser.CreateStore(
		ctx,
		NewStoreRequest(
			in.BlockNumber,
			advRatioBasisPoints,
			liveRatioBasisPoints,
			in.Duration,
			in.Data,
		),
	)
	if err != nil {
		log.Error().Err(err).Msg("CreateStore")
		return &pb.EncodeStoreReply{}, status.Errorf(codes.Internal, "Failed to create a data store")
	}

	err = s.CodedDataCache.Add(store)
	if err != nil {
		log.Error().Err(err).Msg("Failed to add store to cache")
		return &pb.EncodeStoreReply{}, status.Errorf(codes.Internal, "Failed to add encoded into local cache. Overflow")
	}

	// s.metrics.Encode.WithLabelValues("success", "number").Inc()
	return &pb.EncodeStoreReply{
		Store: NewStoreParams(store),
	}, nil

}

func (s *Server) DisperseStore(ctx context.Context, in *pb.DisperseStoreRequest) (*pb.DisperseStoreReply, error) {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering DisperseStore function...")
	defer log.Trace().Msg("Exiting DisperseStore function...")

	// s.metrics.Disperse.WithLabelValues("total", "number").Inc()
	// return nil, status.Errorf(codes.NotFound, "An existing store was not found")
	HeaderHashArray, err := make32ByteArray(in.HeaderHash)
	if err != nil {
		return &pb.DisperseStoreReply{}, status.Errorf(codes.InvalidArgument, "HeaderHash is not 32 bytes")
	}

	store, err := s.CodedDataCache.Get(HeaderHashArray)
	if err != nil {
		return &pb.DisperseStoreReply{}, status.Errorf(codes.NotFound, "An existing store was not found")
	}

	if !reflect.DeepEqual(in.HeaderHash, store.HeaderHash[:]) {
		return &pb.DisperseStoreReply{}, status.Errorf(codes.OutOfRange, "The header hash of existing store did not match")
	}

	stateView, err := s.Disperser.GraphClient.GetStateView(ctx, store.ReferenceBlockNumber)
	if err != nil {
		return &pb.DisperseStoreReply{}, status.Errorf(codes.Internal, "Could not get state info")
	}

	copy(store.MsgHash[:], in.MessageHash)

	// Disperse to dlns
	aggResult, err := s.Aggregator.Aggregate(ctx, &store, stateView)
	if err != nil {
		s.Logger.Error().Err(err).Msg("Could not aggregate signatures")
		return &pb.DisperseStoreReply{}, status.Errorf(codes.DeadlineExceeded, "Could not aggregate signatures")
	}

	nonSignerPubKeys := make([][]byte, 0)
	for _, pubKey := range aggResult.NonSignerPubkeys {
		pubKeyBytes := pubKey.Bytes()
		nonSignerPubKeys = append(nonSignerPubKeys, pubKeyBytes[:])
	}

	aggSigBytes := bls.SerializeG1(aggResult.AggSig)
	storedAggPubKeyBytesG1 := bls.SerializeG1(aggResult.StoredAggPubkeyG1)
	usedAggPubKeyBytesG2 := bls.SerializeG2(aggResult.UsedAggPubkeyG2)

	s.CodedDataCache.Delete(store.HeaderHash)
	// s.metrics.Disperse.WithLabelValues("success", "number").Inc()
	return &pb.DisperseStoreReply{
		Sigs: NewAggregateSignature(
			aggSigBytes[:],
			storedAggPubKeyBytesG1[:],
			usedAggPubKeyBytesG2[:],
			nonSignerPubKeys,
		),
		ApkIndex:        stateView.TotalOperator.Index,
		TotalStakeIndex: uint64(stateView.TotalStake.Index),
	}, nil

}

func (s *Server) EncodeAndDisperseStore(ctx context.Context, in *pb.EncodeStoreRequest) (*pb.EncodeAndDisperseStoreReply, error) {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering EncodeAndDisperseStore function...")
	defer log.Trace().Msg("Exiting EncodeAndDisperseStore function...")
	start := time.Now()

	//get ratio basis points from contract
	reply, err := func() (*pb.EncodeAndDisperseStoreReply, error) {
		advRatioBasisPoints, err := s.ChainClient.GetAdversaryThresholdBasisPoints()
		if err != nil {
			return nil, err
		}

		liveRatioBasisPoints, err := s.ChainClient.GetQuorumThresholdBasisPoints()
		if err != nil {
			return nil, err
		}

		store, aggResult, err := s.Disperser.Disperse(
			ctx,
			NewStoreRequest(
				in.BlockNumber,
				advRatioBasisPoints,
				liveRatioBasisPoints,
				in.Duration,
				in.Data,
			),
		)

		if err != nil {
			return nil, err
		}

		nonSignerPubKeys := make([][]byte, 0)
		for _, pubKey := range aggResult.NonSignerPubkeys {
			pubKeyBytes := pubKey.Bytes()
			nonSignerPubKeys = append(nonSignerPubKeys, pubKeyBytes[:])
		}

		aggSigBytes := bls.SerializeG1(aggResult.AggSig)
		storedAggPubKeyBytesG1 := bls.SerializeG1(aggResult.StoredAggPubkeyG1)
		usedAggPubKeyBytesG2 := bls.SerializeG2(aggResult.UsedAggPubkeyG2)

		// Create store

		return &pb.EncodeAndDisperseStoreReply{
			Store: NewStoreParams(store),
			Sigs: NewAggregateSignature(
				aggSigBytes[:],
				storedAggPubKeyBytesG1[:],
				usedAggPubKeyBytesG2[:],
				nonSignerPubKeys,
			),
			MsgHash: store.MsgHash[:],
			StoreId: store.StoreId,
		}, nil
	}()

	t := time.Now()
	lat := t.Sub(start)

	if err == nil {
		log.Info().Msgf("EncodeAndDisperseStore succeeded in %v", lat)

		s.metrics.RecordNewRequest(true, int(reply.Store.OrigDataSize), lat)
	} else {
		log.Error().Err(err).Msg("EncodeAndDisperseStore failed")
		s.metrics.RecordNewRequest(false, 0, lat)
	}

	return reply, err
}

func NewStoreParams(store *Store) *pb.StoreParams {
	return &pb.StoreParams{
		ReferenceBlockNumber: store.ReferenceBlockNumber,
		TotalOperatorsIndex:  store.TotalOperatorsIndex,
		OrigDataSize:         store.Header.OrigDataSize,
		Quorum:               uint32(store.MantleFirstQuorumThreshold.Uint64()),
		Duration:             uint32(store.Duration),
		NumSys:               store.Header.NumSys,
		NumPar:               store.Header.NumPar,
		KzgCommit:            store.Header.KzgCommit[:],
		LowDegreeProof:       store.Header.LowDegreeProof[:],
		Degree:               store.Header.Degree,
		HeaderHash:           store.HeaderHash[:],
		Fee:                  store.Fee.Bytes(),
	}
}

func NewAggregateSignature(
	aggSigBytes []byte,
	storedAggPubKeyBytesG1 []byte,
	usedAggPubKeyBytesG2 []byte,
	nonSignerPubKeys [][]byte,
) *pb.AggregateSignature {
	return &pb.AggregateSignature{
		AggSig:            aggSigBytes[:],
		StoredAggPubkeyG1: storedAggPubKeyBytesG1[:],
		UsedAggPubkeyG2:   usedAggPubKeyBytesG2[:],
		NonSignerPubkeys:  nonSignerPubKeys,
	}
}

func NewStoreRequest(
	blockNumber uint32,
	advRatioBasisPoints, liveRatioBasisPoints *big.Int,
	duration uint64,
	data []byte,
) StoreRequest {
	return StoreRequest{
		BlockNumber:          blockNumber,
		AdvRatioBasisPoints:  advRatioBasisPoints,
		LiveRatioBasisPoints: liveRatioBasisPoints,
		Duration:             duration,
		Data:                 data,
	}
}
