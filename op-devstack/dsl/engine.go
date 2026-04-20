package dsl

import (
	"errors"
	"fmt"
	"time"

	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/retry"
)

type NewPayloadResult struct {
	T      devtest.T
	Status *eth.PayloadStatusV1
	Err    error
}

func (r *NewPayloadResult) IsPayloadStatus(status eth.ExecutePayloadStatus) *NewPayloadResult {
	r.T.Require().NotNil(r.Status, "payload status nil")
	r.T.Require().Equal(status, r.Status.Status)
	return r
}

func (r *NewPayloadResult) IsSyncing() *NewPayloadResult {
	r.IsPayloadStatus(eth.ExecutionSyncing)
	r.T.Require().NoError(r.Err)
	return r
}

func (r *NewPayloadResult) IsValid() *NewPayloadResult {
	r.IsPayloadStatus(eth.ExecutionValid)
	r.T.Require().NoError(r.Err)
	return r
}

// IsValidOrSyncing accepts either VALID or SYNCING status.
// VALID: geth when parent is in non-canonical chain (has parent state), reth after queuing+processing.
// SYNCING: geth/reth when parent is genuinely unavailable.
func (r *NewPayloadResult) IsValidOrSyncing() *NewPayloadResult {
	r.T.Require().NoError(r.Err)
	r.T.Require().NotNil(r.Status, "payload status nil")
	r.T.Require().True(
		r.Status.Status == eth.ExecutionValid || r.Status.Status == eth.ExecutionSyncing,
		"expected VALID or SYNCING, got %s", r.Status.Status,
	)
	return r
}

func (r *NewPayloadResult) IsInvalid() *NewPayloadResult {
	r.IsPayloadStatus(eth.ExecutionInvalid)
	r.T.Require().NoError(r.Err)
	return r
}

type ForkchoiceUpdateResult struct {
	T          devtest.T
	Refresh    func()
	Result     *eth.ForkchoiceUpdatedResult
	ValidCnt   int // count for VALID response
	SyncingCnt int // count for SYNCING response
	InvalidCnt int // count for INVALID response
	RefreshCnt int
	Err        error
}

func (r *ForkchoiceUpdateResult) IsForkchoiceUpdatedStatus(status eth.ExecutePayloadStatus) *ForkchoiceUpdateResult {
	r.T.Require().NotNil(r.Result, "fcu result nil")
	r.T.Require().Equal(status, r.Result.PayloadStatus.Status)
	return r
}

func (r *ForkchoiceUpdateResult) IsSyncing() *ForkchoiceUpdateResult {
	r.IsForkchoiceUpdatedStatus(eth.ExecutionSyncing)
	r.T.Require().NoError(r.Err)
	return r
}

// IsValidOrSyncing accepts either VALID or SYNCING FCU status.
// On geth, FCU to an unknown/unavailable head returns SYNCING.
// On reth, the same FCU may return VALID if the block was previously buffered in the engine tree.
func (r *ForkchoiceUpdateResult) IsValidOrSyncing() *ForkchoiceUpdateResult {
	r.T.Require().NoError(r.Err)
	r.T.Require().NotNil(r.Result, "fcu result nil")
	status := r.Result.PayloadStatus.Status
	r.T.Require().True(
		status == eth.ExecutionValid || status == eth.ExecutionSyncing,
		"expected VALID or SYNCING, got %s", status,
	)
	return r
}

func (r *ForkchoiceUpdateResult) IsValid() *ForkchoiceUpdateResult {
	r.IsForkchoiceUpdatedStatus(eth.ExecutionValid)
	r.T.Require().NoError(r.Err)
	return r
}

// WaitUntilValid polls the FCU result up to `attempts` times with a 1s fixed
// interval (total wall-clock budget = attempts × 1s). Bridges reth's async
// pipeline, where FCU may transiently return SYNCING before transitioning to
// VALID; on geth (synchronous pipeline) the first attempt typically succeeds.
func (r *ForkchoiceUpdateResult) WaitUntilValid(attempts int) *ForkchoiceUpdateResult {
	tryCnt := 0
	err := retry.Do0(r.T.Ctx(), attempts, &retry.FixedStrategy{Dur: 1 * time.Second},
		func() error {
			r.Refresh()
			tryCnt += 1
			if r.Err != nil {
				return fmt.Errorf("forkchoice returned error: %w", r.Err)
			}
			if r.Result == nil {
				return errors.New("forkchoice has empty result")
			}
			if r.Result.PayloadStatus.Status != eth.ExecutionValid {
				r.T.Logger().Info("Wait for FCU to return valid", "status", r.Result.PayloadStatus.Status, "try_count", tryCnt)
				return errors.New("still syncing")
			}
			return nil
		})
	r.T.Require().NoError(err)
	return r
}

func (r *ForkchoiceUpdateResult) Retry(attempts int) *ForkchoiceUpdateResult {
	tryCnt := 0
	err := retry.Do0(r.T.Ctx(), attempts, &retry.FixedStrategy{Dur: 500 * time.Millisecond},
		func() error {
			r.Refresh()
			tryCnt += 1
			if r.Err != nil {
				return fmt.Errorf("forkchoice returned error: %w", r.Err)
			}
			if r.Result == nil {
				return errors.New("forkchoice has empty result")
			}
			r.T.Logger().Info("Retrying FCU", "status", r.Result.PayloadStatus.Status, "try_count", tryCnt)
			return errors.New("retry")
		})
	r.T.Require().Error(err) // always return error for retrying
	return r
}

func (r *ForkchoiceUpdateResult) ResultAllSyncing() {
	r.T.Require().Equal(r.RefreshCnt, r.SyncingCnt)
}
