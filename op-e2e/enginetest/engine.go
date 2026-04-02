package enginetest

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/services"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	gn "github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/require"
)

var (
	// ErrForkChoiceUpdatedNotValid is returned when a forkChoiceUpdated returns a status other than Valid
	ErrForkChoiceUpdatedNotValid = errors.New("forkChoiceUpdated status was not valid")
	// ErrNewPayloadNotValid is returned when a newPayload call returns a status other than Valid, indicating the new block is invalid
	ErrNewPayloadNotValid = errors.New("newPayload status was not valid")
)

// OpEngine is a backend-agnostic actor that drives an L2 execution engine
// via the Engine API. It supports both in-process geth and external reth nodes.
type OpEngine struct {
	node          services.EthInstance
	l2Engine      *sources.EngineClient
	L2Client      *ethclient.Client
	SystemConfig  eth.SystemConfig
	L1ChainConfig *params.ChainConfig
	L2ChainConfig *params.ChainConfig
	L1Head        eth.BlockInfo
	L2Head        *eth.ExecutionPayload
	sequenceNum   uint64
	lgr           log.Logger
}

// EngineFactory creates an OpEngine backed by a specific L2 client.
// It is intended for writing shared test suites that run against both geth and
// reth without code duplication. Each backend package (opgeth, opreth, etc.)
// implements this signature and passes it to suite runner functions.
type EngineFactory func(t testing.TB, ctx context.Context, cfg *e2esys.SystemConfig) (*OpEngine, error)

// OpEngineConfig holds the already-built objects needed to construct an OpEngine.
// Factory functions (NewGethEngine, NewRethEngine) build these, then pass to NewOpEngine.
type OpEngineConfig struct {
	Node           services.EthInstance    // already started
	RollupCfg      *rollup.Config
	RollupGenesis  rollup.Genesis
	L1ChainConfig  *params.ChainConfig
	L2ChainConfig  *params.ChainConfig
	L1Head         eth.BlockInfo
	GenesisPayload *eth.ExecutionPayload
}

// NewOpEngine constructs an OpEngine from an already-started node and pre-built configs.
// It creates the Engine API client and JSON-RPC client. The node must be started beforehand.
func NewOpEngine(t testing.TB, ctx context.Context, cfg *e2esys.SystemConfig, ec OpEngineConfig) (*OpEngine, error) {
	t.Helper()
	logger := testlog.Logger(t, log.LevelWarn)

	auth := rpc.WithHTTPAuth(gn.NewJWTAuth(cfg.JWTSecret))
	rpcNode, err := client.NewRPC(ctx, logger, ec.Node.AuthRPC().RPC(), client.WithGethRPCOptions(auth))
	require.NoError(t, err)

	l2Engine, err := sources.NewEngineClient(
		rpcNode,
		logger,
		nil,
		sources.EngineClientDefaultConfig(ec.RollupCfg),
	)
	require.NoError(t, err)

	l2Client, err := ethclient.Dial(ec.Node.UserRPC().RPC())
	require.NoError(t, err)

	return &OpEngine{
		node:          ec.Node,
		L2Client:      l2Client,
		l2Engine:      l2Engine,
		SystemConfig:  ec.RollupGenesis.SystemConfig,
		L1ChainConfig: ec.L1ChainConfig,
		L2ChainConfig: ec.L2ChainConfig,
		L1Head:        ec.L1Head,
		L2Head:        ec.GenesisPayload,
		lgr:           logger,
	}, nil
}

// Engine returns the underlying EngineClient for direct Engine API access.
func (d *OpEngine) Engine() *sources.EngineClient {
	return d.l2Engine
}

// Close shuts down all clients and the underlying node.
func (d *OpEngine) Close() {
	if err := d.node.Close(); err != nil {
		d.lgr.Error("error closing node", "err", err)
	}
	d.l2Engine.Close()
	d.L2Client.Close()
}

