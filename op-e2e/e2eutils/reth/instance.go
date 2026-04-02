package reth

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/services"
	"github.com/ethereum-optimism/optimism/op-service/endpoint"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/log"
)

// ExternalInstance manages an op-reth node running as an external subprocess.
// It implements services.EthInstance.
type ExternalInstance struct {
	tb          testing.TB
	name        string
	binPath     string
	cmd         *exec.Cmd
	dataDir     string
	httpPort    int
	wsPort      int
	authPort    int
	jwtPath     string
	genesisPath string
	baseArgs    []string           // CLI args prepared during Init (no ports yet)
	logCancel   context.CancelFunc
	exitCh      chan error          // receives exactly one cmd.Wait() result
	started     atomic.Bool        // guards against double-Start
}

var _ services.EthInstance = (*ExternalInstance)(nil)

// Start allocates ports, launches the op-reth subprocess, and waits for RPC
// readiness. It retries on port-conflict errors (up to 3 times). Other
// startup failures are returned immediately.
func (ei *ExternalInstance) Start() error {
	if !ei.started.CompareAndSwap(false, true) {
		return fmt.Errorf("reth %s: Start called on an already-started instance", ei.name)
	}
	const maxRetries = 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		httpPort, wsPort, authPort, err := allocatePorts()
		if err != nil {
			return fmt.Errorf("reth %s: allocate ports: %w", ei.name, err)
		}
		ei.httpPort = httpPort
		ei.wsPort = wsPort
		ei.authPort = authPort

		cmd := ei.buildCmd(httpPort, wsPort, authPort)
		ei.cmd = cmd

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("reth %s: stdout pipe: %w", ei.name, err)
		}
		// Tee stderr so we can inspect it for port-conflict errors.
		rawStderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("reth %s: stderr pipe: %w", ei.name, err)
		}
		var stderrBuf bytes.Buffer
		stderr := io.TeeReader(rawStderr, &stderrBuf)

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("reth %s: start process: %w", ei.name, err)
		}

		// Stream process output to test logger.
		ctx, cancel := context.WithCancel(context.Background())
		ei.logCancel = cancel
		logger := testlog.Logger(ei.tb, log.LevelInfo)
		prefix := fmt.Sprintf("[reth-%s] ", ei.name)
		go streamToLog(ctx, logger, prefix, stdout)
		go streamToLog(ctx, logger, prefix+"ERR ", stderr)

		// Single cmd.Wait() goroutine — stored on the struct so Close() can drain it.
		exitCh := make(chan error, 1)
		ei.exitCh = exitCh
		go func() { exitCh <- cmd.Wait() }()

		select {
		case exitErr := <-exitCh:
			// Process exited quickly — check stderr for a port-conflict.
			cancel()
			// Give the stderr tee goroutine a moment to flush.
			time.Sleep(100 * time.Millisecond)
			if isPortConflict(stderrBuf.String()) {
				lastErr = exitErr
				// Reset exitCh so the next attempt stores a fresh channel.
				ei.exitCh = nil
				continue // retry with new ports
			}
			return fmt.Errorf("reth %s: process exited during startup: %w", ei.name, exitErr)
		case <-time.After(2 * time.Second):
			// Process still running after 2s — proceed to readiness checks.
		}

		// Wait for RPC to be ready.
		timeout := readyTimeoutFromEnv(ei.tb)
		if err := waitForRPC(ctx, fmt.Sprintf("http://127.0.0.1:%d", httpPort), timeout); err != nil {
			cancel()
			_ = cmd.Process.Kill()
			<-ei.exitCh // drain so Close() does not double-Wait
			return fmt.Errorf("reth %s: RPC not ready within %s: %w", ei.name, timeout, err)
		}
		if err := waitForAuthRPC(ctx, fmt.Sprintf("http://127.0.0.1:%d", authPort), timeout); err != nil {
			cancel()
			_ = cmd.Process.Kill()
			<-ei.exitCh // drain so Close() does not double-Wait
			return fmt.Errorf("reth %s: Auth RPC not ready within %s: %w", ei.name, timeout, err)
		}
		return nil
	}
	return fmt.Errorf("reth %s: failed to start after %d attempts (last error: %w)", ei.name, maxRetries, lastErr)
}

// UserRPC returns the HTTP/WS endpoint for user-facing JSON-RPC.
func (ei *ExternalInstance) UserRPC() endpoint.RPC {
	return endpoint.WsOrHttpRPC{
		WsURL:   fmt.Sprintf("ws://127.0.0.1:%d", ei.wsPort),
		HttpURL: fmt.Sprintf("http://127.0.0.1:%d", ei.httpPort),
	}
}

