package integration

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/Layr-Labs/datalayr/common/interfaces/interfaceRetrieverServer"
	"github.com/Layr-Labs/datalayr/integration/deploy"
	interfaceSequencer "github.com/Layr-Labs/datalayr/middleware/rollup-example/sequencer/proto"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func grpcDialContext(ctx context.Context, timeout time.Duration, socket string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, socket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

var _ = Describe("Integration", func() {

	var (
		ctx = context.Background()
	)

	It("test end to end scenario", func() {
		// Setup disperser client.
		if disperserSocket == "" {
			port := testConfig.Dispersers[0].DL_DISPERSER_GRPC_PORT
			disperserSocket = fmt.Sprintf("localhost:%v", port)
		}

		disConn, err := grpcDialContext(ctx, 5*time.Second, disperserSocket)
		Expect(err).To(BeNil())
		defer disConn.Close()
		disClient := interfaceDL.NewDataDispersalClient(disConn)

		// Setup retriever client.
		if retrieverSocket == "" {
			port := testConfig.Retrievers[0].DL_RETRIEVER_GRPC_PORT
			retrieverSocket = fmt.Sprintf("localhost:%v", port)
		}

		retConn, err := grpcDialContext(ctx, 5*time.Second, retrieverSocket)
		Expect(err).To(BeNil())
		defer retConn.Close()
		retClient := interfaceRetrieverServer.NewDataRetrievalClient(retConn)

		data := strings.Repeat("0", 1000)
		dataByte := []byte(data)
		duration := 1

		Eventually(func() error {
			disRes, err := disClient.EncodeAndDisperseStore(ctx, &interfaceDL.EncodeStoreRequest{
				Duration: uint64(duration),
				Data:     dataByte,
			})
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			if disRes.GetStore() == nil {
				return errors.New("store is nil")
			}

			retRes, err := retClient.RetrieveFramesAndData(ctx, &interfaceRetrieverServer.FramesAndDataRequest{
				DataStoreId: disRes.StoreId,
			})
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			if !reflect.DeepEqual(retRes.Data, dataByte) {
				return errors.New("Reconstructed data is not equal to the original data.")
			}

			return nil
		}, "5000ms", "500ms").Should(BeNil())
	})
})

var _ = Describe("Rollup", func() {

	var (
		ctx        = context.Background()
		fraudBlock = "0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000002d2d5468697320697320612062616420737472696e672e204e6f626f64792073617973207468697320737472696e672e2d2d00000000"
	)
	const (
		COMMITTED_STATUS = "0x0000000000000000000000000000000000000000000000000000000000000001"
		SLASHED_STATUS   = "0x0000000000000000000000000000000000000000000000000000000000000002"
	)

	It("slashes the sequencer when it posts a fraud block.", func() {
		if !testConfig.IsRollupDeployed() || testConfig.Services.Counts.NumRollupSeq < 1 {
			Skip("The rollup experiment is not set up.")
		}

		// Setup sequencer client.
		port := testConfig.Sequencers[0].SEQUENCER_GRPC_PORT
		sequencerSocket := fmt.Sprintf("localhost:%v", port)
		seqConn, err := grpcDialContext(ctx, 5*time.Second, sequencerSocket)
		Expect(err).To(BeNil())
		defer seqConn.Close()
		seqClient := interfaceSequencer.NewSequencerClient(seqConn)

		// Check that the sequencer's status is committed.
		Eventually(func() string {
			sequencerStatus := deploy.CallContract(
				testConfig.Sequencers[0].SEQUENCER_ROLLUP_ADDRESS,
				"sequencerStatus()",
				testConfig.Deployers[0].Rpc,
			)

			return sequencerStatus
		}, "5000ms", "500ms").Should(Equal(COMMITTED_STATUS))

		// Posting a fraud block should succeed.
		Eventually(func() error {
			_, err := seqClient.PostBlock(ctx, &interfaceSequencer.PostBlockRequest{
				Data: fraudBlock,
			})
			if err != nil {
				fmt.Println(err.Error())
				return err
			}

			return nil
		}, "5000ms", "500ms").Should(BeNil())

		// Check that the sequencer is eventually slashed.
		Eventually(func() string {
			sequencerStatus := deploy.CallContract(
				testConfig.Sequencers[0].SEQUENCER_ROLLUP_ADDRESS,
				"sequencerStatus()",
				testConfig.Deployers[0].Rpc,
			)

			return sequencerStatus
		}, "5000ms", "500ms").Should(Equal(SLASHED_STATUS))
	})
})
