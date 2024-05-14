package graphView

import (
	"context"
	"fmt"
	"strconv"

	"github.com/Layr-Labs/datalayr/common/header"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shurcooL/graphql"
)

type DataStoreRetrieveGql struct {
	InitBlockNumber graphql.String
	DataCommitment  graphql.String
	MsgHash         graphql.String
	Header          graphql.String
}

type DataStoreRetrieve struct {
	InitBlockNumber uint32
	HeaderHash      []byte
	MsgHash         []byte
	Header          header.DataStoreHeader
}

func (g *GraphClient) QueryDataStoreInitBlockNumber(id uint32) (*DataStoreRetrieve, error) {
	var query struct {
		DataStore DataStoreRetrieveGql `graphql:"dataStore(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.String(fmt.Sprintf("%v", id)),
	}

	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		g.Logger.Error().Err(err).Msg("QueryInitDataStore error")
		return nil, err
	}

	blockNumber, err := strconv.ParseUint(
		string(query.DataStore.InitBlockNumber), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	dataCommitment, err := hexutil.Decode(string(query.DataStore.DataCommitment))
	if err != nil {
		return nil, err
	}

	msgHash, err := hexutil.Decode(string(query.DataStore.MsgHash))
	if err != nil {
		return nil, err
	}

	headerBytes, err := hexutil.Decode(string(query.DataStore.Header))
	if err != nil {
		return nil, err
	}

	header, err := header.DecodeDataStoreHeader(headerBytes)
	if err != nil {
		return nil, err
	}

	ret := DataStoreRetrieve{
		InitBlockNumber: uint32(blockNumber),
		HeaderHash:      dataCommitment,
		MsgHash:         msgHash,
		Header:          header,
	}

	return &ret, nil
}

type DataStoreInit_DataCommitmentGql struct {
	DataCommitment graphql.String
}

func (d *DataStoreInit_DataCommitmentGql) Convert() ([32]byte, error) {

	dataCommitment, err := hexutil.Decode(string(d.DataCommitment))
	if err != nil {
		return [32]byte{}, err
	}
	var dataArray [32]byte
	copy(dataArray[:], dataCommitment)
	return dataArray, nil
}

func (g *GraphClient) GetExpiringDataStores(fromTime uint64, toTime uint64) ([][32]byte, error) {
	var query struct {
		InitDs []DataStoreInit_DataCommitmentGql `graphql:"dataStores(where:{expireTime_gte: $fromTime,expireTime_lt:$toTime})"`
	}
	variables := map[string]interface{}{
		"fromTime": graphql.String(fmt.Sprint(fromTime)),
		"toTime":   graphql.String(fmt.Sprint(toTime)),
	}
	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		g.Logger.Error().Err(err).Msg("GetExpiringDataStores error")
		return nil, err
	}
	commits := make([][32]byte, 0)
	for _, result := range query.InitDs {
		commit, err := result.Convert()
		if err != nil {
			return nil, err
		}
		commits = append(commits, commit)
	}
	return commits, nil
}
