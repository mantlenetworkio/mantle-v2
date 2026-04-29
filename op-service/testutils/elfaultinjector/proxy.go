// Package elfaultinjector provides a JSON-RPC aware WebSocket proxy that sits
// in front of an Engine API endpoint (op-geth's auth RPC) and synthesizes
// PayloadStatusV1{Status: INVALID} responses for matching engine_newPayloadV3
// or engine_newPayloadV4 requests.
//
// The proxy is intended for tests that need to reproduce execution-layer
// divergence (e.g. the op-conductor split-brain at unsafe head case study,
// where a leader's payload is accepted by the leader's own EL but rejected
// by other ELs in the cluster). All non-matching JSON-RPC traffic — including
// other Engine API methods, getPayload, forkchoiceUpdated, eth_*, etc. —
// passes through unmodified.
//
// The proxy terminates the inbound WebSocket connection from the client
// (op-node) and dials the upstream EL as a WebSocket client, forwarding the
// inbound Authorization header so the upstream's JWT check still runs against
// the original token. The proxy itself does NOT validate the JWT; it relies
// on the upstream for authentication. Per the Engine API authentication spec,
// the JWT is sent only on the WebSocket upgrade handshake — once the
// connection is upgraded, individual JSON-RPC frames are unauthenticated.
package elfaultinjector

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// Rule controls when the proxy synthesizes an INVALID response in place of
// forwarding an engine_newPayloadV{3,4} request to upstream.
type Rule struct {
	// RejectFromBlock, if non-zero, causes any newPayload whose
	// executionPayload.blockNumber is >= RejectFromBlock to be rejected.
	RejectFromBlock uint64

	// RejectTxHashPrefix, if non-empty, causes any newPayload whose payload
	// includes a transaction whose raw bytes start with this prefix to be
	// rejected. Useful for triggering rejection on a specific tx.
	RejectTxHashPrefix []byte

	// MaxInjections, if non-zero, caps the total number of injected
	// rejections this proxy will synthesize before reverting to pure
	// pass-through. Useful for "reject the next N payloads" scenarios.
	MaxInjections int
}

// matches returns true if the given execution payload should be rejected
// under this rule.
func (r Rule) matches(payload *jsonExecPayload) bool {
	if r.RejectFromBlock > 0 {
		blockNum, err := hexToUint64(payload.BlockNumber)
		if err == nil && blockNum >= r.RejectFromBlock {
			return true
		}
	}
	if len(r.RejectTxHashPrefix) > 0 {
		for _, tx := range payload.Transactions {
			raw, err := hexToBytes(tx)
			if err != nil {
				continue
			}
			if len(raw) >= len(r.RejectTxHashPrefix) &&
				bytesEqual(raw[:len(r.RejectTxHashPrefix)], r.RejectTxHashPrefix) {
				return true
			}
		}
	}
	return false
}

// Proxy is an Engine API fault-injection proxy.
type Proxy struct {
	mu       sync.RWMutex
	upstream string // host:port of the upstream Engine API WebSocket
	listener net.Listener
	logger   log.Logger
	rule     *Rule // nil means inactive
	injected atomic.Int64
	stopped  atomic.Bool
	wg       sync.WaitGroup
	upgrader websocket.Upgrader
	dialer   *websocket.Dialer
}

// New constructs a fault-injection proxy. The proxy is inactive until
// Activate is called.
//
// jwtSecret is reserved for future JWT validation; the current implementation
// forwards the inbound Authorization header to the upstream verbatim, so the
// upstream still gets to validate the JWT. The parameter is kept in the
// signature so callers don't have to change when JWT validation moves into
// the proxy.
func New(lgr log.Logger, _jwtSecret [32]byte) *Proxy {
	return &Proxy{
		logger: lgr,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: 10 * time.Second,
			CheckOrigin:      func(*http.Request) bool { return true },
		},
		dialer: &websocket.Dialer{
			HandshakeTimeout: 10 * time.Second,
		},
	}
}

// Start begins listening on a random localhost port. Use Addr to retrieve
// the listening address.
func (p *Proxy) Start() error {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("elfaultinjector: listen: %w", err)
	}
	p.listener = lis
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		srv := &http.Server{Handler: http.HandlerFunc(p.serve)}
		_ = srv.Serve(lis)
	}()
	return nil
}

