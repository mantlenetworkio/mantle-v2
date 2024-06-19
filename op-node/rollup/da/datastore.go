package da

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/shurcooL/graphql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	mdar "github.com/ethereum-optimism/optimism/op-node/rollup/da/interfaceRetrieverServer"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"

	"github.com/Layr-Labs/datalayr/common/graphView"
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
	log.Info("sync block data from mantle da retriever")
	conn, err := grpc.Dial(mda.Cfg.RetrieverSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Connect to da retriever fail", "err", err)
		return nil, err
	}
	defer conn.Close()

	client := mdar.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(MAX_RPC_MESSAGE_SIZE)
	request := &mdar.FramesAndDataRequest{
		Value: &mdar.FramesAndDataRequest_DataStoreId{
			DataStoreId: dataStoreId,
		},
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
	log.Info("get last dataStore success", "dataStoreId", dataStore.StoreNumber)

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

func (mda *MantleDataStore) RetrievalFramesFromDaIndexer(dataStoreId uint32) ([]byte, error) {
	log.Info("sync block data from mantle da indexer")
	conn, err := grpc.Dial(mda.Cfg.MantleDaIndexerSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Connect to mantle da index retriever fail", "err", err)
		return nil, err
	}
	defer conn.Close()

	client := mdar.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(MAX_RPC_MESSAGE_SIZE)
	request := &mdar.FramesAndDataRequest{
		Value: &mdar.FramesAndDataRequest_DataStoreId{
			DataStoreId: dataStoreId,
		},
	}
	reply, err := client.RetrieveFramesAndData(mda.Ctx, request, opt)
	if err != nil {
		log.Error("Retrieve frames and data fail", "err", err)
		return nil, err
	}
	if reply.GetData() == nil {
		lastDataStoreId, err := client.RetrieveLastConfirmDataStoreId(mda.Ctx, &mdar.LastDataStoreIdRequest{}, opt)
		if err != nil {
			log.Error("Retrieve last confirmed data store id fail", "err", err)
			return nil, err
		}
		if dataStoreId < lastDataStoreId.GetDataStoreId() {
			log.Warn("this batch is not confirmed in mantle da,but new batch is confirmed,data corruption exists,need to skip this dataStoreId ", "dataStore id", dataStoreId, "latest dataStore id", lastDataStoreId.GetDataStoreId())
			return nil, nil
		}
		log.Error("frames is nil ,waiting da indexer syncing")
		return nil, fmt.Errorf("frames is nil,maybe da indexer is syncing,need to try again,dataStore id %d", dataStoreId)
	}
	replyData := reply.GetData()
	log.Debug("Get reply data from mantle da success", "replyLength", len(replyData))
	return replyData, nil
}

func (mda *MantleDataStore) IsDaIndexer() bool {
	return mda.Cfg.MantleDAIndexerEnable
}

type EigenDADataStore struct {
	daClient eigenda.IEigenDA
	Cfg      *MantleDataStoreConfig
	Ctx      context.Context
}

func NewEigenDADataStore(ctx context.Context, log log.Logger, daCfg *eigenda.Config, cfg *MantleDataStoreConfig) *EigenDADataStore {
	var daClient eigenda.IEigenDA
	if daCfg != nil {
		daClient = &eigenda.EigenDA{
			Log:    log,
			Config: *daCfg,
		}
	}
	return &EigenDADataStore{
		daClient: daClient,
		Cfg:      cfg,
		Ctx:      ctx,
	}
}

func (da *EigenDADataStore) IsDaIndexer() bool {
	return da.Cfg.MantleDAIndexerEnable
}

func (da *EigenDADataStore) RetrieveBlob(BatchHeaderHash []byte, BlobIndex uint32) ([]byte, error) {
	return da.daClient.RetrieveBlob(da.Ctx, BatchHeaderHash, BlobIndex)
}

func (da *EigenDADataStore) RetrievalFramesFromDaIndexer(txHash string) ([]byte, error) {
	log.Info("sync block data from mantle da retriever")
	conn, err := grpc.Dial(da.Cfg.MantleDaIndexerSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
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
