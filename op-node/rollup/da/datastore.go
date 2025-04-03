package da

import (
	"context"
	"fmt"
	"time"

	"crypto/tls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	mdar "github.com/ethereum-optimism/optimism/op-node/rollup/da/interfaceRetrieverServer"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"

	"github.com/ethereum/go-ethereum/log"
)

const (
	POLLING_INTERVAL     = 1 * time.Second
	MAX_RPC_MESSAGE_SIZE = 1024 * 1024 * 300
)

type EigenDADataStore struct {
	daClient              eigenda.IEigenDA
	MantleDaIndexerSocket string
	MantleDAIndexerEnable bool
	Ctx                   context.Context
}

func NewEigenDADataStore(ctx context.Context, log log.Logger, daCfg *Config, m eigenda.Metrics) *EigenDADataStore {
	var daClient eigenda.IEigenDA
	if daCfg != nil {
		daClient = eigenda.NewEigenDAClient(daCfg.Config, log, m)
	}
	return &EigenDADataStore{
		daClient: daClient,
		Ctx:      ctx,
	}
}

func (da *EigenDADataStore) IsDaIndexer() bool {
	return da.MantleDAIndexerEnable
}

func (da *EigenDADataStore) RetrieveBlob(batchHeaderHash []byte, blobIndex uint32, commitment []byte) ([]byte, error) {
	if len(commitment) == 0 {
		return da.daClient.RetrieveBlob(da.Ctx, batchHeaderHash, blobIndex)
	}
	return da.daClient.RetrieveBlobWithCommitment(da.Ctx, commitment)
}

func (da *EigenDADataStore) RetrieveBlobWithCommitment(commitment []byte) ([]byte, error) {
	return da.daClient.RetrieveBlobWithCommitment(da.Ctx, commitment)
}

func (da *EigenDADataStore) RetrievalFramesFromDaIndexer(txHash string) ([]byte, error) {
	log.Info("sync block data from mantle da retriever")
	conn, err := grpc.Dial(da.MantleDaIndexerSocket, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	if err != nil {
		log.Error("Connect to da retriever fail", "err", err)
		return nil, err
	}
	defer conn.Close()

	client := mdar.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(MAX_RPC_MESSAGE_SIZE)
	request := &mdar.FramesAndDataRequest{
		Value: &mdar.FramesAndDataRequest_DataConfirmHash{
			DataConfirmHash: txHash,
		},
	}
	reply, err := client.RetrieveFramesAndData(da.Ctx, request, opt)
	if err != nil {
		log.Error("Retrieve frames and data fail", "err", err)
		return nil, err
	}

	if reply.Data == nil {
		log.Error("frames is nil ,waiting da indexer syncing")
		return nil, fmt.Errorf("frames is nil,maybe da indexer is syncing,need to try again,txHash %s", txHash)
	}

	replyData := reply.Data
	log.Debug("Get reply data success", "replyLength", len(replyData))
	return replyData, nil
}