// Addr returns the proxy's listening address as host:port.
func (p *Proxy) Addr() string {
	return p.listener.Addr().String()
}

// SetUpstream configures the upstream Engine API WebSocket address. The
// address may be an absolute ws:// URL or a bare host:port. Subsequent
// connections are dialed against this address.
func (p *Proxy) SetUpstream(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.upstream = addr
}

// Activate enables fault injection with the given rule. Subsequent matching
// engine_newPayloadV{3,4} requests are intercepted.
func (p *Proxy) Activate(r Rule) {
	p.mu.Lock()
	defer p.mu.Unlock()
	rc := r
	p.rule = &rc
	p.injected.Store(0)
	p.logger.Info("fault injection activated",
		"rejectFromBlock", r.RejectFromBlock,
		"hasTxPrefix", len(r.RejectTxHashPrefix) > 0,
		"maxInjections", r.MaxInjections)
}

// Deactivate disables fault injection. The proxy reverts to pure pass-through.
func (p *Proxy) Deactivate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.rule = nil
	p.logger.Info("fault injection deactivated", "totalInjections", p.injected.Load())
}

// InjectionCount reports how many INVALID responses the proxy has
// synthesized since the most recent Activate call.
func (p *Proxy) InjectionCount() int64 {
	return p.injected.Load()
}

// Close stops the proxy and waits for in-flight connections to drain.
func (p *Proxy) Close() error {
	if !p.stopped.CompareAndSwap(false, true) {
		return nil
	}
	if p.listener != nil {
		_ = p.listener.Close()
	}
	p.wg.Wait()
	return nil
}

// currentRule returns a snapshot of the active rule (or nil).
func (p *Proxy) currentRule() *Rule {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.rule == nil {
		return nil
	}
	rc := *p.rule
	return &rc
}

// upstreamHost returns the configured upstream host:port.
func (p *Proxy) upstreamHost() (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.upstream == "" {
		return "", errors.New("upstream not set")
	}
	return p.upstream, nil
}

// serve handles a single inbound HTTP request. It upgrades to WebSocket and
// then proxies frames to the upstream.
func (p *Proxy) serve(w http.ResponseWriter, r *http.Request) {
	upHost, err := p.upstreamHost()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Upgrade the inbound connection.
	downConn, err := p.upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Error("upgrade failed", "err", err)
		return
	}
	defer downConn.Close()

	// Dial the upstream as a WebSocket client. Forward the Authorization
	// header so the upstream's JWT check still runs.
	upURL := normalizeWSURL(upHost)
	hdr := http.Header{}
	if auth := r.Header.Get("Authorization"); auth != "" {
		hdr.Set("Authorization", auth)
	}
	upConn, _, err := p.dialer.Dial(upURL, hdr)
	if err != nil {
		p.logger.Error("upstream dial failed", "addr", upURL, "err", err)
		return
	}
	defer upConn.Close()

	stop := make(chan struct{}, 2)
	p.wg.Add(2)

	// downstream → (filter) → upstream
	go func() {
		defer p.wg.Done()
		defer func() { stop <- struct{}{} }()
		p.pumpDownToUp(downConn, upConn)
	}()
	// upstream → downstream (verbatim)
	go func() {
		defer p.wg.Done()
		defer func() { stop <- struct{}{} }()
		p.pumpUpToDown(upConn, downConn)
	}()

	<-stop
	_ = downConn.Close()
	_ = upConn.Close()
}

// pumpDownToUp reads JSON-RPC requests from the client (op-node), and either
// forwards them to upstream verbatim, or — if a matching newPayload is seen
// while a rule is active — synthesizes an INVALID response and writes it
// back to the client without ever touching the upstream.
func (p *Proxy) pumpDownToUp(down, up *websocket.Conn) {
	for {
		mt, data, err := down.ReadMessage()
		if err != nil {
			return
		}
		// Only inspect text frames; pass binary frames through unmodified.
		if mt != websocket.TextMessage {
			if err := up.WriteMessage(mt, data); err != nil {
				return
			}
			continue
		}

		intercepted, resp := p.maybeIntercept(data)
		if intercepted {
			if err := down.WriteMessage(websocket.TextMessage, resp); err != nil {
				return
			}
			continue
		}

		if err := up.WriteMessage(mt, data); err != nil {
			return
		}
	}
}

