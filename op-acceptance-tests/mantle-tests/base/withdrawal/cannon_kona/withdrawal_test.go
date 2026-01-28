package withdrawal

import (
	"testing"

	"github.com/ethereum-optimism/optimism/op-acceptance-tests/mantle-tests/base/withdrawal"
	gameTypes "github.com/ethereum-optimism/optimism/op-challenger/game/types"
)

func TestWithdrawalMNT_CannonKona(gt *testing.T) {
	withdrawal.TestWithdrawalMNT(gt, gameTypes.CannonKonaGameType)
}

func TestWithdrawalETH_CannonKona(gt *testing.T) {
	withdrawal.TestWithdrawalETH(gt, gameTypes.CannonKonaGameType)
}
