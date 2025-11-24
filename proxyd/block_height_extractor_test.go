package proxyd

import (
	"encoding/json"
	"math"
	"testing"
)

func TestExtractBlockHeight_HeightBasedMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		params   string
		expected uint64
		success  bool
	}{
		{
			name:     "eth_getBlockByNumber with hex",
			method:   "eth_getBlockByNumber",
			params:   `["0x77359400", true]`,
			expected: 2000000000,
			success:  true,
		},
		{
			name:     "eth_getBlockByNumber with latest",
			method:   "eth_getBlockByNumber",
			params:   `["latest", true]`,
			expected: math.MaxUint64,
			success:  true,
		},
		{
			name:     "eth_getBlockByNumber with earliest",
			method:   "eth_getBlockByNumber",
			params:   `["earliest", true]`,
			expected: 0,
			success:  true,
		},
		{
			name:     "eth_getBalance with address and hex block",
			method:   "eth_getBalance",
			params:   `["0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "0x5F5E100"]`,
			expected: 100000000,
			success:  true,
		},
		{
			name:     "eth_getBalance with latest",
			method:   "eth_getBalance",
			params:   `["0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "latest"]`,
			expected: math.MaxUint64,
			success:  true,
		},
		{
			name:     "eth_getCode with address and block",
			method:   "eth_getCode",
			params:   `["0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "0x1"]`,
			expected: 1,
			success:  true,
		},
		{
			name:     "eth_getStorageAt with address, position and block",
			method:   "eth_getStorageAt",
			params:   `["0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", "0x0", "0x100"]`,
			expected: 256,
			success:  true,
		},
		{
			name:     "eth_call with transaction object and block",
			method:   "eth_call",
			params:   `[{"to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"}, "0xA"]`,
			expected: 10,
			success:  true,
		},
		{
			name:     "eth_call with missing block parameter (defaults to latest)",
			method:   "eth_call",
			params:   `[{"to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"}]`,
			expected: math.MaxUint64,
			success:  true,
		},
		{
			name:     "eth_getBlockTransactionCountByNumber",
			method:   "eth_getBlockTransactionCountByNumber",
			params:   `["0x64"]`,
			expected: 100,
			success:  true,
		},
	}

	extractor := NewBlockHeightExtractor(125000000, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RPCReq{
				Method: tt.method,
				Params: json.RawMessage(tt.params),
			}

		result := extractor.ExtractBlockHeight(req)

		if tt.success {
			// expect successful height extraction
			if result.Type != BlockParamHeight {
				t.Errorf("Expected type=BlockParamHeight, got %v", result.Type)
			}
			if result.Height != tt.expected {
				t.Errorf("Expected height=%d, got %d", tt.expected, result.Height)
			}
		} else {
			// expect failure, should return Unknown
			if result.Type != BlockParamUnknown {
				t.Errorf("Expected type=BlockParamUnknown, got %v", result.Type)
			}
		}
		})
	}
}

func TestExtractBlockHeight_HashBasedMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
		params string
	}{
		{
			name:   "eth_getBlockByHash",
			method: "eth_getBlockByHash",
			params: `["0xabc123", true]`,
		},
		{
			name:   "eth_getTransactionByHash",
			method: "eth_getTransactionByHash",
			params: `["0xabc123"]`,
		},
		{
			name:   "eth_getTransactionReceipt",
			method: "eth_getTransactionReceipt",
			params: `["0xabc123"]`,
		},
		{
			name:   "eth_getTxStatusByHash",
			method: "eth_getTxStatusByHash",
			params: `["0xabc123"]`,
		},
	}

	extractor := NewBlockHeightExtractor(125000000, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RPCReq{
				Method: tt.method,
				Params: json.RawMessage(tt.params),
			}

		result := extractor.ExtractBlockHeight(req)

		// Hash methods should return Unknown (use fallback chain)
		if result.Type != BlockParamUnknown {
			t.Errorf("Expected type=BlockParamUnknown for hash-based method, got %v", result.Type)
		}
		})
	}
}

