package proxyd

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ethereum/go-ethereum/log"
)

// BackendRouter routes RPC requests based on block height
type BackendRouter struct {
	extractor       *BlockHeightExtractor
	primaryBackend  *Backend
	fallbackBackend *Backend
	cutoffHeight    uint64
}

// NewBackendRouter creates a new backend router
func NewBackendRouter(
	cutoffHeight uint64,
	primaryBackend *Backend,
	fallbackBackend *Backend,
	customIndices map[string]int,
) *BackendRouter {
	return &BackendRouter{
		extractor:       NewBlockHeightExtractor(cutoffHeight, customIndices),
		primaryBackend:  primaryBackend,
		fallbackBackend: fallbackBackend,
		cutoffHeight:    cutoffHeight,
	}
}

// OrderBackends determines backend order for single or batch requests
// Returns:
//   - [primary]            - route to primary only
//   - [fallback]           - route to fallback only
//   - [primary, fallback]  - try primary first, fallback on null result
func (r *BackendRouter) OrderBackends(ctx context.Context, reqs []*RPCReq) []*Backend {
	if len(reqs) == 0 {
		return []*Backend{r.primaryBackend, r.fallbackBackend}
	}

	needsPrimary, needsFallback := false, false

	for _, req := range reqs {
		if req == nil {
			continue
		}

		result := r.extractor.ExtractBlockHeight(req)

		switch result.Type {
		case BlockParamHeight:
			if result.Height >= r.cutoffHeight {
				needsPrimary = true
			} else {
				needsFallback = true
			}

		case BlockParamRange:
			backends := r.handleRangeQueryBackends(ctx, result)
			if len(backends) == 1 {
				if backends[0] == r.primaryBackend {
					needsPrimary = true
				} else {
					needsFallback = true
				}
			} else {
				needsPrimary, needsFallback = true, true
			}

		default:
			// Unknown params: use fallback chain
			needsPrimary, needsFallback = true, true
		}

		if needsPrimary && needsFallback {
			break
		}
	}

	if needsPrimary && needsFallback {
		log.Debug("Route with fallback chain", "req_id", GetReqID(ctx))
		return []*Backend{r.primaryBackend, r.fallbackBackend}
	}

	if needsPrimary {
		log.Debug("Route to primary only", "req_id", GetReqID(ctx))
		return []*Backend{r.primaryBackend}
	}

	log.Debug("Route to fallback only", "req_id", GetReqID(ctx))
	return []*Backend{r.fallbackBackend}
}

// handleRangeQueryBackends selects backends for range queries
func (r *BackendRouter) handleRangeQueryBackends(ctx context.Context, result *ExtractionResult) []*Backend {
	cutoff := r.cutoffHeight

	// Range entirely before cutoff -> fallback (archive) only
	if result.ToBlock < cutoff {
		log.Debug(
			"Range query entirely before cutoff, route to fallback only",
			"from", result.FromBlock,
			"to", result.ToBlock,
			"cutoff", cutoff,
			"req_id", GetReqID(ctx),
		)
		return []*Backend{r.fallbackBackend}
	}

	// Range entirely after cutoff -> primary only
	if result.FromBlock >= cutoff {
		log.Debug(
			"Range query entirely after cutoff, route to primary only",
			"from", result.FromBlock,
			"to", result.ToBlock,
			"cutoff", cutoff,
			"req_id", GetReqID(ctx),
		)
		return []*Backend{r.primaryBackend}
	}

	// Range spans cutoff -> fallback has full data
	log.Info(
		"Range query spans cutoff, using fallback with full data",
		"from", result.FromBlock,
		"to", result.ToBlock,
		"cutoff", cutoff,
		"req_id", GetReqID(ctx),
	)
	return []*Backend{r.fallbackBackend}
}

// ShouldFallbackOnNull checks if responses are null and should trigger fallback
func (r *BackendRouter) ShouldFallbackOnNull(responses []*RPCRes) bool {
	return isNullResult(responses)
}

// isNullResult checks if all responses are null
//
// Based on testing: Reth returns {"result": null} when data doesn't exist,
// not error messages. Any error response blocks fallback.
func isNullResult(responses []*RPCRes) bool {
	if len(responses) == 0 {
		return false
	}

	for _, res := range responses {
		// Any error response blocks fallback
		if res.Error != nil {
			return false
		}

		// Check if result is non-null
		if res.Result != nil {
			resultBytes, err := resultToBytes(res.Result)
			if err != nil {
				return false
			}
			if result := strings.TrimSpace(string(resultBytes)); result != "" && result != "null" {
				return false
			}
		}
	}

	return true
}

// resultToBytes converts Result to bytes for null checking
func resultToBytes(result interface{}) ([]byte, error) {
	if result == nil {
		return []byte("null"), nil
	}

	switch v := result.(type) {
	case []byte:
		return v, nil
	case json.RawMessage:
		return []byte(v), nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(v)
	}
}
