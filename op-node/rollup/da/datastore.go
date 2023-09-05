package da

import (
	"context"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/shurcooL/graphql"
	"google.golang.org/grpc"

	"github.com/ethereum/go-ethereum/log"

	"github.com/Layr-Labs/datalayr/common/graphView"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceRetrieverServer"
)

type MantleDataStoreConfig struct {
	RetrieverSocket  string
	RetrieverTimeout time.Duration
	GraphProvider    string
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
	dataStore, err := query.DataStore.Convert()
	if err != nil {
		log.Error("DataStoreGql convert to DataStore fail", "err", err)
		return nil, err
	}
	return dataStore, nil
}

func (mda *MantleDataStore) getFramesByDataStoreId(dataStoreId uint32) ([]byte, error) {
	conn, err := grpc.Dial(mda.Cfg.RetrieverSocket, grpc.WithInsecure())
	if err != nil {
		log.Error("Connect to da retriever fail", "err", err)
		return nil, err
	}
	defer conn.Close()
	client := pb.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(1024 * 1024 * 300)
	request := &pb.FramesAndDataRequest{
		DataStoreId: dataStoreId,
	}
	reply, err := client.RetrieveFramesAndData(mda.Ctx, request, opt)
	if err != nil {
		log.Error("Retrieve frames and data fail", "err", err)
		return nil, err
	}
	return reply.GetData(), nil
}

func (mda *MantleDataStore) RetrievalFramesFromDa(dataStoreId uint32) ([]byte, error) {
	if dataStoreId <= 0 {
		return nil, errors.New("dataStoreId less than 0")
	}
	dataStore, err := mda.getDataStoreById(dataStoreId)
	if err != nil {
		log.Error("Get datastore fail", "err", err)
	}
	if !dataStore.Confirmed {
		log.Debug("This batch is not confirmed")
		return nil, errors.New("This batch is not confirmed")
	}
	frames, err := mda.getFramesByDataStoreId(dataStoreId)
	if err != nil {
		log.Error("Get frames fail", "err", err)
		return nil, err
	}
	return frames, nil
}
