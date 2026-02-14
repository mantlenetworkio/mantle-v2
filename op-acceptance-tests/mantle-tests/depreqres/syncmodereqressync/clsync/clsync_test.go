package clsync

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/depreqres/common"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
)

// Regression: req/resp sync 开启（默认）时，CL sync 模式下断网/重启的停滞与恢复。
func TestUnsafeChainNotStalling_CLSync_Short(gt *testing.T) {
	common.UnsafeChainNotStalling_Disconnect(gt, sync.CLSync, 20*time.Second)
}

func TestUnsafeChainNotStalling_CLSync_Long(gt *testing.T) {
	common.UnsafeChainNotStalling_Disconnect(gt, sync.CLSync, 95*time.Second)
}

func TestUnsafeChainNotStalling_CLSync_RestartOpNode_Long(gt *testing.T) {
	common.UnsafeChainNotStalling_RestartOpNode(gt, sync.CLSync, 95*time.Second)
}
