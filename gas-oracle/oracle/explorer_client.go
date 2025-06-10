package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/log"
)

const (
	DefaultBlockscoutExplorerURL = "https://explorer.mantle.xyz"
	StatsAPIHandler              = "/api/v2/stats"
	DailyBlockCount              = 43200
)

// ExplorerStats represents the response from the Mantle explorer stats API
type ExplorerStats struct {
	TransactionsToday string `json:"transactions_today"`
}

// ExplorerClient provides access to Mantle explorer APIs
type ExplorerClient struct {
	client  *http.Client
	baseURL string
}

// NewExplorerClient creates a new explorer client
func NewExplorerClient(baseURL string) *ExplorerClient {
	if baseURL == "" {
		baseURL = DefaultBlockscoutExplorerURL
	}
	return &ExplorerClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: baseURL,
	}
}

// DailyTxCount fetches the daily transaction count from the Mantle explorer API
func (e *ExplorerClient) DailyTxCount(ctx context.Context) (uint64, error) {
	apiURL, err := url.JoinPath(e.baseURL, StatsAPIHandler)
	if err != nil {
		return 0, fmt.Errorf("failed to construct API URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("accept", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch transaction count: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	var stats ExplorerStats
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	txCount, err := strconv.ParseUint(stats.TransactionsToday, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse transaction count: %w", err)
	}

	log.Debug("Fetched daily transaction count", "count", txCount, "url", apiURL)

	return txCount, nil
}

func (e *ExplorerClient) DailyTxCountFromUser(ctx context.Context) (uint64, error) {
	txCount, err := e.DailyTxCount(ctx)
	if err != nil {
		return 0, err
	}

	if txCount <= DailyBlockCount {
		return 0, fmt.Errorf("transaction count is less than daily block count")
	}

	return txCount - DailyBlockCount, nil
}
