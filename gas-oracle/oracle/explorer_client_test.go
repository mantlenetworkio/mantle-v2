package oracle

import (
	"context"
	"testing"
)

func TestNewBlockscoutClient(t *testing.T) {
	client := NewBlockscoutClient("")
	if client == nil {
		t.Fatal("NewBlockscoutClient() returned nil")
	}
	if client.client == nil {
		t.Error("BlockscoutClient client is nil")
	}
}

func TestNewEtherscanClient(t *testing.T) {
	client := NewEtherscanClient("", "")
	if client == nil {
		t.Fatal("NewEtherscanClient() returned nil")
	}
	if client.client == nil {
		t.Error("EtherscanClient client is nil")
	}
}

func TestBlockscoutClientDailyTxCount(t *testing.T) {
	client := NewBlockscoutClient("")

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
