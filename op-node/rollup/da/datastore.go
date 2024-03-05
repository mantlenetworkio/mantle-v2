package da

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Layr-Labs/datalayr/common/graphView"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceRetrieverServer"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/log"
)

const (
	POLLING_INTERVAL     = 1 * time.Second
	MAX_RPC_MESSAGE_SIZE = 1024 * 1024 * 300
)

type MantleDataStoreConfig struct {
	RetrieverSocket          string
	RetrieverTimeout         time.Duration
	GraphProvider            string
	DataStorePollingDuration time.Duration
	MantleDaIndexerSocket    string
	MantleDAIndexerEnable    bool
}

type MantleDataStore struct {
	Ctx           context.Context
	Cfg           *MantleDataStoreConfig
	GraphClient   *graphView.GraphClient
	GraphqlClient *graphql.Client
}

func NewMantleDataStore(ctx context.Context, cfg *MantleDataStoreConfig) *MantleDataStore {
	graphClient := graphView.NewGraphClient(cfg.GraphProvider, nil)
	graphqlClient := graphql.NewClient(graphClient.GetEndpoint(), nil)
	mDatastore := &MantleDataStore{
		Ctx:           ctx,
		Cfg:           cfg,
		GraphClient:   graphClient,
		GraphqlClient: graphqlClient,
	}
	return mDatastore
}

func (mda *MantleDataStore) getDataStoreById(dataStoreId uint32) (*graphView.DataStore, error) {
	var query struct {
		DataStore graphView.DataStoreGql `graphql:"dataStore(id: $storeId)"`
	}
	variables := map[string]interface{}{
		"storeId": graphql.String(strconv.FormatUint(uint64(dataStoreId), 10)),
	}
	err := mda.GraphqlClient.Query(mda.Ctx, &query, variables)
	if err != nil {
		log.Error("Query subgraph fail", "err", err)
		return nil, err
	}
	log.Debug("Query dataStore success",
		"DurationDataStoreId", query.DataStore.DurationDataStoreId,
		"Confirmed", query.DataStore.Confirmed,
		"ConfirmTxHash", query.DataStore.ConfirmTxHash)
	dataStore, err := query.DataStore.Convert()
	if err != nil {
		log.Warn("DataStoreGql convert to DataStore fail", "err", err)
		return nil, err
	}
	return dataStore, nil
}

func (mda *MantleDataStore) getFramesByDataStoreId(dataStoreId uint32) ([]byte, error) {
	conn, err := grpc.Dial(mda.Cfg.RetrieverSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Connect to da retriever fail", "err", err)
		return nil, err
	}
	defer conn.Close()
	client := pb.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(MAX_RPC_MESSAGE_SIZE)
	request := &pb.FramesAndDataRequest{
		DataStoreId: dataStoreId,
	}
	reply, err := client.RetrieveFramesAndData(mda.Ctx, request, opt)
	if err != nil {
		log.Error("Retrieve frames and data fail", "err", err)
		return nil, err
	}
	replyData := reply.GetData()
	log.Debug("Get reply data success", "replyLength", len(replyData))
	return replyData, nil
}

func (mda *MantleDataStore) getFramesFromIndexerByDataStoreId(dataStoreId uint32) ([]byte, error) {
	conn, err := grpc.Dial(mda.Cfg.MantleDaIndexerSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Connect to mantle da index retriever fail", "err", err)
		return nil, err
	}
	defer conn.Close()
	client := pb.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(MAX_RPC_MESSAGE_SIZE)
	request := &pb.FramesAndDataRequest{
		DataStoreId: dataStoreId,
	}
	reply, err := client.RetrieveFramesAndData(mda.Ctx, request, opt)
	if err != nil {
		log.Error("Retrieve frames and data fail", "err", err)
		return nil, err
	}
	replyData := reply.GetData()
	log.Debug("Get reply data from mantle da success", "replyLength", len(replyData))
	return replyData, nil
}

func (mda *MantleDataStore) RetrievalFramesFromDa(dataStoreId uint32) ([]byte, error) {
	pollingTimeout := time.NewTimer(mda.Cfg.DataStorePollingDuration)
	defer pollingTimeout.Stop()
	intervalTicker := time.NewTicker(POLLING_INTERVAL)
	defer intervalTicker.Stop()
	for {
		select {
		case <-intervalTicker.C:
			if dataStoreId <= 0 {
				log.Error("DataStoreId less than zero", "dataStoreId", dataStoreId)
				return nil, errors.New("dataStoreId less than 0")
			}
			dataStore, err := mda.getDataStoreById(dataStoreId)
			if err != nil {
				log.Warn("Get datastore by id fail", "err", err)
				continue
			}
			log.Info("Get dataStore success",
				"DurationDataStoreId", dataStore.DurationDataStoreId,
				"Confirmed", dataStore.Confirmed,
				"ConfirmTxHash", hexutil.Encode(dataStore.ConfirmTxHash[:]))
			if !dataStore.Confirmed {
				log.Warn("This batch is not confirmed")
				continue
			}
			var frames []byte
			if mda.Cfg.MantleDAIndexerEnable { // from mantle da indexer
				log.Info("sync block data from mantle da indexer")
				frames, err = mda.getFramesFromIndexerByDataStoreId(dataStoreId)
				if err != nil {
					log.Warn("Get frames from indexer fail", "err", err)
					continue
				}
			} else { // from mantle da retriever
				log.Info("sync block data from mantle da retriever")
				frames, err = mda.getFramesByDataStoreId(dataStoreId)
				if err != nil {
					log.Warn("Get frames from mantle da retriever fail", "err", err)
					continue
				}
			}
			if frames == nil {
				continue
			}
			return frames, nil
		case <-pollingTimeout.C:
			return nil, errors.New("Get frame ticker exit")
		case err := <-mda.Ctx.Done():
			log.Warn("Retrieval service shutting down", "err", err)
			return nil, errors.New("Retrieval service shutting down")
		}
	}
}