// pumpUpToDown forwards every frame from upstream to downstream verbatim.
func (p *Proxy) pumpUpToDown(up, down *websocket.Conn) {
	for {
		mt, data, err := up.ReadMessage()
		if err != nil {
			return
		}
		if err := down.WriteMessage(mt, data); err != nil {
			return
		}
	}
}

// maybeIntercept inspects a JSON-RPC request frame. If the request is an
// engine_newPayloadV{3,4} matching the active rule, returns (true, response).
// Otherwise returns (false, nil) and the caller should forward verbatim.
func (p *Proxy) maybeIntercept(frame []byte) (bool, []byte) {
	rule := p.currentRule()
	if rule == nil {
		return false, nil
	}
	if rule.MaxInjections > 0 && p.injected.Load() >= int64(rule.MaxInjections) {
		return false, nil
	}

	var req jsonRPCRequest
	if err := json.Unmarshal(frame, &req); err != nil {
		return false, nil
	}
	if req.Method != string(eth.NewPayloadV3) && req.Method != string(eth.NewPayloadV4) {
		return false, nil
	}
	// params[0] is the ExecutionPayload.
	if len(req.Params) < 1 {
		return false, nil
	}
	var payload jsonExecPayload
	if err := json.Unmarshal(req.Params[0], &payload); err != nil {
		return false, nil
	}
	if !rule.matches(&payload) {
		return false, nil
	}

	parent := payload.ParentHash
	validationErr := "elfaultinjector: forced invalid"
	resp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result: &eth.PayloadStatusV1{
			Status:          eth.ExecutionInvalid,
			LatestValidHash: &parent,
			ValidationError: &validationErr,
		},
	}
	out, err := json.Marshal(resp)
	if err != nil {
		return false, nil
	}
	count := p.injected.Add(1)
	p.logger.Info("synthesized INVALID for newPayload",
		"method", req.Method,
		"blockNumber", payload.BlockNumber,
		"injectionCount", count)
	return true, out
}

// jsonRPCRequest is a minimal JSON-RPC request envelope. We use
// json.RawMessage for ID and Params so we can faithfully echo them back.
type jsonRPCRequest struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      json.RawMessage   `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

// jsonRPCResponse is a JSON-RPC response envelope.
type jsonRPCResponse struct {
	JSONRPC string               `json:"jsonrpc"`
	ID      json.RawMessage      `json:"id"`
	Result  *eth.PayloadStatusV1 `json:"result,omitempty"`
	Error   *jsonRPCError        `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// jsonExecPayload is a minimal view of the execution-payload object embedded
// in an engine_newPayloadV{3,4} call. Only fields we filter on are decoded.
type jsonExecPayload struct {
	ParentHash   common.Hash `json:"parentHash"`
	BlockNumber  string      `json:"blockNumber"`
	Transactions []string    `json:"transactions"`
}

// hexToUint64 parses a 0x-prefixed hex quantity (e.g. "0x10").
func hexToUint64(s string) (uint64, error) {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		return strconv.ParseUint(s[2:], 16, 64)
	}
	return strconv.ParseUint(s, 16, 64)
}

// hexToBytes parses a 0x-prefixed hex byte string. Returns the raw byte slice.
func hexToBytes(s string) ([]byte, error) {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		s = s[2:]
	}
	if len(s)%2 != 0 {
		return nil, errors.New("odd hex length")
	}
	out := make([]byte, len(s)/2)
	for i := 0; i < len(s); i += 2 {
		hi, ok1 := hexNibble(s[i])
		lo, ok2 := hexNibble(s[i+1])
		if !ok1 || !ok2 {
			return nil, errors.New("invalid hex")
		}
		out[i/2] = hi<<4 | lo
	}
	return out, nil
}

func hexNibble(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// normalizeWSURL accepts either a host:port or a full ws://... URL and returns
// a URL string suitable for the gorilla/websocket Dialer.
func normalizeWSURL(addr string) string {
	if u, err := url.Parse(addr); err == nil && (u.Scheme == "ws" || u.Scheme == "wss") {
		return addr
	}
	return "ws://" + addr
}