// AddL2Block appends a new L2 block to the current chain including the specified transactions.
// The L1Info transaction is automatically prepended to the created block.
func (d *OpEngine) AddL2Block(ctx context.Context, txs ...*types.Transaction) (*eth.ExecutionPayloadEnvelope, error) {
	attrs, err := d.CreatePayloadAttributes(txs...)
	if err != nil {
		return nil, err
	}
	res, err := d.StartBlockBuilding(ctx, attrs)
	if err != nil {
		return nil, fmt.Errorf("start block building: %w", err)
	}

	envelope, err := d.l2Engine.GetPayload(ctx, eth.PayloadInfo{ID: *res.PayloadID, Timestamp: uint64(attrs.Timestamp)})
	if err != nil {
		return nil, fmt.Errorf("get payload: %w", err)
	}
	payload := envelope.ExecutionPayload

	if !reflect.DeepEqual(payload.Transactions, attrs.Transactions) {
		return nil, errors.New("required transactions were not included")
	}

	status, err := d.l2Engine.NewPayload(ctx, payload, envelope.ParentBeaconBlockRoot)
	if err != nil {
		return nil, fmt.Errorf("new payload: %w", err)
	}
	if status.Status != eth.ExecutionValid {
		return nil, fmt.Errorf("%w: %s", ErrNewPayloadNotValid, status.Status)
	}

	fc := eth.ForkchoiceState{
		HeadBlockHash: payload.BlockHash,
		SafeBlockHash: payload.BlockHash,
	}
	res, err = d.l2Engine.ForkchoiceUpdate(ctx, &fc, nil)
	if err != nil {
		return nil, fmt.Errorf("forkchoice update: %w", err)
	}
	if res.PayloadStatus.Status != eth.ExecutionValid {
		return nil, fmt.Errorf("%w: %s", ErrForkChoiceUpdatedNotValid, res.PayloadStatus.Status)
	}
	d.L2Head = payload
	d.sequenceNum = d.sequenceNum + 1
	return envelope, nil
}

// StartBlockBuilding sends engine_forkChoiceUpdated to begin block building.
// The current L2Head is used as the parent of the new block.
func (d *OpEngine) StartBlockBuilding(ctx context.Context, attrs *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	fc := eth.ForkchoiceState{
		HeadBlockHash: d.L2Head.BlockHash,
		SafeBlockHash: d.L2Head.BlockHash,
	}
	res, err := d.l2Engine.ForkchoiceUpdate(ctx, &fc, attrs)
	if err != nil {
		return nil, err
	}
	if res.PayloadStatus.Status != eth.ExecutionValid {
		return nil, fmt.Errorf("%w: %s", ErrForkChoiceUpdatedNotValid, res.PayloadStatus.Status)
	}
	if res.PayloadID == nil {
		return nil, errors.New("forkChoiceUpdated returned nil PayloadID")
	}
	return res, nil
}

// CreatePayloadAttributes creates valid PayloadAttributes containing an L1Info deposit transaction
// followed by the supplied user transactions.
func (d *OpEngine) CreatePayloadAttributes(txs ...*types.Transaction) (*eth.PayloadAttributes, error) {
	timestamp := d.L2Head.Timestamp + 2
	l1Info, err := derive.L1InfoDepositBytes(d.l2Engine.RollupConfig(), d.L1ChainConfig, d.SystemConfig, d.sequenceNum, d.L1Head, uint64(timestamp))
	if err != nil {
		return nil, err
	}

	var txBytes []hexutil.Bytes
	txBytes = append(txBytes, l1Info)
	for _, tx := range txs {
		bin, err := tx.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("tx marshalling failed: %w", err)
		}
		txBytes = append(txBytes, bin)
	}

	var withdrawals *types.Withdrawals
	if d.L2ChainConfig.IsCanyon(uint64(timestamp)) {
		withdrawals = &types.Withdrawals{}
	}

	var parentBeaconBlockRoot *common.Hash
	if d.L2ChainConfig.IsEcotone(uint64(timestamp)) {
		parentBeaconBlockRoot = d.L1Head.ParentBeaconRoot()
		if parentBeaconBlockRoot == nil {
			parentBeaconBlockRoot = &(common.Hash{})
		}
	}

	attrs := eth.PayloadAttributes{
		Timestamp:             timestamp,
		Transactions:          txBytes,
		NoTxPool:              true,
		GasLimit:              (*eth.Uint64Quantity)(&d.SystemConfig.GasLimit),
		Withdrawals:           withdrawals,
		ParentBeaconBlockRoot: parentBeaconBlockRoot,
	}
	if d.L2ChainConfig.IsJovian(uint64(timestamp)) {
		attrs.MinBaseFee = new(uint64)
		*attrs.MinBaseFee = d.SystemConfig.MinBaseFee
	}
	if d.L2ChainConfig.IsHolocene(uint64(timestamp)) {
		attrs.EIP1559Params = new(eth.Bytes8)
		*attrs.EIP1559Params = d.SystemConfig.EIP1559Params
	}
	return &attrs, nil
}
