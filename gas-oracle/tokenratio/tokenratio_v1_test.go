package tokenratio

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTokenPriceV1(t *testing.T) {
	rpc := rpcURL(t)
	tokenPricer := NewClient(cexURL(t), rpc, 3)
	ethPrice, err := tokenPricer.query(ETHUSDT)
	require.NoError(t, err)
	t.Logf("ETH price:%v", ethPrice)

	mntPrice, err := tokenPricer.query(MNTUSDT)
	require.NoError(t, err)
	t.Logf("MNT price:%v", mntPrice)
}
