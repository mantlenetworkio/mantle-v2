package graphView

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shurcooL/graphql"
)

// some field are not strictly needed for disperse logics
// but provided for event in contracts. Need 1-to-1 relation
type DataStoreInitGql struct {
	Id                   graphql.String
	StoreNumber          graphql.String
	DurationDataStoreId  graphql.String
	Index                graphql.String
	DataCommitment       graphql.String
	MsgHash              graphql.String
	ReferenceBlockNumber graphql.String
	InitTime             graphql.String
	ExpireTime           graphql.String
	Duration             graphql.Int
	StorePeriodLength    graphql.String
	Fee                  graphql.String
	Header               graphql.String
	Confirmer            graphql.String
	InitTxHash           graphql.String
	InitGasUsed          graphql.String
	InitBlockNumber      graphql.String
}

type DataStoreInit struct {
	StoreNumber          uint32
	DurationDataStoreId  uint32
	Index                uint32
	DataCommitment       [32]byte
	MsgHash              [32]byte
	ReferenceBlockNumber uint32
	InitTime             uint32
	ExpireTime           uint32
	Duration             uint8
	StorePeriodLength    uint32
	Fee                  *big.Int
	Header               []byte
	Confirmer            string
	InitTxHash           [32]byte
	InitGasUsed          uint64
	InitBlockNumber      *big.Int
}

func (g *GraphClient) QueryInitDataStoreByMsgHash(msgHash []byte) (*DataStoreInit, error) {
	var query struct {
		InitDs []DataStoreInitGql `graphql:"dataStores(where:{msgHash: $msgHash})"`
	}
	msgHashHex := hexutil.Encode(msgHash)
	lowerCaseHex := strings.ToLower(msgHashHex)

	variables := map[string]interface{}{
		"msgHash": graphql.String(lowerCaseHex),
	}

	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		g.Logger.Error().Err(err).Msg("QueryInitDataStore error")
		return nil, err
	}
	if len(query.InitDs) == 0 {
		return nil, errors.New("QueryEmpty")
	}
	return query.InitDs[0].Convert()
}

func (g *GraphClient) QueryInitDataStoreByTxHash(txHash []byte) (*DataStoreInit, error) {
	var query struct {
		InitDs []DataStoreInitGql `graphql:"dataStores(where:{initTxHash: $initTxHash})"`
	}
	txHashHex := hexutil.Encode(txHash)
	lowerCaseHex := strings.ToLower(txHashHex)

	variables := map[string]interface{}{
		"initTxHash": graphql.String(lowerCaseHex),
	}
	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		g.Logger.Error().Err(err).Msg("QueryInitDataStore error")
		return nil, err
	}
	if len(query.InitDs) == 0 {
		return nil, errors.New("QueryEmpty")
	}
	return query.InitDs[0].Convert()
}

func (d *DataStoreInitGql) Convert() (*DataStoreInit, error) {
	sn, err := strconv.ParseUint(
		string(d.StoreNumber), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	ddsId, err := strconv.ParseUint(
		string(d.DurationDataStoreId), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	index, err := strconv.ParseUint(
		string(d.Index), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	dataCommitment, err := hexutil.Decode(string(d.DataCommitment))
	if err != nil {
		return nil, err
	}
	var dataArray [32]byte
	copy(dataArray[:], dataCommitment)

	referenceBlockNumber, err := strconv.ParseUint(
		string(d.ReferenceBlockNumber), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	initTime, err := strconv.ParseUint(
		string(d.InitTime), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	expireTime, err := strconv.ParseUint(
		string(d.ExpireTime), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	storePeriodLength, err := strconv.ParseUint(
		string(d.StorePeriodLength), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	fee, ok := new(big.Int).SetString(
		string(d.Fee),
		10,
	)
	if !ok {
		return nil, ErrFeeStringNotParseable
	}

	txHash, err := hexutil.Decode(string(d.InitTxHash))
	if err != nil {
		return nil, err
	}
	var hashArray [32]byte
	copy(hashArray[:], txHash)

	initGasUsed, err := strconv.ParseUint(
		string(d.InitGasUsed), 10, 64,
	)
	if err != nil {
		return nil, err
	}

	initBlockNumber, ok := new(big.Int).SetString(string(d.InitBlockNumber), 10)
	if !ok {
		return nil, errors.New("improperly formatted block number")
	}

	header, err := hexutil.Decode(string(d.Header))
	if err != nil {
		return nil, err
	}

	msgHash, err := hexutil.Decode(string(d.MsgHash))
	if err != nil {
		return nil, err
	}
	var msgHashArray [32]byte
	copy(msgHashArray[:], msgHash)

	data := DataStoreInit{
		StoreNumber:          uint32(sn),
		Index:                uint32(index),
		DurationDataStoreId:  uint32(ddsId),
		DataCommitment:       dataArray,
		MsgHash:              msgHashArray,
		ReferenceBlockNumber: uint32(referenceBlockNumber),
		InitTime:             uint32(initTime),
		ExpireTime:           uint32(expireTime),
		Duration:             uint8(d.Duration),
		StorePeriodLength:    uint32(storePeriodLength),
		Fee:                  fee,
		Header:               header,
		Confirmer:            string(d.Confirmer),
		InitTxHash:           hashArray,
		InitGasUsed:          initGasUsed,
		InitBlockNumber:      initBlockNumber,
	}
	return &data, nil
}
