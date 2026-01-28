package withdrawal

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/withdrawal"
	gameTypes "github.com/ethereum-optimism/optimism/op-challenger/game/types"
)

func TestWithdrawalMNT_Permissioned(gt *testing.T) {
	withdrawal.TestWithdrawalMNT(gt, gameTypes.PermissionedGameType)
}

func TestWithdrawalETH_Permissioned(gt *testing.T) {
	withdrawal.TestWithdrawalETH(gt, gameTypes.PermissionedGameType)
}
