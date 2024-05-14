package dln

import (
	"context"
	"fmt"
	"net/http"

	"net"
	"time"

	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/common/middleware/correlation"
	"github.com/Layr-Labs/datalayr/common/middleware/logger"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"google.golang.org/grpc"
)

const localhost = "0.0.0.0"

// Server contains the configuration and pointers to the other services the DLNs need to use
type Server struct {
	pb.UnimplementedDataDispersalServer

	*Dln
	*Config
	logger *logging.Logger
}

// NewServer creates a new Server struct with the provided parameters
//
// Note: The Server's frame store will be created at config.DbPath+"/frame"
func NewServer(config *Config, dln *Dln, logger *logging.Logger) *Server {
	return &Server{
		Config: config,
		logger: logger,
		Dln:    dln,
	}
}

func (s *Server) Start() {
	s.logger.Trace().Msg("Entering Start function...")
	defer s.logger.Trace().Msg("Exiting Start function...")

	// Serve grpc requests
	//todo: how to handle this err?
	go func(s *Server) {
		addr := fmt.Sprintf("%s:%s", localhost, s.Config.GrpcPort)
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			s.logger.Fatal().Err(err).Msg("Could not start tcp listener")
		}

		opt := grpc.MaxRecvMsgSize(1024 * 1024 * 300)
		gs := grpc.NewServer(
			opt,
			grpc.ChainUnaryInterceptor(
				correlation.UnaryServerInterceptor(),
				logger.UnaryServerInterceptor(*s.logger.Logger),
			),
			grpc.ChainStreamInterceptor(
				correlation.StreamServerInterceptor(),
				logger.StreamServerInterceptor(*s.logger.Logger),
			),
		)
		pb.RegisterDataDispersalServer(gs, s)
		go func() {
			http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
		}()

		s.logger.Info().Str("port", s.Config.GrpcPort).Str("address", listener.Addr().String()).Msg("GRPC Listening")
		if err := gs.Serve(listener); err != nil {
			s.logger.Fatal().Err(err).Msgf("Could not start GRPC server")
		}
	}(s)

	go s.expireLoop()
}

// expireLoop is a loop that is run once per second while the node is running
// it updates the latest expiry time among all active datastores the node is serving
func (s *Server) expireLoop() {
	expireFrom, ok := s.Dln.store.GetLatestBlockTime()
	if !ok {
		expireFrom = 0
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		header, err := s.ChainClient.ChainClient.HeaderByNumber(context.Background(), nil)
		if err != nil {
			s.logger.Error().Err(err).Msg("Could not get latest header")
			continue
		}

		if header.Time <= expireFrom {
			continue
		}

		commits, err := s.GraphClient.GetExpiringDataStores(expireFrom, header.Time)
		if err != nil {
			s.logger.Warn().Err(err).Msg("Could not get expiring data stores")
			continue
		}

		for _, commit := range commits {
			s.logger.Debug().Msgf("Expiring store commitment: %v", hexutil.Encode(commit[:]))
			s.Dln.store.Expire(commit[:])
		}

		// Update expireFrom
		expireFrom = header.Time
		s.store.UpdateLatestBlockTime(expireFrom)
	}

}

// sign returns the serialized signature of the DLN on msgHash
func (s *Server) sign(msgHash []byte) []byte {
	signature := s.Dln.bls.SignMessage(msgHash).Bytes()
	return signature[:]
}

// StoreFrames is called by dispersers on DLNs to store data
func (s *Server) StoreFrames(ctx context.Context, in *pb.StoreFramesRequest) (*pb.StoreFrameReply, error) {
	log := s.logger.SubloggerId(ctx)
	log.Trace().Msg("Entering StoreFrames function...")
	defer log.Trace().Msg("Exiting StoreFrames function...")

	// Measure num requests
	s.Dln.metrics.AccNumRequest.Inc()

	// Measure latency
	start := time.Now()
	defer log.Debug().Msgf("Return at %v.", time.Since(start))

	msgHash := in.GetMsgHash()
	frameBytes := in.GetFrame()

	// Validate DataStore
	err := s.ValidateDataStore(ctx, msgHash, frameBytes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to validate data store")
		return nil, err
	}

	// Store the data in the nodes db
	err = s.Dln.store.InsertCommit(ctx, msgHash[:], frameBytes)
	if err != nil {
		log.Error().Err(err).Msg("Cannot save to Store")
		return nil, pb.ErrLocalCannotSave
	}

	// Sign the msgHash if all validation checks pass
	sig := s.sign(msgHash)

	log.Info().Msg("StoreFrames succeeded")
	s.Dln.metrics.AcceptNewStore(len(frameBytes[0]))

	return &pb.StoreFrameReply{Signature: sig}, nil
}

// RetrieveFrame returns the frames associated with the commit from the request
// and writes the frames to the provided stream
func (s *Server) RetrieveFrame(in *pb.RetrieveFrameRequest, stream pb.DataDispersal_RetrieveFrameServer) error {
	log := s.logger.SubloggerId(stream.Context())
	log.Trace().Msg("Entering RetrieveFrame function...")
	defer log.Trace().Msg("Exiting RetrieveFrame function...")

	framesByte, ok := s.Dln.store.GetFrames(stream.Context(), in.GetCommit())
	if !ok {
		log.Error().Err(ErrKeyNotFoundOrExpired).Msg("Failed to get store frames")
		return ErrKeyNotFoundOrExpired
	} else {
		for _, frameByte := range framesByte {
			reply := &pb.RetrieveFrameReply{
				Frame: frameByte,
			}

			if err := stream.Send(reply); err != nil {
				log.Error().Err(err).Msg("Failed to send reply")
				return err
			}
		}
	}

	log.Info().Msg("RetrieveFrame succeeded")

	return nil
}
