package elfaultinjector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/testlog"
)

// fakeUpstream is a WebSocket server that records inbound JSON-RPC frames and
// echoes a synthetic VALID response. Tests can inspect what reached the
// upstream to verify pass-through vs. interception.
type fakeUpstream struct {
	t          *testing.T
	server     *httptest.Server
	mu         sync.Mutex
	received   []map[string]any
	upgrader   websocket.Upgrader
	failOnRead atomic.Bool
}

func newFakeUpstream(t *testing.T) *fakeUpstream {
	t.Helper()
	fu := &fakeUpstream{
		t: t,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
	fu.server = httptest.NewServer(http.HandlerFunc(fu.handle))
	return fu
}

func (f *fakeUpstream) URL() string {
	// httptest.NewServer returns http://; the gorilla dialer accepts ws://.
	return strings.Replace(f.server.URL, "http://", "ws://", 1)
}

func (f *fakeUpstream) handle(w http.ResponseWriter, r *http.Request) {
	conn, err := f.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if f.failOnRead.Load() {
			f.t.Errorf("upstream received frame it should not have: %s", string(data))
		}
		var msg map[string]any
		_ = json.Unmarshal(data, &msg)
		f.mu.Lock()
		f.received = append(f.received, msg)
		f.mu.Unlock()
		// Echo a VALID payload-status reply so the client can read.
		id := msg["id"]
		resp := map[string]any{
			"jsonrpc": "2.0",
			"id":      id,
			"result":  map[string]any{"status": "VALID"},
		}
		out, _ := json.Marshal(resp)
		if err := conn.WriteMessage(websocket.TextMessage, out); err != nil {
			return
		}
	}
}

func (f *fakeUpstream) Received() []map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	cp := make([]map[string]any, len(f.received))
	copy(cp, f.received)
	return cp
}

func (f *fakeUpstream) Close() {
	f.server.Close()
}

// startProxy wires up a Proxy in front of the given upstream and returns a
// connected client WebSocket plus a teardown.
func startProxy(t *testing.T, fu *fakeUpstream) (*Proxy, *websocket.Conn) {
	t.Helper()
	lgr := testlog.Logger(t, log.LevelDebug)
	p := New(lgr, [32]byte{1, 2, 3})
	require.NoError(t, p.Start())
	t.Cleanup(func() { _ = p.Close() })
	p.SetUpstream(fu.URL())

	dialer := &websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	cli, _, err := dialer.DialContext(context.Background(), "ws://"+p.Addr(), nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = cli.Close() })
	return p, cli
}

// newPayloadFrame builds a synthetic engine_newPayloadV3 request frame.
func newPayloadFrame(id int, blockNumber uint64, parent common.Hash, txs []string) []byte {
	if txs == nil {
		txs = []string{}
	}
	payload := map[string]any{
		"parentHash":   parent.Hex(),
		"blockNumber":  toHex(blockNumber),
		"transactions": txs,
	}
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      id,
		"method":  "engine_newPayloadV3",
		"params":  []any{payload, []string{}, common.Hash{}.Hex()},
	}
	out, _ := json.Marshal(req)
	return out
}

func toHex(n uint64) string {
	const digits = "0123456789abcdef"
	if n == 0 {
		return "0x0"
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = digits[n&0xF]
		n >>= 4
	}
	return "0x" + string(buf[i:])
}

func readResponse(t *testing.T, cli *websocket.Conn) map[string]any {
	t.Helper()
	require.NoError(t, cli.SetReadDeadline(time.Now().Add(5*time.Second)))
	_, data, err := cli.ReadMessage()
	require.NoError(t, err)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(data, &resp))
	return resp
}

func TestProxy_PassThrough_NotActivated(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	_, cli := startProxy(t, fu)

	frame := newPayloadFrame(1, 100, common.HexToHash("0xabc"), nil)
	require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame))
	resp := readResponse(t, cli)

	// Should be the upstream's VALID echo.
	result, ok := resp["result"].(map[string]any)
	require.True(t, ok, "expected map result, got %#v", resp["result"])
	require.Equal(t, "VALID", result["status"])
	require.Len(t, fu.Received(), 1, "upstream should have seen the frame")
}

func TestProxy_PassThrough_BelowRejectFromBlock(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	p, cli := startProxy(t, fu)
	p.Activate(Rule{RejectFromBlock: 200})

	frame := newPayloadFrame(2, 100, common.HexToHash("0xabc"), nil)
	require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame))
	resp := readResponse(t, cli)

	// Block 100 < threshold 200 → should pass through to upstream.
	result, ok := resp["result"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "VALID", result["status"])
	require.Len(t, fu.Received(), 1)
	require.EqualValues(t, 0, p.InjectionCount())
}

