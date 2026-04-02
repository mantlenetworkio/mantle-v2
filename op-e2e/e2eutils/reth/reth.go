package reth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/log"
)

// RethOption configures op-reth CLI arguments.
// Each option receives the current args slice and returns a (possibly modified) copy.
// Options must not mutate the input slice in place.
type RethOption func(args []string) ([]string, error)

// InitReth creates an ExternalInstance (not yet started).
//
// Steps performed:
//  1. Creates a temporary data directory.
//  2. Writes genesis.json to dataDir.
//  3. Runs `op-reth init --datadir <dir> --chain genesis.json` to initialize the chain.
//  4. Prepares base CLI args for the `node` subcommand (ports are injected later in Start).
//  5. Applies all RethOptions.
//  6. Registers tb.Cleanup to ensure the process is stopped even on panic or t.Fatal.
//
// Returns an unstarted ExternalInstance. Call Start() to launch the subprocess.
func InitReth(
	tb testing.TB,
	name string,
	genesis *core.Genesis,
	jwtPath string,
	binPath string,
	opts ...RethOption,
) (*ExternalInstance, error) {
	tb.Helper()

	// 1. Create temp data directory.
	dataDir, err := os.MkdirTemp("", fmt.Sprintf("op-e2e-reth-%s-", name))
	if err != nil {
		return nil, fmt.Errorf("reth %s: create datadir: %w", name, err)
	}

	// 2. Write genesis.json.
	genesisPath := filepath.Join(dataDir, "genesis.json")
	genesisBytes, err := json.Marshal(genesis)
	if err != nil {
		_ = os.RemoveAll(dataDir)
		return nil, fmt.Errorf("reth %s: marshal genesis: %w", name, err)
	}
	if err := os.WriteFile(genesisPath, genesisBytes, 0o644); err != nil {
		_ = os.RemoveAll(dataDir)
		return nil, fmt.Errorf("reth %s: write genesis.json: %w", name, err)
	}

	// 3. Run `op-reth init` to initialize the database from genesis.
	initCmd := exec.Command(binPath, "init",
		"--datadir", dataDir,
		"--chain", genesisPath,
	)
	initOut, err := initCmd.CombinedOutput()
	if err != nil {
		_ = os.RemoveAll(dataDir)
		return nil, fmt.Errorf("reth %s: init failed: %w\noutput: %s", name, err, string(initOut))
	}

	// 4. Prepare base node args (ports injected in Start).
	logsDir := filepath.Join(dataDir, "logs")
	baseArgs := []string{
		"node",
		"-vvv",
		"--chain", genesisPath,
		"--datadir", dataDir,
		"--log.file.directory", logsDir,
		"--http",
		"--http.addr", "127.0.0.1",
		"--http.api", "eth,net,web3,debug,txpool,admin",
		"--ws",
		"--ws.addr", "127.0.0.1",
		"--ws.api", "eth,web3,net",
		"--authrpc.addr", "127.0.0.1",
		"--authrpc.jwtsecret", jwtPath,
		"--txpool.nolocals",
		"--engine.persistence-threshold", "0",
		"--engine.memory-block-buffer-target", "0",
		"--disable-discovery",
		// Use port 0 for P2P so parallel tests don't conflict on 30303.
		"--port", "0",
		// Disable IPC server: the default path /tmp/reth.ipc is shared across
		// all processes and causes conflicts when tests run in parallel.
		"--ipcdisable",
	}

	// 5. Apply options.
	for _, opt := range opts {
		baseArgs, err = opt(baseArgs)
		if err != nil {
			_ = os.RemoveAll(dataDir)
			return nil, fmt.Errorf("reth %s: apply option: %w", name, err)
		}
	}

	inst := &ExternalInstance{
		tb:          tb,
		name:        name,
		binPath:     binPath,
		dataDir:     dataDir,
		jwtPath:     jwtPath,
		genesisPath: genesisPath,
		baseArgs:    baseArgs,
	}

	// 6. Register cleanup to stop the process even on panic or t.Fatal.
	tb.Cleanup(func() {
		_ = inst.Close()
	})

	return inst, nil
}

// WithSequencerHTTP configures --rollup.sequencer-http for verifier/replica nodes.
func WithSequencerHTTP(url string) RethOption {
	return func(args []string) ([]string, error) {
		return append(args, "--rollup.sequencer-http", url), nil
	}
}

// WithSyncMode configures --rollup.sync-mode (e.g. "full", "consensus-layer").
func WithSyncMode(mode string) RethOption {
	return func(args []string) ([]string, error) {
		return append(args, "--rollup.sync-mode", mode), nil
	}
}

// WithVerbosity configures the log level via -v repetitions.
// level 1=error, 2=warn, 3=info, 4=debug (reth uses -v repetition, not --log.level).
func WithVerbosity(level int) RethOption {
	return func(args []string) ([]string, error) {
		if level < 1 {
			level = 1
		}
		if level > 5 {
			level = 5
		}
		flag := "-" + repeatV(level)
		// Remove existing verbosity flag if present; allocate a new slice to
		// avoid mutating the input (args[:0] shares the backing array).
		filtered := make([]string, 0, len(args))
		for _, a := range args {
			if len(a) > 1 && a[0] == '-' && allV(a[1:]) {
				continue
			}
			filtered = append(filtered, a)
		}
		return append(filtered, flag), nil
	}
}

// WithExtraArgs appends arbitrary CLI arguments as a passthrough escape hatch.
func WithExtraArgs(extraArgs ...string) RethOption {
	return func(args []string) ([]string, error) {
		return append(args, extraArgs...), nil
	}
}

// streamToLog reads lines from r and writes them to logger with the given prefix.
// It exits when ctx is cancelled or r is closed.
func streamToLog(ctx context.Context, logger log.Logger, prefix string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}
		logger.Info(prefix + scanner.Text())
	}
}

func repeatV(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'v'
	}
	return string(b)
}

func allV(s string) bool {
	for _, c := range s {
		if c != 'v' {
			return false
		}
	}
	return len(s) > 0
}

// Logger wraps testlog to provide a named logger for a test.
func Logger(tb testing.TB) log.Logger {
	return testlog.Logger(tb, log.LevelInfo)
}