// AuthRPC returns the Engine API (authenticated) endpoint.
func (ei *ExternalInstance) AuthRPC() endpoint.RPC {
	return endpoint.HttpURL(fmt.Sprintf("http://127.0.0.1:%d", ei.authPort))
}

// Close sends SIGTERM to the op-reth process group, waits up to 10 s, then
// SIGKILL, cancels log goroutines, and removes the temporary data directory.
func (ei *ExternalInstance) Close() error {
	if ei.logCancel != nil {
		// Cancel log-streaming goroutines before waiting for the process to
		// exit. This is necessary because cmd.Wait() blocks until all I/O
		// goroutines finish — and those goroutines only check ctx.Done() after
		// a Scan() returns. Cancelling early avoids a deadlock when reth
		// spawns child processes that keep the pipes open.
		ei.logCancel()
	}
	if ei.cmd != nil && ei.cmd.Process != nil && ei.exitCh != nil {
		pid := ei.cmd.Process.Pid
		// Send SIGTERM to the entire process group (negative PID = group).
		_ = syscall.Kill(-pid, syscall.SIGTERM)
		select {
		case <-ei.exitCh:
			// Process group exited gracefully.
		case <-time.After(10 * time.Second):
			// Force-kill the entire process group.
			_ = syscall.Kill(-pid, syscall.SIGKILL)
			select {
			case <-ei.exitCh:
			case <-time.After(5 * time.Second):
				// If still blocked, give up — the OS will eventually reap.
			}
		}
	}
	if ei.dataDir != "" {
		if err := os.RemoveAll(ei.dataDir); err != nil {
			// Log but don't fail — reth may leave lock files after SIGKILL.
			ei.tb.Logf("warning: failed to remove reth datadir %s: %v", ei.dataDir, err)
		}
	}
	return nil
}

// buildCmd constructs an exec.Cmd for the op-reth node subcommand with the
// given ports. A new Cmd is built each time because exec.Cmd cannot be reused.
// Setpgid is set so that the entire process group can be killed on close.
func (ei *ExternalInstance) buildCmd(httpPort, wsPort, authPort int) *exec.Cmd {
	args := make([]string, len(ei.baseArgs))
	copy(args, ei.baseArgs)
	args = append(args,
		"--http.port", fmt.Sprintf("%d", httpPort),
		"--ws.port", fmt.Sprintf("%d", wsPort),
		"--authrpc.port", fmt.Sprintf("%d", authPort),
	)
	cmd := exec.Command(ei.binPath, args...)
	// Create a new process group so we can kill all descendant processes at once.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	return cmd
}

// allocatePorts allocates 3 ephemeral TCP ports by binding and immediately
// releasing them. There is a small TOCTOU window between release and reth
// binding; the retry logic in Start() handles the rare conflict case.
func allocatePorts() (httpPort, wsPort, authPort int, err error) {
	httpPort, err = getFreePort()
	if err != nil {
		return
	}
	wsPort, err = getFreePort()
	if err != nil {
		return
	}
	authPort, err = getFreePort()
	return
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// isPortConflict returns true when the captured stderr output contains a port
// binding error. Checking stderr (rather than the ExitError string) is
// necessary because exec.ExitError only carries the exit code, not stdout/stderr.
func isPortConflict(stderr string) bool {
	return strings.Contains(stderr, "address already in use")
}

// waitForRPC polls the given URL with eth_chainId until it responds or timeout elapses.
func waitForRPC(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	body := `{"jsonrpc":"2.0","method":"eth_chainId","params":[],"id":1}`
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return fmt.Errorf("timed out waiting for RPC at %s", url)
}

// waitForAuthRPC polls the auth RPC with engine_exchangeCapabilities until ready.
func waitForAuthRPC(ctx context.Context, url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 2 * time.Second}
	body := `{"jsonrpc":"2.0","method":"engine_exchangeCapabilities","params":[[]],"id":1}`
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			// Auth RPC returns 401 without JWT but the port is alive; accept that too.
			if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusUnauthorized {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return fmt.Errorf("timed out waiting for auth RPC at %s", url)
}

// readyTimeoutFromEnv reads OP_E2E_L2_READY_TIMEOUT; defaults to 30 s.
// Note: this is intentionally a local copy — importing l2backend here would
// create a circular import (l2backend imports reth).
func readyTimeoutFromEnv(tb testing.TB) time.Duration {
	if s := os.Getenv("OP_E2E_L2_READY_TIMEOUT"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			tb.Logf("warning: invalid OP_E2E_L2_READY_TIMEOUT %q, using default 30s: %v", s, err)
			return 30 * time.Second
		}
		return d
	}
	return 30 * time.Second
}
