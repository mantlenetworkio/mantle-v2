package graphView

import (
	"context"
	"fmt"
	"math/big"
	"strconv"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/ethereum/go-ethereum/common"
	"github.com/shurcooL/graphql"
)

type RegistrantView struct {
	*Registrant
	QuorumStakes []*big.Int
	Index        uint32
}

type StateView struct {
	Registrants   []*RegistrantView
	RegistrantMap map[common.Address]*RegistrantView
	BlockNumber   uint32
	TotalStake    *TotalStake
	TotalOperator *TotalOperator
}

// Query and Convert graphql struct to view struct
func (g *GraphClient) GetStateView(ctx context.Context, blockNumber uint32) (*StateView, error) {
	ops, err := g.QueryOperatorsViewByBlock(ctx, blockNumber)
	if err != nil {
		return nil, err
	}

	// format
	registrants := make([]*RegistrantView, len(ops))
	registrantMap := make(map[common.Address]*RegistrantView)
	for i, op := range ops {
		registrantView, err := op.RegistrantView()
		if err != nil {
			return nil, err
		}

		registrants[i] = registrantView
		registrantMap[registrantView.Address] = registrantView
	}

	totalStake, totalOp, err := g.QueryOperatorTotalsByBlock(ctx, blockNumber)
	if err != nil {
		return nil, err
	}
	stateView := StateView{
		Registrants:   registrants,
		RegistrantMap: registrantMap,
		BlockNumber:   blockNumber,
		TotalStake:    totalStake,
		TotalOperator: totalOp,
	}

	return &stateView, nil
}

func (g *GraphClient) QueryOperatorTotalsByBlock(ctx context.Context, blockNumber uint32) (
	*TotalStake,
	*TotalOperator,
	error,
) {
	log := g.Logger.SubloggerId(ctx)

	variables := map[string]interface{}{
		"blockNumber": graphql.String(fmt.Sprintf("%v", blockNumber)),
	}

	var stakeQuery struct {
		TotalStakes []TotalStakeGql `graphql:"totalStakes(first:1, sorderBy:toBlockNumber, orderDirection:desc, where: {toBlockNumber_gte:4294967295})"`
	}

	client := graphql.NewClient(g.GetEndpoint(), nil)
	err := client.Query(context.Background(), &stakeQuery, variables)
	if err != nil {
		log.Error().Err(err).Msg("Query Stake error")
		return nil, nil, err
	}

	if len(stakeQuery.TotalStakes) == 0 {
		return nil, nil, ErrEmptyTotalStakes
	}

	totalStake, err := stakeQuery.TotalStakes[0].Convert()
	if err != nil {
		log.Error().Err(err).Msg("Parse stakes error")
		return nil, nil, err
	}

	// Query ops idx
	var opsQuery struct {
		TotalOperators []TotalOperatorGql `graphql:"totalOperators(first:1, sorderBy:toBlockNumber, orderDirection:desc, where: {toBlockNumber_gte:4294967295})"`
	}
	err = client.Query(context.Background(), &opsQuery, variables)
	if err != nil {
		log.Error().Err(err).Msg("Query ops error")
		return nil, nil, err
	}

	if len(opsQuery.TotalOperators) == 0 {
		return nil, nil, ErrEmptyTotalOperators
	}

	totalOps, err := opsQuery.TotalOperators[0].Convert()
	if err != nil {
		log.Error().Err(err).Msg("parse ops  error")
		return nil, nil, err
	}

	return totalStake, totalOps, nil
}

func (d *OperatorBlockView) RegistrantView() (
	*RegistrantView,
	error,
) {
	address := common.HexToAddress(string(d.Id))

	if len(d.Pubkeys.PubkeyG2) != 4 {
		return nil, ErrInvalidG2PubkeyLength
	}

	var pubkeyG2Strings [4]string
	for i := 0; i < len(d.Pubkeys.PubkeyG2); i++ {
		pubkeyG2Strings[i] = string(d.Pubkeys.PubkeyG2[i])
	}

	pubkeyG2 := bls.ConvertStringsToG2Point(pubkeyG2Strings)

	if len(d.Pubkeys.PubkeyG1) != 2 {
		return nil, ErrInvalidG1PubkeyLength
	}

	var pubkeyG1Strings [2]string
	for i := 0; i < len(d.Pubkeys.PubkeyG1); i++ {
		pubkeyG1Strings[i] = string(d.Pubkeys.PubkeyG1[i])
	}

	pubkeyG1 := bls.ConvertStringsToG1Point(pubkeyG1Strings)

	mantleFirstStake, ok := new(big.Int).SetString(
		string(d.StakeHistory[0].MantleFirstStake),
		10,
	)
	if !ok {
		return nil, ErrMantleFirstStakedStringNotParseable
	}

	mantleSencodStake, ok := new(big.Int).SetString(
		string(d.StakeHistory[0].MantleSencodStake),
		10,
	)
	if !ok {
		return nil, ErrMantleSencodStakedStringNotParseable
	}

	index, err := strconv.ParseUint(
		string(d.IndexHistory[0].Index), 10, 32,
	)
	if err != nil {
		return nil, err
	}

	registrant := Registrant{
		Address:  address,
		Socket:   string(d.Socket),
		PubkeyG1: pubkeyG1,
		PubkeyG2: pubkeyG2,
	}

	registrantView := RegistrantView{
		Registrant:   &registrant,
		QuorumStakes: []*big.Int{mantleFirstStake, mantleSencodStake},
		Index:        uint32(index),
	}

	return &registrantView, nil
}
