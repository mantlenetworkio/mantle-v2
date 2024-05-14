package graphView

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/shurcooL/graphql"
)

type OperatorStakeGql struct {
	ToBlockNumber     graphql.String
	MantleFirstStake  graphql.String
	MantleSencodStake graphql.String
}

type OperatorIndexGql struct {
	ToBlockNumber graphql.String
	Index         graphql.String
}

type OperatorPubkeys struct {
	PubkeyHash graphql.String
	PubkeyG1   []graphql.String
	PubkeyG2   []graphql.String
}

type OperatorGql struct {
	Id              graphql.String
	Pubkeys         OperatorPubkeys `graphql:"pubkeys"`
	FromBlockNumber graphql.String
	ToBlockNumber   graphql.String
	Socket          graphql.String
}

type OperatorBlockView struct {
	Id           graphql.String
	Pubkeys      OperatorPubkeys `graphql:"pubkeys"`
	Socket       graphql.String
	StakeHistory []struct {
		MantleFirstStake  graphql.String
		MantleSencodStake graphql.String
	} `graphql:"stakeHistory(first:1, orderBy:toBlockNumber, orderDirection:desc, where: {toBlockNumber_gte:$blockNumber})"`
	IndexHistory []struct {
		Index graphql.String
	} `graphql:"indexHistory(first:1, orderBy:toBlockNumber, orderDirection:desc, where: {toBlockNumber_gte:$blockNumber})"`
}

// Used to check if an operator is already registered
func (g *GraphClient) QueryOperator(addr string) (OperatorGql, error) {
	var query struct {
		Operator OperatorGql `graphql:"operator(id: $id)"`
	}

	lowerCaseAddr := strings.ToLower(addr)

	variables := map[string]interface{}{
		"id": graphql.String(lowerCaseAddr),
	}
	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		return OperatorGql{}, err
	} else {
		if query.Operator.Id == "" {
			obj, _ := json.Marshal(query.Operator)
			g.Logger.Debug().Str("address", addr).Str("Operator", string(obj)).Msg("Returned no operator")
			return OperatorGql{}, ErrOperatorNotFound
		}
		return query.Operator, nil
	}
}

func (g *GraphClient) QueryOperatorsByStatus() ([]OperatorGql, error) {
	var query struct {
		Operators []OperatorGql `graphql:"operators(where: {status:0})"`
	}
	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		g.Logger.Error().Err(err).Msg("query error")
		return nil, err
	}
	g.Logger.Info().Msg("Query result")
	for _, op := range query.Operators {
		g.Logger.Info().Msgf("Id %v", op.Id)
		g.Logger.Info().Msgf("PubKey %v", op.Pubkeys)
		g.Logger.Info().Msgf("FromBlockNumber %v", op.FromBlockNumber)
	}
	return query.Operators, nil
}

func (g *GraphClient) QueryOperators() ([]OperatorGql, error) {
	var query struct {
		Operators []OperatorGql `graphql:"operators"`
	}

	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		g.Logger.Error().Err(err).Msg("query error")
		return nil, err
	}
	g.Logger.Info().Msg("Query result")
	for _, op := range query.Operators {
		g.Logger.Info().Msgf("Id %v", op.Id)
		g.Logger.Info().Msgf("PubKey %v", op.Pubkeys)
		g.Logger.Info().Msgf("FromBlockNumber %v", op.FromBlockNumber)
	}
	return query.Operators, nil
}

func (g *GraphClient) QueryOperatorsViewByBlock(ctx context.Context, blockNumber uint32) (
	[]OperatorBlockView,
	error,
) {
	log := g.Logger.SubloggerId(ctx)

	var query struct {
		Operators []OperatorBlockView `graphql:"operators(where: {fromBlockNumber_lte:$blockNumber,toBlockNumber_gt:$blockNumber,status:0})"`
	}
	variables := map[string]interface{}{
		"blockNumber": graphql.String(fmt.Sprintf("%v", blockNumber)),
	}

	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &query, variables)
	if err != nil {
		log.Error().Err(err).Msg("QueryBlockOperatorMeta error")
		return nil, err
	}
	log.Trace().Msgf("QueryBlockOperatorMeta %v", query)

	return query.Operators, nil
}

type TotalOperatorGql struct {
	ToBlockNumber graphql.String
	Count         graphql.String
	AggPubKeyHash graphql.String
	AggPubKey     []graphql.String
	Index         graphql.Int
}

func (d *TotalOperatorGql) Convert() (*TotalOperator, error) {
	toBlockNumber, err := strconv.ParseUint(
		string(d.ToBlockNumber), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	count, err := strconv.ParseUint(
		string(d.Count), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	aggPubKeyHash, err := hexutil.Decode(string(d.AggPubKeyHash))
	if err != nil {
		return nil, err
	}
	var aggPubKeyHashArray [32]byte
	copy(aggPubKeyHashArray[:], aggPubKeyHash)

	if len(d.AggPubKey) != 2 {
		return nil, ErrInvalidG1PubkeyLength
	}

	var aggPubKeyStrings [2]string
	for i := 0; i < len(d.AggPubKey); i++ {
		aggPubKeyStrings[i] = string(d.AggPubKey[i])
	}

	aggPubKey := bls.ConvertStringsToG1Point(aggPubKeyStrings)

	index := uint32(d.Index)

	return &TotalOperator{
		ToBlockNumber: uint64(toBlockNumber),
		Count:         uint64(count),
		AggPubKeyHash: aggPubKeyHashArray,
		AggPubKey:     aggPubKey,
		Index:         uint32(index),
	}, nil
}

type TotalStakeGql struct {
	ToBlockNumber     graphql.String
	MantleFirstStake  graphql.String
	MantleSencodStake graphql.String
	Index             graphql.Int
}

func (d *TotalStakeGql) Convert() (*TotalStake, error) {

	toBlockNumber, err := strconv.ParseUint(
		string(d.ToBlockNumber), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	mantleFirstStake, ok := new(big.Int).SetString(
		string(d.MantleFirstStake),
		10,
	)
	if !ok {
		return nil, ErrMantleFirstStakedStringNotParseable
	}

	mantleSencodStake, ok := new(big.Int).SetString(
		string(d.MantleSencodStake),
		10,
	)
	if !ok {
		return nil, ErrEigenSignedStringNotParseable
	}

	index := uint64(d.Index)

	return &TotalStake{
		ToBlockNumber: toBlockNumber,
		QuorumStakes:  []*big.Int{mantleFirstStake, mantleSencodStake},
		Index:         uint32(index),
	}, nil
}
