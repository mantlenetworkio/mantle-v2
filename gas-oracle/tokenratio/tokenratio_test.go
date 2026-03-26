package tokenratio

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// Integration tests require external services (Bybit CEX API + Ethereum mainnet RPC for Uniswap).
// Set TOKEN_RATIO_RPC_URL to an Ethereum mainnet RPC endpoint to enable Uniswap tests.
// Set TOKEN_RATIO_CEX_URL to a Bybit API endpoint to enable CEX tests (defaults to https://api.bybit.com).
// Example:
//
//	TOKEN_RATIO_RPC_URL=https://mainnet.infura.io/v3/<key> go test ./gas-oracle/tokenratio/...

func cexURL(t *testing.T) string {
	t.Helper()
	if url := os.Getenv("TOKEN_RATIO_CEX_URL"); url != "" {
		return url
	}
	return "https://api.bybit.com"
}

func rpcURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TOKEN_RATIO_RPC_URL")
	if url == "" {
		t.Skip("TOKEN_RATIO_RPC_URL not set, skipping test that requires Ethereum mainnet RPC")
	}
	return url
}

func TestGetTokenPrice(t *testing.T) {
	rpc := rpcURL(t)
	tokenPricer := NewClient(cexURL(t), rpc, 3)

	ethPrice, err := tokenPricer.queryV5("ETHUSDT")
	require.NoError(t, err)
	t.Logf("ETH price:%v", ethPrice)

	mntPrice, err := tokenPricer.queryV5("MNTUSDT")
	require.NoError(t, err)
	t.Logf("MNT price:%v", mntPrice)

	t.Logf("ratio:%v", ethPrice/mntPrice)
	mntPrice, ethPrice = tokenPricer.getTokenPricesFromCex()
	t.Logf("ETH price:%v", ethPrice)
	t.Logf("MNT price:%v", mntPrice)
	t.Logf("ratio:%v", ethPrice/mntPrice)

	eth2bitPrice, err := tokenPricer.getTokenPriceFromUniswap(wETHAddress, mntTokenAddress, mntTokenDecimals)
	require.NoError(t, err)
	t.Logf("ETH/MNT:%v", eth2bitPrice)

	ethPrice, err = tokenPricer.getTokenPriceFromUniswap(wETHAddress, usdtAddress, usdtDecimals)
	require.NoError(t, err)
	t.Logf("ETH/USDT:%v", ethPrice)
}

func TestGetTokenPriceWithRealTokenRatioMode(t *testing.T) {
	rpc := rpcURL(t)
	tokenPricer := NewClient(cexURL(t), rpc, 3)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	t.Logf("ratio:%v", ratio)
}

func TestGetTokenPriceWithOneDollarTokenRatioMode(t *testing.T) {
	rpc := rpcURL(t)
	tokenPricer := NewClient(cexURL(t), rpc, 3)

	ethPrice, err := tokenPricer.queryV5("ETHUSDT")
	require.NoError(t, err)
	t.Logf("ETH price:%v", ethPrice)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	t.Logf("ratio:%v", ratio)
}

func TestGetTokenPriceWithOneDollarTokenRatioMode2(t *testing.T) {
	rpc := rpcURL(t)
	tokenPricer := NewClient("", rpc, 3)

	_, ethPrice := tokenPricer.getTokenPricesFromUniswap()
	t.Logf("ETH price:%v", ethPrice)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	t.Logf("ratio:%v", ratio)
}

// TestGetTokenPriceWithOneDollarTokenRatioMode3 tests fallback when both CEX and RPC URLs are invalid.
// When all price sources fail, tokenRatio() falls back to lastEthPrice/lastMntPrice = DefaultTokenRatio.
// Does not require real endpoints.
func TestGetTokenPriceWithOneDollarTokenRatioMode3(t *testing.T) {
	tokenPricer := NewClient("", "https://mainnet.infura.io/v3", 3)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	require.Equal(t, DefaultTokenRatio, ratio)
	t.Logf("ratio:%v", ratio)
}

func TestGetTokenPriceWithDefaultTokenRatioMode(t *testing.T) {
	rpc := rpcURL(t)
	tokenPricer := NewClient(cexURL(t), rpc, 3)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	require.Equal(t, DefaultTokenRatio, ratio)
	t.Logf("ratio:%v", ratio)
}

// TestGetTokenPriceWithNoSource tests fallback when both sources are invalid.
// Does not require real endpoints.
func TestGetTokenPriceWithNoSource(t *testing.T) {
	// source url are both invalid, so can not access correct prices
	tokenPricer := NewClient("https://api.bybit.co", "https://mainnet.infura.io/v3/4f4692085f1340c2a645ae04d36c232", 3)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	require.Equal(t, DefaultTokenRatio, ratio)
	t.Logf("ratio:%v", ratio)
}

func TestGetTokenPriceWithOnlySource1(t *testing.T) {
	// uniswapURL is invalid, so can not access correct prices from Uniswap
	tokenPricer := NewClient(cexURL(t), "https://mainnet.infura.io/v3/4f4692085f1340c2a645ae04d36c232", 3)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	t.Logf("ratio:%v", ratio)
}

func TestGetTokenPriceWithOnlySource2(t *testing.T) {
	rpc := rpcURL(t)
	// only uniswapURL is valid
	tokenPricer := NewClient("https://api.bybit.co", rpc, 3)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	t.Logf("ratio:%v", ratio)
}

func TestGetTokenPriceWithMNT(t *testing.T) {
	rpc := rpcURL(t)
	tokenPricer := NewClient(cexURL(t), rpc, 3)

	ratio, err := tokenPricer.tokenRatio()
	require.NoError(t, err)
	t.Logf("ratio:%v", ratio)
}

func Test_getMedian(t *testing.T) {
	result := getMedian([]float64{0, 0, 0})
	require.Equal(t, float64(0), result)

	result = getMedian([]float64{1.1, 0, 0})
	require.Equal(t, 1.1, result)

	result = getMedian([]float64{1.1, 2.1, 0})
	require.Equal(t, 1.6, result)

	result = getMedian([]float64{2.1, 1.1})
	require.Equal(t, 1.6, result)

	result = getMedian([]float64{1.1, 2.1, 3.1})
	require.Equal(t, 2.1, result)

	result = getMedian([]float64{1.1, 3.1, 2.1})
	require.Equal(t, 2.1, result)

	result = getMedian([]float64{1.1, 3.1, 2.1, 4.1})
	require.Equal(t, 2.6, result)
}

func Test_getMax(t *testing.T) {
	result := getMax(1.1, 2.1)
	require.Equal(t, 2.1, result)
}

func Test_getMin(t *testing.T) {
	result := getMin(1.1, 2.1)
	require.Equal(t, 1.1, result)
}
