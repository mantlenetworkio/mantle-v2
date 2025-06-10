package oracle

import (
	"context"
	"testing"
)

func TestNewExplorerClient(t *testing.T) {
	client := NewExplorerClient("")
	if client == nil {
		t.Fatal("NewExplorerClient() returned nil")
	}
	if client.client == nil {
		t.Error("ExplorerClient client is nil")
	}
	if client.baseURL != DefaultBlockscoutExplorerURL {
		t.Errorf("Expected baseURL %s, got %s", DefaultBlockscoutExplorerURL, client.baseURL)
	}
}

func TestNewExplorerClientWithCustomURL(t *testing.T) {
	customURL := "https://custom-explorer.com"
	client := NewExplorerClient(customURL)
	if client == nil {
		t.Fatal("NewExplorerClient() returned nil")
	}
	if client.baseURL != customURL {
		t.Errorf("Expected baseURL %s, got %s", customURL, client.baseURL)
	}
}

func TestExplorerClientDailyTxCount(t *testing.T) {
	client := NewExplorerClient("")

	// Test the actual API call
	txCount, err := client.DailyTxCount(context.Background())
	if err != nil {
		t.Logf("DailyTxCount() returned error (this might be expected if API is not available): %v", err)
		// Don't fail the test if the API is not available
		return
	}

	if txCount == 0 {
		t.Log("DailyTxCount() returned 0 (this might be expected)")
	} else {
		t.Logf("DailyTxCount() returned: %d", txCount)
	}
}
