package proxyd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// mustJSON converts any value to json.RawMessage, panics on error
func mustJSON(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// createTestBackend creates a backend for testing
func createTestBackend(name string) *Backend {
	return &Backend{
		Name: name,
	}
}

func TestOrderBackends(t *testing.T) {
	primary := createTestBackend("primary")
	fallback := createTestBackend("fallback")
	router := NewBackendRouter(125000000, primary, fallback, nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		reqs     []*RPCReq
		expected int // expected number of backends
		wantName string
	}{
		{
			name:     "empty batch",
			reqs:     []*RPCReq{},
			expected: 2, // fallback chain
			wantName: "primary",
		},
		{
			name: "single request - recent block",
			reqs: []*RPCReq{
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x77359400", false})}, // 2000000000
			},
			expected: 1,
			wantName: "primary",
		},
		{
			name: "single request - historical block",
			reqs: []*RPCReq{
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x1", false})}, // block 1
			},
			expected: 1,
			wantName: "fallback",
		},
		{
			name: "all primary requests",
			reqs: []*RPCReq{
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x77359400", false})}, // 2000000000
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x77359401", false})}, // 2000000001
				{Method: "eth_getBalance", Params: mustJSON([]interface{}{"0x123", "0x77359400"})},     // recent block
			},
			expected: 1,
			wantName: "primary",
		},
		{
			name: "all fallback requests",
			reqs: []*RPCReq{
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x1", false})},  // block 1
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x64", false})}, // block 100
			},
			expected: 1,
			wantName: "fallback",
		},
		{
			name: "mixed requests - primary and fallback",
			reqs: []*RPCReq{
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x77359400", false})}, // recent
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x1", false})},        // historical
			},
			expected: 2, // fallback chain
			wantName: "primary",
		},
		{
			name: "mixed with hash query",
			reqs: []*RPCReq{
				{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x77359400", false})}, // recent
				{Method: "eth_getTransactionByHash", Params: mustJSON([]interface{}{"0xabc"})},         // hash (uncertain)
			},
			expected: 2, // fallback chain
			wantName: "primary",
		},
		{
			name: "all hash queries",
			reqs: []*RPCReq{
				{Method: "eth_getTransactionByHash", Params: mustJSON([]interface{}{"0xabc"})},
				{Method: "eth_getTransactionReceipt", Params: mustJSON([]interface{}{"0xdef"})},
			},
			expected: 2, // fallback chain
			wantName: "primary",
		},
		{
			name: "range query - before cutoff",
			reqs: []*RPCReq{
				{Method: "eth_getLogs", Params: mustJSON([]interface{}{map[string]interface{}{
					"fromBlock": "0x1",
					"toBlock":   "0x64",
				}})},
			},
			expected: 1,
			wantName: "fallback",
		},
		{
			name: "range query - after cutoff",
			reqs: []*RPCReq{
				{Method: "eth_getLogs", Params: mustJSON([]interface{}{map[string]interface{}{
					"fromBlock": "0x77359400",
					"toBlock":   "0x77359500",
				}})},
			},
			expected: 1,
			wantName: "primary",
		},
		{
			name: "range query - spans cutoff",
			reqs: []*RPCReq{
				{Method: "eth_getLogs", Params: mustJSON([]interface{}{map[string]interface{}{
					"fromBlock": "0x1",
					"toBlock":   "0x77359400",
				}})},
			},
			expected: 1,
			wantName: "fallback", // geth has full data
		},
		{
			name: "batch with mixed range queries",
			reqs: []*RPCReq{
				{Method: "eth_getLogs", Params: mustJSON([]interface{}{map[string]interface{}{
					"fromBlock": "0x1",
					"toBlock":   "0x64",
				}})},
				{Method: "eth_getLogs", Params: mustJSON([]interface{}{map[string]interface{}{
					"fromBlock": "0x77359400",
					"toBlock":   "0x77359500",
				}})},
			},
			expected: 2, // mixed, use fallback chain
			wantName: "primary",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backends := router.OrderBackends(ctx, tt.reqs)
			require.Equal(t, tt.expected, len(backends), "unexpected number of backends")
			require.Equal(t, tt.wantName, backends[0].Name, "unexpected first backend")
		})
	}
}

func TestOrderBackendsEdgeCases(t *testing.T) {
	primary := createTestBackend("primary")
	fallback := createTestBackend("fallback")
	router := NewBackendRouter(125000000, primary, fallback, nil)
	ctx := context.Background()

	t.Run("nil request in batch", func(t *testing.T) {
		reqs := []*RPCReq{
			{Method: "eth_blockNumber", Params: mustJSON([]interface{}{})},
			nil,
			{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x1", false})},
		}

		backends := router.OrderBackends(ctx, reqs)
		require.NotNil(t, backends)
		require.Greater(t, len(backends), 0)
	})

	t.Run("request with invalid params", func(t *testing.T) {
		reqs := []*RPCReq{
			{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"invalid"})},
		}

		backends := router.OrderBackends(ctx, reqs)
		require.NotNil(t, backends)
		require.Equal(t, 2, len(backends)) // falls back to chain
	})
}

// BenchmarkOrderBackends ensures routing overhead is minimal
func BenchmarkOrderBackends(b *testing.B) {
	primary := createTestBackend("primary")
	fallback := createTestBackend("fallback")
	router := NewBackendRouter(125000000, primary, fallback, nil)
	ctx := context.Background()

	reqs := []*RPCReq{
		{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x77359400", false})},
		{Method: "eth_getTransactionByHash", Params: mustJSON([]interface{}{"0xabc"})},
		{Method: "eth_blockNumber", Params: mustJSON([]interface{}{})},
		{Method: "eth_getLogs", Params: mustJSON([]interface{}{map[string]interface{}{
			"fromBlock": "0x77359400",
			"toBlock":   "0x77359500",
		}})},
		{Method: "eth_getBlockByNumber", Params: mustJSON([]interface{}{"0x1", false})},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.OrderBackends(ctx, reqs)
	}
}
