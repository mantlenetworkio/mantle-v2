package tokenratio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTokenPriceV5(t *testing.T) {
	tokenPricer := NewClient("https://api.bybit.com", "https://mainnet.infura.io/v3/4f4692085f1340c2a645ae04d36c2321", 3, 0)
	ethPrice, err := tokenPricer.queryV5(ETHUSDT)
	require.NoError(t, err)
	t.Logf("ETH price:%v", ethPrice)

	mntPrice, err := tokenPricer.queryV5(MNTUSDT)
	require.NoError(t, err)
	t.Logf("MNT price:%v", mntPrice)
}
