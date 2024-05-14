package graphView

import (
	"context"
	"encoding/hex"
	"math/big"
	"strconv"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shurcooL/graphql"
)

type DataStoreConfirmGql struct {
	EthSigned             graphql.String
	EigenSigned           graphql.String
	NonSignerPubKeyHashes []graphql.String
	ConfirmTxHash         graphql.String
	ConfirmGasUsed        graphql.String
}

type DataStoreConfirm struct {
	EthSigned      *big.Int
	EigenSigned    *big.Int
	PubKeyHashes   [][32]byte
	ConfirmTxHash  [32]byte
	ConfirmGasUsed uint64
}

func (d *DataStoreConfirmGql) Convert() (*DataStoreConfirm, error) {

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

	txHash, err := hexutil.Decode(string(d.ConfirmTxHash))
	if err != nil {
		return nil, err
	}
	var hashArray [32]byte
	copy(hashArray[:], txHash)

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

	data := DataStoreConfirm{
		EthSigned:      ethSigned,
		EigenSigned:    eigenSigned,
		PubKeyHashes:   pubKeyHashes,
		ConfirmTxHash:  hashArray,
		ConfirmGasUsed: confirmGasUsed,
	}
	return &data, nil
}

func (g *GraphClient) QueryDataStoreConfirmation(headerHash []byte) error {
	var query struct {
		DataStoreConfirmation DataStoreConfirmGql `graphql:"dataStore(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(hex.EncodeToString(headerHash)),
	}

	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		g.Logger.Error().Err(err).Msg("query error")
		return err
	}

	return nil
}
