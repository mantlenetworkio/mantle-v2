package sources

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"time"

	"github.com/ethereum-optimism/optimism/op-node/client"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/sources/caching"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rpc"
)

type EngineClientConfig struct {
	L2ClientConfig
}

func EngineClientDefaultConfig(config *rollup.Config) *EngineClientConfig {
	return &EngineClientConfig{
		// engine is trusted, no need to recompute responses etc.
		L2ClientConfig: *L2ClientDefaultConfig(config, true),
	}
}

type EngineVersionProvider interface {
	ForkchoiceUpdatedVersion(attr *eth.PayloadAttributes) eth.EngineAPIMethod
	NewPayloadVersion(timestamp uint64) eth.EngineAPIMethod
	GetPayloadVersion(timestamp uint64) eth.EngineAPIMethod
}

// EngineClient extends L2Client with engine API bindings.
type EngineClient struct {
	*L2Client
	evp EngineVersionProvider
}

func NewEngineClient(client client.RPC, log log.Logger, metrics caching.Metrics, config *EngineClientConfig) (*EngineClient, error) {
	l2Client, err := NewL2Client(client, log, metrics, &config.L2ClientConfig)
	if err != nil {
		return nil, err
	}

	return &EngineClient{
		L2Client: l2Client,
		evp:      config.RollupCfg,
	}, nil
}

// ForkchoiceUpdate updates the forkchoice on the execution client. If attributes is not nil, the engine client will also begin building a block
// based on attributes after the new head block and return the payload ID.
//
// The RPC may return three types of errors:
// 1. Processing error: ForkchoiceUpdatedResult.PayloadStatusV1.ValidationError or other non-success PayloadStatusV1,
// 2. `error` as eth.InputError: the forkchoice state or attributes are not valid.
// 3. Other types of `error`: temporary RPC errors, like timeouts.
func (s *EngineClient) ForkchoiceUpdate(ctx context.Context, fc *eth.ForkchoiceState, attributes *eth.PayloadAttributes) (*eth.ForkchoiceUpdatedResult, error) {
	e := s.log.New("state", fc, "attr", attributes)
	e.Trace("Sharing forkchoice-updated signal")
	fcCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	var result eth.ForkchoiceUpdatedResult
	method := s.evp.ForkchoiceUpdatedVersion(attributes)
	err := s.client.CallContext(fcCtx, &result, string(method), fc, attributes)
	if err == nil {
		e.Trace("Shared forkchoice-updated signal")
		if attributes != nil { // block building is optional, we only get a payload ID if we are building a block
			e.Trace("Received payload id", "payloadId", result.PayloadID)
		}
		return &result, nil
	} else {
		e.Warn("Failed to share forkchoice-updated signal", "err", err)
		if rpcErr, ok := err.(rpc.Error); ok {
			code := eth.ErrorCode(rpcErr.ErrorCode())
			switch code {
			case eth.InvalidForkchoiceState, eth.InvalidPayloadAttributes:
				return nil, eth.InputError{
					Inner: err,
					Code:  code,
				}
			default:
				return nil, fmt.Errorf("unrecognized rpc error: %w", err)
			}
		}
		return nil, err
	}
}

// NewPayload executes a full block on the execution engine.
// This returns a PayloadStatusV1 which encodes any validation/processing error,
// and this type of error is kept separate from the returned `error` used for RPC errors, like timeouts.
func (s *EngineClient) NewPayload(ctx context.Context, payload *eth.ExecutionPayload, parentBeaconBlockRoot *common.Hash) (*eth.PayloadStatusV1, error) {
	e := s.log.New("block_hash", payload.BlockHash)
	e.Trace("sending payload for execution")

	execCtx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var result eth.PayloadStatusV1

	var err error
	switch method := s.evp.NewPayloadVersion(uint64(payload.Timestamp)); method {
	case eth.NewPayloadV4:
		err = s.client.CallContext(execCtx, &result, string(method), payload, []common.Hash{}, parentBeaconBlockRoot, []hexutil.Bytes{})
	case eth.NewPayloadV1:
		err = s.client.CallContext(execCtx, &result, string(method), payload)
	default:
		return nil, fmt.Errorf("unsupported NewPayload version: %s", method)
	}

	e.Trace("Received payload execution result", "status", result.Status, "latestValidHash", result.LatestValidHash, "message", result.ValidationError)
	if err != nil {
		e.Error("Payload execution failed", "err", err)
		return nil, fmt.Errorf("failed to execute payload: %w", err)
	}
	return &result, nil
}

// GetPayload gets the execution payload associated with the PayloadId.
// There may be two types of error:
// 1. `error` as eth.InputError: the payload ID may be unknown
// 2. Other types of `error`: temporary RPC errors, like timeouts.
func (s *EngineClient) GetPayload(ctx context.Context, payloadInfo eth.PayloadInfo) (*eth.ExecutionPayloadEnvelope, error) {
	e := s.log.New("payload_id", payloadInfo.ID)
	e.Trace("getting payload")
	var result eth.ExecutionPayloadEnvelope
	method := s.evp.GetPayloadVersion(payloadInfo.Timestamp)
	err := s.client.CallContext(ctx, &result, string(method), payloadInfo.ID)
	if err != nil {
		e.Warn("Failed to get payload", "payload_id", payloadInfo.ID, "err", err)
		if rpcErr, ok := err.(rpc.Error); ok {
			code := eth.ErrorCode(rpcErr.ErrorCode())
			switch code {
			case eth.UnknownPayload:
				return nil, eth.InputError{
					Inner: err,
					Code:  code,
				}
			default:
				return nil, fmt.Errorf("unrecognized rpc error: %w", err)
			}
		}
		return nil, err
	}
	e.Trace("Received payload")
	return &result, nil
}
