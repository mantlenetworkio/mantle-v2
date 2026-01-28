package withdrawal

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/withdrawal"
	gameTypes "github.com/ethereum-optimism/optimism/op-challenger/game/types"
)

func TestWithdrawalMNT_Cannon(gt *testing.T) {
	withdrawal.TestWithdrawalMNT(gt, gameTypes.CannonGameType)
}

func TestWithdrawalETH_Cannon(gt *testing.T) {
	withdrawal.TestWithdrawalETH(gt, gameTypes.CannonGameType)
}