func TestExtractBlockHeight_RangeQueryMethods(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		params    string
		fromBlock uint64
		toBlock   uint64
		success   bool
	}{
		{
			name:      "eth_getLogs with hex range",
			method:    "eth_getLogs",
			params:    `[{"fromBlock": "0x64", "toBlock": "0xC8"}]`,
			fromBlock: 100,
			toBlock:   200,
			success:   true,
		},
		{
			name:      "eth_getLogs with latest and earliest",
			method:    "eth_getLogs",
			params:    `[{"fromBlock": "earliest", "toBlock": "latest"}]`,
			fromBlock: 0,
			toBlock:   math.MaxUint64,
			success:   true,
		},
		{
			name:      "eth_getLogs with only toBlock",
			method:    "eth_getLogs",
			params:    `[{"toBlock": "0x100"}]`,
			fromBlock: 0,
			toBlock:   256,
			success:   true,
		},
		{
			name:      "eth_newFilter with range",
			method:    "eth_newFilter",
			params:    `[{"fromBlock": "0x1", "toBlock": "0x1000"}]`,
			fromBlock: 1,
			toBlock:   4096,
			success:   true,
		},
		{
			name:      "eth_getBlockRange",
			method:    "eth_getBlockRange",
			params:    `["0xA", "0x14", true]`,
			fromBlock: 10,
			toBlock:   20,
			success:   true,
		},
	}

	extractor := NewBlockHeightExtractor(125000000, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RPCReq{
				Method: tt.method,
				Params: json.RawMessage(tt.params),
			}

		result := extractor.ExtractBlockHeight(req)

		if tt.success {
			// expect successful range extraction
			if result.Type != BlockParamRange {
				t.Errorf("Expected type=BlockParamRange, got %v", result.Type)
			}
			if result.FromBlock != tt.fromBlock {
				t.Errorf("Expected fromBlock=%d, got %d", tt.fromBlock, result.FromBlock)
			}
			if result.ToBlock != tt.toBlock {
				t.Errorf("Expected toBlock=%d, got %d", tt.toBlock, result.ToBlock)
			}
		} else {
			// expect failure, should return Unknown
			if result.Type != BlockParamUnknown {
				t.Errorf("Expected type=BlockParamUnknown, got %v", result.Type)
			}
		}
		})
	}
}

func TestExtractBlockHeight_RealtimeMethods(t *testing.T) {
	tests := []struct {
		name   string
		method string
		params string
	}{
		{
			name:   "eth_blockNumber",
			method: "eth_blockNumber",
			params: `[]`,
		},
		{
			name:   "eth_chainId",
			method: "eth_chainId",
			params: `[]`,
		},
		{
			name:   "eth_gasPrice",
			method: "eth_gasPrice",
			params: `[]`,
		},
		{
			name:   "eth_sendRawTransaction",
			method: "eth_sendRawTransaction",
			params: `["0xabcd"]`,
		},
	}

	extractor := NewBlockHeightExtractor(125000000, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RPCReq{
				Method: tt.method,
				Params: json.RawMessage(tt.params),
			}

		result := extractor.ExtractBlockHeight(req)

		// Real-time methods should return Unknown
		// They should be routed via rpc_method_mappings to the appropriate group
		if result.Type != BlockParamUnknown {
			t.Errorf("Expected type=BlockParamUnknown for realtime method, got %v", result.Type)
		}
	})
	}
}

// TestDetermineRoutingStrategy removed
// Routing strategy is now determined by BackendRouter, tests in backend_router_test.go

func TestParseBlockNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected uint64
		success  bool
	}{
		{
			name:     "Hex number",
			input:    `"0x100"`,
			expected: 256,
			success:  true,
		},
		{
			name:     "Decimal number as string (should fail)",
			input:    `"100"`,
			expected: 0,
			success:  false,
		},
		{
			name:     "latest tag",
			input:    `"latest"`,
			expected: math.MaxUint64,
			success:  true,
		},
		{
			name:     "earliest tag",
			input:    `"earliest"`,
			expected: 0,
			success:  true,
		},
		{
			name:     "pending tag",
			input:    `"pending"`,
			expected: math.MaxUint64,
			success:  true,
		},
		{
			name:     "safe tag",
			input:    `"safe"`,
			expected: math.MaxUint64,
			success:  true,
		},
		{
			name:     "finalized tag",
			input:    `"finalized"`,
			expected: math.MaxUint64,
			success:  true,
		},
		{
			name:     "Direct number (should fail)",
			input:    `256`,
			expected: 0,
			success:  false,
		},
		{
			name:     "Block hash (should fail)",
			input:    `"0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"`,
			expected: 0,
			success:  false,
		},
		{
			name:     "Invalid string",
			input:    `"invalid"`,
			expected: 0,
			success:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			height, ok := parseBlockNumber(json.RawMessage(tt.input))

			if ok != tt.success {
				t.Errorf("Expected success=%v, got %v", tt.success, ok)
			}

			if ok && height != tt.expected {
				t.Errorf("Expected height=%d, got %d", tt.expected, height)
			}
		})
	}
}

// TestRoutingStrategyString removed
// RoutingStrategy type has been refactored to BlockParamType

// TestCustomMethodIndices tests custom method index configuration
func TestCustomMethodIndices(t *testing.T) {
	// Custom configuration: specify indices for custom methods
	customIndices := map[string]int{
		"custom_getDataAtBlock": 1, // second parameter is block number
		"custom_queryByHeight":  0, // first parameter is block number
		"eth_getBalance":        2, // override default config (default is 1)
	}

	extractor := NewBlockHeightExtractor(125000000, customIndices)

	tests := []struct {
		name     string
		method   string
		params   string
		wantType BlockParamType
		wantVal  uint64
	}{
		{
			name:     "custom method with index 1",
			method:   "custom_getDataAtBlock",
			params:   `["0xcontract", "0x7735940"]`,
			wantType: BlockParamHeight,
			wantVal:  125000000,
		},
		{
			name:     "custom method with index 0",
			method:   "custom_queryByHeight",
			params:   `["0x7735940", "0xdata"]`,
			wantType: BlockParamHeight,
			wantVal:  125000000,
		},
		{
			name:     "override default: eth_getBalance now uses index 2",
			method:   "eth_getBalance",
			params:   `["0xaddr", "extra_param", "0x7735940"]`,
			wantType: BlockParamHeight,
			wantVal:  125000000,
		},
		{
			name:     "override default: eth_getBalance with old index fails",
			method:   "eth_getBalance",
			params:   `["0xaddr", "0x7735940"]`, // only 2 params, but config requires index 2
			wantType: BlockParamHeight,
			wantVal:  math.MaxUint64, // missing parameter, returns latest
		},
		{
			name:     "unconfigured method returns Unknown",
			method:   "custom_unknownMethod",
			params:   `["0x123"]`,
			wantType: BlockParamUnknown,
		},
		{
			name:     "default method still works",
			method:   "eth_getBlockByNumber",
			params:   `["0x7735940", true]`,
			wantType: BlockParamHeight,
			wantVal:  125000000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &RPCReq{
				Method: tt.method,
				Params: json.RawMessage(tt.params),
			}

			result := extractor.ExtractBlockHeight(req)

			if result.Type != tt.wantType {
				t.Errorf("Type: want %v, got %v", tt.wantType, result.Type)
			}

			if tt.wantType == BlockParamHeight && result.Height != tt.wantVal {
				t.Errorf("Height: want %d, got %d", tt.wantVal, result.Height)
			}
		})
	}
}
