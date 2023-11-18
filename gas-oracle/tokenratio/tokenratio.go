package tokenratio

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/go-resty/resty/v2"
)

// Client is an HTTP based TokenPriceClient
type Client struct {
	ctx context.Context

	client              *resty.Client
	uniswapQuoterClient *uniswapClient

	frequency time.Duration

	lastEthPrice float64
	lastMntPrice float64
	lastRatio    float64
	latestRatio  float64
}

var (
	errHTTPError = errors.New("http error")

	// DefaultTokenRatio is eth_price / mnt_price, 4000 = $1800/$0.45
	DefaultTokenRatio = float64(4000)
	// MaxTokenRatio token_ratio upper bounds
	MaxTokenRatio = float64(100000)
	// MinTokenRatio token_ratio lower bounds
	MinTokenRatio = float64(100)

	// DefaultETHPrice is default eth_price
	// If SwitchOneDollarTokenRatio valid, use DefaultETHPrice to set token_ratio to make mnt_price is 1$
	DefaultETHPrice = float64(1800)
	// ETHPriceMax eth_price upper bounds
	ETHPriceMax = float64(1000000)
	// ETHPriceMin eth_price lower bounds
	ETHPriceMin = float64(100)

	DefaultMNTPrice = 0.45
	// MNTPriceMax mnt_price upper bounds
	MNTPriceMax = float64(100)
	// MNTPriceMin mnt_price lower bounds
	MNTPriceMin = 0.01

	// ETHUSDT used to query eth/usdt price
	ETHUSDT = "ETHUSDT"
	// MNTUSDT used to query mnt/usdt price
	MNTUSDT = "MNTUSDT"
)

// NewClient create a new Client given a remote HTTP url, update frequency for token ratio
func NewClient(url, uniswapURL string, frequency uint64) *Client {
	client := resty.New()
	client.SetHostURL(url)
	client.OnAfterResponse(func(c *resty.Client, r *resty.Response) error {
		statusCode := r.StatusCode()
		if statusCode >= 400 {
			method := r.Request.Method
			url := r.Request.URL
			return fmt.Errorf("%d cannot %s %s: %w", statusCode, method, url, errHTTPError)
		}
		return nil
	})

	uniswapQuoterClient, err := newUniswapClient(uniswapURL)
	if err != nil {
		return nil
	}

	tokenRatioClient := &Client{
		ctx:                 context.Background(),
		client:              client,
		uniswapQuoterClient: uniswapQuoterClient,
		frequency:           time.Duration(frequency) * time.Second,
		lastRatio:           DefaultTokenRatio,
		latestRatio:         DefaultTokenRatio,
		lastEthPrice:        DefaultETHPrice,
		lastMntPrice:        DefaultMNTPrice,
	}

	go tokenRatioClient.loop()

	return tokenRatioClient
}

func (c *Client) loop() {
	timer := time.NewTicker(c.frequency)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			tokenRatio, err := c.tokenRatio()
			if err != nil {
				log.Error("token ratio", "tokenRatio", err)
				time.Sleep(c.frequency)
				continue
			}
			c.lastRatio = c.latestRatio
			c.latestRatio = tokenRatio
			log.Info("token ratio", "lastTokenRatio", c.lastRatio, "latestTokenRatio", c.latestRatio)
		case <-c.ctx.Done():
			break
		}
	}
}

func (c *Client) TokenRatio() float64 {
	return c.latestRatio
}

func (c *Client) tokenRatio() (float64, error) {
	// Todo query token prices concurrent
	var mntPrices, ethPrices []float64
	// get token price from oracle1(dex)
	mntPrice1, ethPrice1 := c.getTokenPricesFromUniswap()
	mntPrices = append(mntPrices, mntPrice1)
	ethPrices = append(ethPrices, ethPrice1)
	log.Info("query prices from oracle1", "mntPrice", mntPrice1, "ethPrice", ethPrice1)

	// get token price from oracle2(cex)
	mntPrice2, ethPrice2 := c.getTokenPricesFromCex()
	mntPrices = append(mntPrices, mntPrice2)
	ethPrices = append(ethPrices, ethPrice2)
	log.Info("query prices from oracle2", "mntPrice", mntPrice2, "ethPrice", ethPrice2)

	// get token price from oracle3(cex)
	// Todo add a third oracle to query prices
	mntPrice3, ethPrice3 := c.getTokenPricesFromCex()
	mntPrices = append(mntPrices, mntPrice3)
	ethPrices = append(ethPrices, ethPrice3)
	log.Info("query prices from oracle3", "mntPrice", mntPrice3, "ethPrice", ethPrice3)

	// median price for eth & mnt
	medianMNTPrice := getMedian(mntPrices)
	medianETHPrice := getMedian(ethPrices)

	// determine mnt_price, eth_price
	mntPrice := c.determineMNTPrice(medianMNTPrice)
	ethPrice := c.determineETHPrice(medianETHPrice)
	log.Info("prices after determine", "mntPrice", mntPrice, "ethPrice", ethPrice)

	// calculate ratio
	ratio := c.determineTokenRatio(mntPrice, ethPrice)

	c.lastRatio = ratio
	c.lastEthPrice = ethPrice
	c.lastMntPrice = mntPrice

	return ratio, nil
}

func (c *Client) getTokenPricesFromCex() (float64, float64) {
	ethPrice, err := c.queryV5(ETHUSDT)
	if err != nil {
		log.Warn("get token prices", "query ethPrice error", err)
		return 0, 0
	}
	mntPrice, err := c.queryV5(MNTUSDT)
	if err != nil {
		log.Warn("get token prices", "query mntPrice error", err)
		return 0, ethPrice
	}

	return mntPrice, ethPrice
}

func (c *Client) determineMNTPrice(price float64) float64 {
	if price > MNTPriceMax || price < MNTPriceMin {
		return c.lastMntPrice
	}

	return price
}

func (c *Client) determineETHPrice(price float64) float64 {
	if price > ETHPriceMax || price < ETHPriceMin {
		return c.lastEthPrice
	}

	return price
}

func (c *Client) determineTokenRatio(mntPrice, ethPrice float64) float64 {
	// calculate [tokenRatioMin, tokenRatioMax]
	tokenRatioMin := getMax(c.lastRatio*0.95, MinTokenRatio)
	tokenRatioMax := getMin(c.lastRatio*1.05, MaxTokenRatio)

	ratio := ethPrice / mntPrice
	if ratio <= tokenRatioMin {
		return tokenRatioMin
	}
	if ratio >= tokenRatioMax {
		return tokenRatioMax
	}

	return ratio
}

func getMedian(nums []float64) float64 {
	nonZeros := make([]float64, 0)
	for _, num := range nums {
		if num != 0 {
			nonZeros = append(nonZeros, num)
		}
	}
	sort.Float64s(nonZeros)
	if len(nonZeros) == 0 {
		return 0
	}
	return nonZeros[len(nonZeros)/2]
}

func getMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func getMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
