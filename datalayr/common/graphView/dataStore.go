package graphView

import (
	"context"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shurcooL/graphql"
)

type DataStoreGql struct {
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
	NumSys               graphql.String
	NumPar               graphql.String
	Degree               graphql.String
	StorePeriodLength    graphql.String
	Fee                  graphql.String
	Confirmer            graphql.String
	Header               graphql.String
	InitTxHash           graphql.String
	InitGasUsed          graphql.String
	InitBlockNumber      graphql.String

	Confirmed             graphql.Boolean
	EthSigned             graphql.String
	EigenSigned           graphql.String
	NonSignerPubKeyHashes []graphql.String
	SignatoryRecord       graphql.String
	ConfirmTxHash         graphql.String
	ConfirmGasUsed        graphql.String
}

type DataStore struct {
	StoreNumber          uint32
	DurationDataStoreId  uint32
	Index                uint32
	DataCommitment       [32]byte
	MsgHash              [32]byte
	ReferenceBlockNumber uint32
	InitTime             uint32
	ExpireTime           uint32
	Duration             uint8
	NumSys               uint32
	NumPar               uint32
	Degree               uint32
	StorePeriodLength    uint32
	Fee                  *big.Int
	Header               []byte
	InitTxHash           [32]byte
	InitGasUsed          uint64
	InitBlockNumber      *big.Int

	Confirmer       string
	Confirmed       bool
	EthSigned       *big.Int
	EigenSigned     *big.Int
	PubKeyHashes    [][32]byte
	SignatoryRecord [32]byte
	ConfirmTxHash   [32]byte
	ConfirmGasUsed  uint64
}

func (d *DataStoreGql) Convert() (*DataStore, error) {
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

	msgHash, err := hexutil.Decode(string(d.MsgHash))
	if err != nil {
		return nil, err
	}
	var msgHashArray [32]byte
	copy(msgHashArray[:], msgHash)

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

	numSys, err := strconv.ParseUint(
		string(d.NumSys), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	numPar, err := strconv.ParseUint(
		string(d.NumPar), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	degree, err := strconv.ParseUint(
		string(d.Degree), 10, 32,
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

	initTxHash, err := hexutil.Decode(string(d.InitTxHash))
	if err != nil {
		return nil, err
	}
	var initHashArray [32]byte
	copy(initHashArray[:], initTxHash)

	initGasUsed, err := strconv.ParseUint(
		string(d.InitGasUsed), 10, 64,
	)
	if err != nil {
		return nil, err
	}

	initBlockNumber, err := strconv.ParseUint(
		string(d.InitBlockNumber), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	header, err := hexutil.Decode(string(d.Header))
	if err != nil {
		return nil, err
	}

	ethSigned, ok := new(big.Int).SetString(
		string(d.EthSigned),
		10,
	)
	if !ok {
		return nil, ErrEthSignedStringNotParseable
	}

	eigenSigned, ok := new(big.Int).SetString(
		string(d.EigenSigned),
		10,
	)
	if !ok {
		return nil, ErrEigenSignedStringNotParseable
	}

	confirmTxHash, err := hexutil.Decode(string(d.ConfirmTxHash))
	if err != nil {
		return nil, err
	}
	var confirmHashArray [32]byte
	copy(confirmHashArray[:], confirmTxHash)

	confirmGasUsed, err := strconv.ParseUint(
		string(d.ConfirmGasUsed), 10, 64,
	)
	if err != nil {
		return nil, err
	}

	pubKeyHashes := make([][32]byte, 0)
	for _, pubKeyHash := range d.NonSignerPubKeyHashes {
		var pubKeyHashBytes [32]byte
		copy([]byte(string(pubKeyHash)), pubKeyHashBytes[:])
		pubKeyHashes = append(pubKeyHashes, pubKeyHashBytes)
	}

	signatoryRecord, err := hexutil.Decode(string(d.SignatoryRecord))
	if err != nil {
		return nil, err
	}
	var signatoryRecordArray [32]byte
	copy(signatoryRecordArray[:], signatoryRecord)

	return &DataStore{
		StoreNumber:          uint32(sn),
		DurationDataStoreId:  uint32(ddsId),
		Index:                uint32(index),
		DataCommitment:       dataArray,
		MsgHash:              msgHashArray,
		ReferenceBlockNumber: uint32(referenceBlockNumber),
		InitTime:             uint32(initTime),
		ExpireTime:           uint32(expireTime),
		Duration:             uint8(d.Duration),
		NumSys:               uint32(numSys),
		NumPar:               uint32(numPar),
		Degree:               uint32(degree),
		StorePeriodLength:    uint32(storePeriodLength),
		Fee:                  fee,
		Confirmer:            string(d.Confirmer),
		Header:               header,
		InitTxHash:           initHashArray,
		InitGasUsed:          initGasUsed,
		InitBlockNumber:      new(big.Int).SetUint64(initBlockNumber),
		Confirmed:            bool(d.Confirmed),
		EthSigned:            ethSigned,
		EigenSigned:          eigenSigned,
		PubKeyHashes:         pubKeyHashes,
		SignatoryRecord:      signatoryRecordArray,
		ConfirmTxHash:        confirmHashArray,
		ConfirmGasUsed:       confirmGasUsed,
	}, nil

}

func (g *GraphClient) GetParticipatingDataStores(address string) ([]*DataStore, error) {

	operator, err := g.QueryOperator(address)
	if err != nil {
		return nil, err
	}

	var query struct {
		DataStores []DataStoreGql `graphql:"dataStores(where:{nonSignerPubKeyHashes_not_contains:$pubKeyHash,referenceBlockNumber_gte:$fromBlock,referenceBlockNumber_lte:$toBlock})"`
	}

	variables := map[string]interface{}{
		"pubKeyHash": []graphql.String{operator.Pubkeys.PubkeyHash},
		"fromBlock":  operator.FromBlockNumber,
		"toBlock":    operator.ToBlockNumber,
	}

	client := graphql.NewClient(g.GetEndpoint(), nil)
	err = client.Query(context.Background(), &query, variables)
	if err != nil {
		g.Logger.Error().Err(err).Msg("query error")
		return nil, err
	}

	dataStores := make([]*DataStore, 0)
	for _, dsGql := range query.DataStores {
		ds, err := dsGql.Convert()
		if err != nil {
			g.Logger.Error().Err(err).Msg("error decoding response")
			return nil, err
		}
		dataStores = append(dataStores, ds)
	}

	return dataStores, nil
}
