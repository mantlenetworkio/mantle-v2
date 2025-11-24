package proxyd

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// BlockHeightExtractor extracts block height information from RPC requests
type BlockHeightExtractor struct {
	cutoffHeight  uint64
	customIndices map[string]int // custom method parameter index overrides
}

// NewBlockHeightExtractor creates a new block height extractor
func NewBlockHeightExtractor(cutoffHeight uint64, customIndices map[string]int) *BlockHeightExtractor {
	return &BlockHeightExtractor{
		cutoffHeight:  cutoffHeight,
		customIndices: customIndices,
	}
}

// BlockParamType represents the type of block parameter
type BlockParamType int

const (
	BlockParamHeight  BlockParamType = iota // single block height
	BlockParamRange                         // block range query
	BlockParamUnknown                       // unknown (use fallback chain)
)

func (t BlockParamType) String() string {
	switch t {
	case BlockParamHeight:
		return "height"
	case BlockParamRange:
		return "range"
	case BlockParamUnknown:
		return "unknown"
	default:
		return "invalid"
	}
}

// ExtractionResult holds extracted block information
type ExtractionResult struct {
	Type      BlockParamType
	Height    uint64 // valid when Type == BlockParamHeight
	FromBlock uint64 // valid when Type == BlockParamRange
	ToBlock   uint64 // valid when Type == BlockParamRange
}

// HeightResult creates a height extraction result
func HeightResult(height uint64) *ExtractionResult {
	return &ExtractionResult{
		Type:   BlockParamHeight,
		Height: height,
	}
}

// RangeResult creates a range extraction result
func RangeResult(from, to uint64) *ExtractionResult {
	return &ExtractionResult{
		Type:      BlockParamRange,
		FromBlock: from,
		ToBlock:   to,
	}
}

// UnknownResult creates an unknown result (fallback chain)
func UnknownResult() *ExtractionResult {
	return &ExtractionResult{
		Type: BlockParamUnknown,
	}
}

// blockParamIndices maps method names to their block parameter index
var blockParamIndices = map[string]int{
	// Block queries - first parameter is block number
	"eth_getBlockByNumber":                       0,
	"eth_getBlockTransactionCountByNumber":       0,
	"eth_getHeaderByNumber":                      0,
	"eth_getUncleByBlockNumberAndIndex":          0,
	"eth_getUncleCountByBlockNumber":             0,
	"eth_getTransactionByBlockNumberAndIndex":    0,
	"eth_getRawTransactionByBlockNumberAndIndex": 0,

	// Account state queries - second parameter is block number
	"eth_getBalance":          1,
	"eth_getCode":             1,
	"eth_getStorageAt":        2,
	"eth_getTransactionCount": 1,
	"eth_getProof":            2,

	// Contract calls - second parameter is block number
	"eth_call":        1,
	"eth_estimateGas": 1,

	// Chain info - second parameter is block number
	"eth_feeHistory": 1,
}

// ExtractBlockHeight extracts block information from an RPC request
func (e *BlockHeightExtractor) ExtractBlockHeight(req *RPCReq) *ExtractionResult {
	method := req.Method

	// Handle range queries specially
	switch method {
	case "eth_getLogs", "eth_newFilter":
		return e.extractFromLogFilter(req.Params)
	case "eth_getBlockRange":
		return e.extractFromBlockRange(req.Params)
	}

	// Check for block height methods
	_, hasCustom := e.customIndices[method]
	_, hasDefault := blockParamIndices[method]

	if hasCustom || hasDefault {
		return e.extractFromHeightMethod(method, req.Params)
	}

	// Hash queries, real-time queries, unknown methods -> fallback chain
	return UnknownResult()
}

// extractFromHeightMethod extracts block height from height-based methods
func (e *BlockHeightExtractor) extractFromHeightMethod(method string, params json.RawMessage) *ExtractionResult {
	var paramList []json.RawMessage
	if err := json.Unmarshal(params, &paramList); err != nil {
		return UnknownResult()
	}

	// Get block parameter index (custom overrides default)
	blockParamIndex, ok := e.customIndices[method]
	if !ok {
		blockParamIndex, ok = blockParamIndices[method]
		if !ok {
			return UnknownResult()
		}
	}

	if blockParamIndex >= len(paramList) {
		// Missing parameter defaults to "latest"
		return HeightResult(math.MaxUint64)
	}

	height, ok := parseBlockNumber(paramList[blockParamIndex])
	if !ok {
		return UnknownResult()
	}

	return HeightResult(height)
}

// parseBlockNumber parses a block number parameter
func parseBlockNumber(param json.RawMessage) (uint64, bool) {
	var str string
	if err := json.Unmarshal(param, &str); err != nil {
		return 0, false
	}

	// Handle block tags
	switch strings.ToLower(str) {
	case "earliest":
		return 0, true
	case "latest", "pending", "safe", "finalized", "unsafe":
		return math.MaxUint64, true
	}

	// Parse hex number
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		if i, err := strconv.ParseUint(str[2:], 16, 64); err == nil {
			return i, true
		}
	}

	return 0, false
}

// extractFromLogFilter extracts block range from log filter
func (e *BlockHeightExtractor) extractFromLogFilter(params json.RawMessage) *ExtractionResult {
	var paramList []json.RawMessage
	if err := json.Unmarshal(params, &paramList); err != nil {
		return UnknownResult()
	}

	if len(paramList) == 0 {
		return UnknownResult()
	}

	var filter struct {
		FromBlock interface{} `json:"fromBlock"`
		ToBlock   interface{} `json:"toBlock"`
	}

	if err := json.Unmarshal(paramList[0], &filter); err != nil {
		return UnknownResult()
	}

	fromBlock := uint64(0) // defaults to earliest
	if filter.FromBlock != nil {
		if fromBlockJSON, _ := json.Marshal(filter.FromBlock); fromBlockJSON != nil {
			if fb, ok := parseBlockNumber(fromBlockJSON); ok {
				fromBlock = fb
			}
		}
	}

	toBlock := uint64(math.MaxUint64) // defaults to latest
	if filter.ToBlock != nil {
		if toBlockJSON, _ := json.Marshal(filter.ToBlock); toBlockJSON != nil {
			if tb, ok := parseBlockNumber(toBlockJSON); ok {
				toBlock = tb
			}
		}
	}

	return RangeResult(fromBlock, toBlock)
}

// extractFromBlockRange extracts block range from eth_getBlockRange
func (e *BlockHeightExtractor) extractFromBlockRange(params json.RawMessage) *ExtractionResult {
	var paramList []json.RawMessage
	if err := json.Unmarshal(params, &paramList); err != nil {
		return UnknownResult()
	}

	if len(paramList) < 2 {
		return UnknownResult()
	}

	fromBlock, okFrom := parseBlockNumber(paramList[0])
	toBlock, okTo := parseBlockNumber(paramList[1])

	if !okFrom || !okTo {
		return UnknownResult()
	}

	return RangeResult(fromBlock, toBlock)
}
