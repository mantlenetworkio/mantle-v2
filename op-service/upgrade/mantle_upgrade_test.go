package upgrade

import (
	"fmt"
	"math/big"
	"testing"
)

func TestDaUpgrade(t *testing.T) {
	tt := []int{5000, 5003, 5003003, 17, 0}
	for _, t := range tt {
		cid := big.NewInt(int64(t))
		cfg := GetUpgradeConfigForMantle(cid)
		fmt.Println(t, 3274000, cfg.IsEqualEigenDaUpgradeBlock(big.NewInt(3274001)))
		fmt.Println(t, 3274000, cfg.IsEqualEigenDaUpgradeBlock(big.NewInt(3273999)))
		fmt.Println(t, 3274000, cfg.IsEqualEigenDaUpgradeBlock(big.NewInt(3274000)))
		fmt.Println(t, 3274000, cfg.IsUseEigenDa(big.NewInt(3274001)))
		fmt.Println(t, 3274000, cfg.IsUseEigenDa(big.NewInt(3273999)))
		fmt.Println(t, 3274000, cfg.IsUseEigenDa(big.NewInt(3274000)))
	}
}
