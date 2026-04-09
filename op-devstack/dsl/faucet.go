package dsl

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/retry"
)

// Faucet wraps a stack.Faucet interface for DSL operations.
// A Faucet is chain-specific.
// Note: Faucet wraps a stack component, to share faucet operations in kurtosis by hosting it as service,
// and prevent race-conditions with the account that sends out the faucet funds.
type Faucet struct {
	commonImpl
	inner stack.Faucet
}

// NewFaucet creates a new Faucet DSL wrapper
func NewFaucet(inner stack.Faucet) *Faucet {
	return &Faucet{
		commonImpl: commonFromT(inner.T()),
		inner:      inner,
	}
}

func (f *Faucet) String() string {
	return f.inner.ID().String()
}

// Escape returns the underlying stack.Faucet
func (f *Faucet) Escape() stack.Faucet {
	return f.inner
}

// Fund funds the given address with the given amount of ETH
func (f *Faucet) Fund(addr common.Address, amount eth.ETH) {
	if amount.IsZero() {
		return
	}
	err := retry.Do0(f.ctx, 3, retry.Exponential(), func() error {
		// Use a per-attempt context with a fixed 30s timeout instead of f.ctx.
		//
		// retry.Do0 treats context.DeadlineExceeded originating from f.ctx as a
		// non-retryable signal and stops immediately. By isolating the HTTP call
		// into a child context, a slow or rate-limited faucet response times out
		// within the child context while f.ctx remains alive, allowing retry.Do0
		// to continue its retry loop.
		//
		// 依据：sysext faucet (:39001) 在连续测试间可能处于 rate-limit 或 busy 状态
		// （实测 TestDepositMNTByBridge_ZeroValue 96s 后立即发起请求会在 12s 超时），
		// HTTP 超时错误来自 child context，f.ctx 仍有效，retry.Do0 继续重试。
		// WARN 日志只出现一次证明旧代码中重试被提前终止。
		reqCtx, cancel := context.WithTimeout(f.ctx, 30*time.Second)
		defer cancel()
		err := f.inner.API().RequestETH(reqCtx, addr, amount)
		if err != nil {
			f.log.Warn("Failed to fund address", "addr", addr, "amount", amount, "err", err)
		}
		return err
	})
	f.require.NoError(err, "must fund account %s with %s", addr, amount)
}
