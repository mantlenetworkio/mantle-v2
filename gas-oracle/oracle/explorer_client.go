package oracle

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/go-resty/resty/v2"
)

const (
	DefaultBlockscoutExplorerURL = "https://explorer.mantle.xyz"
	DefaultEtherscanExplorerURL  = "https://mantle.etherscan.io"
	BlockscoutStatsAPIHandler    = "/api/v2/stats"
	EtherscanStatsAPIHandler     = "/v2/api"
	DailyBlockCount              = 43200
)

type ExplorerClientInterface interface {
	DailyTxCount(ctx context.Context) (uint64, error)
	DailyTxCountFromUser(ctx context.Context) (uint64, error)
}

// BlockscoutStatsResponse represents the response from the Mantle explorer stats API
type BlockscoutStatsResponse struct {
	TransactionsToday string `json:"transactions_today"`
}

// EtherscanDailyTxResponse represents the response from Etherscan-style API
type EtherscanDailyTxResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []struct {
		UTCDate          string `json:"UTCDate"`
		UnixTimeStamp    string `json:"unixTimeStamp"`
		TransactionCount string `json:"transactionCount"`
	} `json:"result"`
}

// BlockscoutClient provides access to Mantle explorer APIs
type BlockscoutClient struct {
	client *resty.Client
}

// EtherscanClient provides access to Etherscan-style APIs
type EtherscanClient struct {
	client *resty.Client
	apiKey string
}

// NewBlockscoutClient creates a new blockscout client
func NewBlockscoutClient(baseURL string) *BlockscoutClient {
	if baseURL == "" {
		baseURL = DefaultBlockscoutExplorerURL
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(10*time.Second).
		SetHeader("accept", "application/json")

	return &BlockscoutClient{
		client: client,
	}
}

// NewEtherscanClient creates a new Etherscan-style client
func NewEtherscanClient(baseURL, apiKey string) *EtherscanClient {
	if baseURL == "" {
		baseURL = DefaultEtherscanExplorerURL
	}

	client := resty.New().
		SetBaseURL(baseURL).
		SetTimeout(10*time.Second).
		SetHeader("accept", "application/json")

	return &EtherscanClient{
		client: client,
		apiKey: apiKey,
	}
}

// DailyTxCount fetches the daily transaction count from the Mantle explorer API
func (e *BlockscoutClient) DailyTxCount(ctx context.Context) (uint64, error) {
	var stats BlockscoutStatsResponse

	resp, err := e.client.R().
		SetContext(ctx).
		SetResult(&stats).
		Get(BlockscoutStatsAPIHandler)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch transaction count: %w", err)
	}

	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("API request failed with status: %d", resp.StatusCode())
	}

	txCount, err := strconv.ParseUint(stats.TransactionsToday, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse transaction count: %w", err)
	}

	log.Debug("Fetched daily transaction count", "count", txCount, "url", resp.Request.URL)

	return txCount, nil
}

// DailyTxCount fetches the daily transaction count from the Etherscan-style API
func (e *EtherscanClient) DailyTxCount(ctx context.Context) (uint64, error) {
	// Get today's date in YYYY-MM-DD format
	today := time.Now().Format("2006-01-02")

	var response EtherscanDailyTxResponse

	resp, err := e.client.R().
		SetContext(ctx).
		SetResult(&response).
		SetQueryParams(map[string]string{
			"chainid":   "1",
			"module":    "stats",
			"action":    "dailytx",
			"startdate": today,
			"enddate":   today,
			"sort":      "asc",
			"apikey":    e.apiKey,
		}).
		Get(EtherscanStatsAPIHandler)

	if err != nil {
		return 0, fmt.Errorf("failed to fetch transaction count: %w", err)
	}

	if resp.StatusCode() != 200 {
		return 0, fmt.Errorf("API request failed with status: %d", resp.StatusCode())
	}

	if response.Status != "1" {
		return 0, fmt.Errorf("API returned error: %s", response.Message)
	}

	if len(response.Result) == 0 {
		return 0, fmt.Errorf("no transaction data found for today")
	}

	// Get the transaction count from the first (and only) result
	txCount, err := strconv.ParseUint(response.Result[0].TransactionCount, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse transaction count: %w", err)
	}

	log.Debug("Fetched daily transaction count from Etherscan API", "count", txCount, "url", resp.Request.URL)

	return txCount, nil
}

func (e *BlockscoutClient) DailyTxCountFromUser(ctx context.Context) (uint64, error) {
	txCount, err := e.DailyTxCount(ctx)
	if err != nil {
		return 0, err
	}

	if txCount <= DailyBlockCount {
		return 0, fmt.Errorf("transaction count is less than daily block count")
	}

	return txCount - DailyBlockCount, nil
}

func (e *EtherscanClient) DailyTxCountFromUser(ctx context.Context) (uint64, error) {
	txCount, err := e.DailyTxCount(ctx)
	if err != nil {
		return 0, err
	}

	if txCount <= DailyBlockCount {
		return 0, fmt.Errorf("transaction count is less than daily block count")
	}

	return txCount - DailyBlockCount, nil
}
