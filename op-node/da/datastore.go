package da

import (
	"context"
	"github.com/shurcooL/graphql"
	"google.golang.org/grpc"

	"github.com/Layr-Labs/datalayr/common/graphView"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceRetrieverServer"

	"github.com/ethereum/go-ethereum/log"
)

type MantleDataStoreConfig struct {
	RetrieverSocket               string
	RetrieverTimeout              string
	DatalayrServiceMangerContract string
	GraphProvider                 string
}

type MantleDataStore struct {
	Ctx           context.Context
	Cfg           *MantleDataStoreConfig
	GraphClient   *graphView.GraphClient
	GraphqlClient *graphql.Client
	cancel        func()
}

func (mda *MantleDataStore) GetDataStoreById(dataStoreId string) (*graphView.DataStore, error) {
	var query struct {
		dDataStore graphView.DataStoreGql `graphql:"dataStore(id: $storeId)"`
	}
	variables := map[string]interface{}{
		"storeId": graphql.String(dataStoreId),
	}
	err := mda.GraphqlClient.Query(mda.Ctx, &query, variables)
	if err != nil {
		log.Error("query subgraph fail", "err", err)
		return nil, err
	}
	return query.DataStore, nil
}

func (mda *MantleDataStore) GetTransactionListByStoreNumber(dataStoreId string) ([]byte, error) {
	conn, err := grpc.Dial(s.Cfg.RetrieverSocket, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	client := pb.NewDataRetrievalClient(conn)
	opt := grpc.MaxCallRecvMsgSize(1024 * 1024 * 300)
	request := &pb.FramesAndDataRequest{
		DataStoreId: dataStoreId,
	}
	reply, err := client.RetrieveFramesAndData(s.Ctx, request, opt)
	if err != nil {
		log.Error("retrieve frames and data fail", "err", err)
		return nil, err
	}
	return reply.GetData(), nil
}