func TestProxy_Intercept_AtOrAboveRejectFromBlock(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	// The upstream must NOT receive any frame in this test.
	fu.failOnRead.Store(true)

	p, cli := startProxy(t, fu)
	p.Activate(Rule{RejectFromBlock: 100})

	parent := common.HexToHash("0xdeadbeef")
	frame := newPayloadFrame(7, 100, parent, nil)
	require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame))
	resp := readResponse(t, cli)

	result, ok := resp["result"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "INVALID", result["status"])
	require.Equal(t, parent.Hex(), result["latestValidHash"])
	verr, _ := result["validationError"].(string)
	require.Contains(t, verr, "elfaultinjector")

	require.EqualValues(t, 7, intOf(resp["id"]))
	require.Empty(t, fu.Received(), "upstream must not see frames while intercepting")
	require.EqualValues(t, 1, p.InjectionCount())
}

func TestProxy_DeactivateRestoresPassthrough(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	p, cli := startProxy(t, fu)

	p.Activate(Rule{RejectFromBlock: 1})
	frame1 := newPayloadFrame(10, 5, common.HexToHash("0xa"), nil)
	require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame1))
	resp1 := readResponse(t, cli)
	result1 := resp1["result"].(map[string]any)
	require.Equal(t, "INVALID", result1["status"])

	p.Deactivate()

	frame2 := newPayloadFrame(11, 5, common.HexToHash("0xa"), nil)
	require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame2))
	resp2 := readResponse(t, cli)
	result2 := resp2["result"].(map[string]any)
	require.Equal(t, "VALID", result2["status"])

	rcvd := fu.Received()
	require.Len(t, rcvd, 1, "upstream should have seen exactly one frame (after deactivate)")
}

func TestProxy_NonNewPayloadMethodAlwaysPassesThrough(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	p, cli := startProxy(t, fu)
	p.Activate(Rule{RejectFromBlock: 1}) // would match any block

	// Send a non-newPayload method.
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      99,
		"method":  "engine_forkchoiceUpdatedV3",
		"params":  []any{},
	}
	frame, _ := json.Marshal(req)
	require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame))
	resp := readResponse(t, cli)

	// Should reach upstream (which echoes VALID).
	result := resp["result"].(map[string]any)
	require.Equal(t, "VALID", result["status"])
	require.Len(t, fu.Received(), 1)
	require.EqualValues(t, 0, p.InjectionCount())
}

func TestProxy_MaxInjectionsCap(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	p, cli := startProxy(t, fu)
	p.Activate(Rule{RejectFromBlock: 1, MaxInjections: 2})

	// First two should be intercepted, third should pass through.
	for i := 1; i <= 3; i++ {
		frame := newPayloadFrame(i, uint64(i), common.HexToHash("0xa"), nil)
		require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame))
		resp := readResponse(t, cli)
		result := resp["result"].(map[string]any)
		if i <= 2 {
			require.Equal(t, "INVALID", result["status"], "request %d should be rejected", i)
		} else {
			require.Equal(t, "VALID", result["status"], "request %d should pass through (cap reached)", i)
		}
	}
	require.EqualValues(t, 2, p.InjectionCount())
	require.Len(t, fu.Received(), 1)
}

func TestProxy_ConcurrentActivateDeactivate_RaceFree(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	p, _ := startProxy(t, fu)

	// Hammer Activate / Deactivate concurrently. With -race, this surfaces
	// any unsynchronized access to the rule pointer.
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				if (i+j)%2 == 0 {
					p.Activate(Rule{RejectFromBlock: uint64(j + 1)})
				} else {
					p.Deactivate()
				}
				_ = p.InjectionCount()
				_ = p.currentRule()
			}
		}(i)
	}
	wg.Wait()
}

func TestProxy_ResponseEchoesRequestID(t *testing.T) {
	fu := newFakeUpstream(t)
	defer fu.Close()
	fu.failOnRead.Store(true)
	p, cli := startProxy(t, fu)
	p.Activate(Rule{RejectFromBlock: 1})

	// Use a string id to verify json.RawMessage faithfully echoes.
	req := map[string]any{
		"jsonrpc": "2.0",
		"id":      "abc-123",
		"method":  "engine_newPayloadV3",
		"params": []any{
			map[string]any{
				"parentHash":   common.Hash{}.Hex(),
				"blockNumber":  "0x5",
				"transactions": []string{},
			},
			[]string{},
			common.Hash{}.Hex(),
		},
	}
	frame, _ := json.Marshal(req)
	require.NoError(t, cli.WriteMessage(websocket.TextMessage, frame))
	resp := readResponse(t, cli)
	require.Equal(t, "abc-123", resp["id"])
}

func intOf(v any) int {
	switch x := v.(type) {
	case float64:
		return int(x)
	case int:
		return x
	case int64:
		return int(x)
	}
	return -1
}
