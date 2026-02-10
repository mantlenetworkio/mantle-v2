package elsync

import (
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/depreqres/common"
	"github.com/ethereum-optimism/optimism/op-node/rollup/sync"
)

// Regression: req/resp sync 关闭时，EL sync 模式下的分叉停滞与恢复。
func TestUnsafeChainNotStalling_ELSync_Short(gt *testing.T) {
	common.UnsafeChainNotStalling_Disconnect(gt, sync.ELSync, 20*time.Second)
}

func TestUnsafeChainNotStalling_ELSync_Long(gt *testing.T) {
	common.UnsafeChainNotStalling_Disconnect(gt, sync.ELSync, 95*time.Second)
}

func TestUnsafeChainNotStalling_ELSync_RestartOpNode_Long(gt *testing.T) {
	common.UnsafeChainNotStalling_RestartOpNode(gt, sync.ELSync, 95*time.Second)
}
