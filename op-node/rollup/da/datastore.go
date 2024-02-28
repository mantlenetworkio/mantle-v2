package da

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"strconv"
	"time"

	"github.com/shurcooL/graphql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Layr-Labs/datalayr/common/graphView"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceRetrieverServer"

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

func NewMantleDataStore(ctx context.Context, cfg *MantleDataStoreConfig) (*MantleDataStore, error) {
	graphClient := graphView.NewGraphClient(cfg.GraphProvider, nil)
	graphqlClient := graphql.NewClient(graphClient.GetEndpoint(), nil)
	mDatastore := &MantleDataStore{
		Ctx:           ctx,
		Cfg:           cfg,
		GraphClient:   graphClient,
		GraphqlClient: graphqlClient,
	}
	return mDatastore, nil
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

func (mda *MantleDataStore) getLatestDataStoreId() (*graphView.DataStore, error) {
	var query struct {
		DataStores []graphView.DataStoreGql `graphql:"dataStores(first:1,orderBy:storeNumber,orderDirection:desc,where:{confirmed: true})"`
	}
	err := mda.GraphqlClient.Query(mda.Ctx, &query, nil)
	if err != nil {
		log.Error("failed to query lastest dataStore id", "err", err)
		return nil, err
	}
	if len(query.DataStores) == 0 {
		return nil, errors.New("no data store found in this round")
	}
	dataStore, err := query.DataStores[0].Convert()
	if err != nil {
		log.Error("DataStoreGql convert to DataStore fail", "err", err)
		return nil, err
	}
	return dataStore, nil
}

func (mda *MantleDataStore) getFramesByDataStoreId(dataStoreId uint32) ([]byte, error) {

	var conn *grpc.ClientConn
	var err error
	if mda.Cfg.MantleDAIndexerEnable {
		log.Info("sync block data from mantle da indexer")
		conn, err = grpc.Dial(mda.Cfg.MantleDaIndexerSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Error("Connect to mantle da index retriever fail", "err", err)
			return nil, err
		}
	} else {
		log.Info("sync block data from mantle da retriever")
		conn, err = grpc.Dial(mda.Cfg.RetrieverSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Error("Connect to da retriever fail", "err", err)
			return nil, err
		}
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
	log.Debug("Get reply data success", "replyLength", len(reply.GetData()))
	return reply.GetData(), nil
}

func (mda *MantleDataStore) RetrievalFramesFromDa(dataStoreId uint32) ([]byte, error) {
	dataStore, err := mda.getDataStoreById(dataStoreId)
	if err != nil {
		log.Error("get datastore by id fail", "err", err)
		return nil, err
	}

	log.Info("get dataStore success",
		"durationDataStoreId", dataStore.DurationDataStoreId,
		"confirmed", dataStore.Confirmed,
		"confirmTxHash", hexutil.Encode(dataStore.ConfirmTxHash[:]))
	lastDataStore, err := mda.getLatestDataStoreId()
	if err != nil {
		log.Error("get lastest datastore fail", "err", err)
		return nil, err
	}

	if !dataStore.Confirmed && dataStoreId < lastDataStore.StoreNumber {
		log.Warn("this batch is not confirmed in mantle da,but new batch is confirmed,data corruption exists,need to skip this dataStoreId ", "dataStore id", dataStoreId, "latest dataStore id", lastDataStore.StoreNumber)
		return nil, nil
	} else if !dataStore.Confirmed {
		log.Warn("this batch is not confirmed in mantle da ", "dataStore id", dataStoreId)
		return nil, fmt.Errorf("this batch is not confirmed in mantle da,datastore id %d", dataStoreId)
	}
	var frames []byte
	frames, err = mda.getFramesByDataStoreId(dataStoreId)
	if err != nil {
		log.Error("Get frames from indexer fail", "err", err)
		return nil, err
	}
	if frames == nil {
		log.Error("frames is nil ,waiting da indexer syncing")
		return nil, fmt.Errorf("frames is nil,maybe da indexer is syncing,need to try again,dataStore id %d", dataStoreId)
	}
	return frames, nil
}
