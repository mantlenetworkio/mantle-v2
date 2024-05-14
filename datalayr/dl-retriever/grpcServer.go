package retriever

import (
	"context"
	"fmt"
	"net"

	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceRetrieverServer"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/Layr-Labs/datalayr/common/middleware/correlation"
	"github.com/Layr-Labs/datalayr/common/middleware/logger"
	rs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const localhost = "0.0.0.0"

type Server struct {
	*Config

	pb.UnimplementedDataRetrievalServer

	logger    *logging.Logger
	Retriever *Retriever
}

func NewServer(config *Config, retriever *Retriever, logger *logging.Logger) *Server {
	return &Server{
		Config:    config,
		logger:    logger,
		Retriever: retriever,
	}
}

func (s *Server) Start() {
	//todo: resolve fatal
	go func(s *Server) {
		addr := fmt.Sprintf("%s:%s", localhost, s.GrpcPort)
		s.logger.Info().Msgf("addr %v", addr)

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
		pb.RegisterDataRetrievalServer(gs, s)

		s.logger.Info().Str("port", s.GrpcPort).Str("address", listener.Addr().String()).Msg("GRPC Listening")
		if err := gs.Serve(listener); err != nil {
			s.logger.Fatal().Err(err).Msgf("Could not GRPC server")
		}
	}(s)
}

func (s *Server) RetrieveFramesAndData(ctx context.Context, in *pb.FramesAndDataRequest) (*pb.FramesAndDataReply, error) {
	log := s.logger.SubloggerId(ctx)
	dataStoreId := in.GetDataStoreId()

	frames, indices, ds, err := s.Retriever.RetrieveAllFrames(ctx, uint32(dataStoreId))
	if err != nil {
		log.Warn().Err(err).Msg("could not retrieve frames")
		return nil, err
	}
	log.Debug().Msgf("ds.Header.NumSys %v ds.Header.NumPar %v num frame %v", ds.Header.NumSys, ds.Header.NumPar, len(frames))
	data, recodeFrames, err := s.Retriever.RecoverFrames(ctx, frames, indices, ds.Header)
	if err != nil {
		return nil, err
	}

	framesBytes, err := transformLibrary(recodeFrames)
	if err != nil {
		return nil, err
	}

	return &pb.FramesAndDataReply{
		Data:   data,
		Frames: framesBytes,
	}, nil
}

// Todo remove it after clean crypto library
func transformLibrary(frames []rs.Frame) ([][]byte, error) {
	framesBytes := make([][]byte, len(frames))
	for i, f := range frames {

		fb, err := f.Encode()
		if err != nil {
			return nil, err
		}
		framesBytes[i] = fb
	}

	return framesBytes, nil
}
