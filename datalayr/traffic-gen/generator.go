package trafficGen

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GetRunner(version string) func(cliCtx *cli.Context) error {
	return func(ctx *cli.Context) error {

		config := NewConfig(ctx)

		generators, err := NewTrafficGeneator(config)
		if err != nil {
			return err
		}

		generators.Run()

		return nil
	}
}

type TrafficGeneator struct {
	metrics *Metrics
	Logger  *logging.Logger
	config  Config
}

func NewTrafficGeneator(config Config) (*TrafficGeneator, error) {
	target := fmt.Sprintf("%v:%v", config.Hostname, config.GrpcPort)
	logger, err := logging.GetLogger(config.LoggingConfig)
	metrics := NewMetrics(config.Number, target, logger)
	if err != nil {
		return nil, err
	}

	return &TrafficGeneator{
		metrics: metrics,
		Logger:  logger,
		config:  config,
	}, nil
}

func (g *TrafficGeneator) Run() error {
	for i := 0; i < int(g.config.Number); i++ {
		go func(id int) {
			g.startGenerator(id)
		}(i)
	}
	return nil
}

// Run infinitely
func (g *TrafficGeneator) startGenerator(id int) {
	// log every 60 sec
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case _ = <-ticker.C:
				g.metrics.Log()
			}
		}
	}()

	data := make([]byte, g.config.DataSize)
	for {
		// send query
		storeId, err := sendGRPC(g.config, data, g.Logger)

		// metrics
		g.metrics.Update(id, storeId, g.config.DataSize, err)

		// wait for some time
		idle(g.config)
	}
}

func sendGRPC(config Config, data []byte, logger *logging.Logger) (uint32, error) {
	socket := fmt.Sprintf("%v:%v", config.Hostname, config.GrpcPort)
	conn, err := grpc.Dial(socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error().Err(err).Msgf("---- Err. Traffic Generator Cannot connect to %v. %v", socket)
		return 0, err
	}
	defer conn.Close()

	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.Timeout))
	defer cancel()

	//construct request to encode data, just ignore block number
	request := &pb.EncodeStoreRequest{
		Duration: config.StoreDuration,
		Data:     data,
	}

	//request disperser to encode data
	opt := grpc.MaxCallSendMsgSize(1024 * 1024 * 300)
	reply, err := c.EncodeAndDisperseStore(ctx, request, opt)
	if err != nil {
		logger.Error().Err(err).Msgf("msgHash Failed")
		return 0, err
	}

	return reply.StoreId, nil
}

func idle(config Config) {
	dur := int(rand.NormFloat64()*float64(config.IdlePeriodStd)) + int(config.IdlePeriod)
	if dur < 0 {
		dur = 0
	}

	time.Sleep(time.Duration(dur) * time.Millisecond)
}
