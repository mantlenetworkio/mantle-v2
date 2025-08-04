package oracle

import (
	"flag"
	"os"
	"testing"
	"time"

	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/stretchr/testify/assert"
)

var runSpecialTest = flag.Bool("special", false, "Run special test cases")

func TestNewTxCounter(t *testing.T) {
	tests := []struct {
		name           string
		countInterval  uint64
		expectedResult uint64
	}{
		{
			name:           "valid interval",
			countInterval:  7200, // 2 hours
			expectedResult: 7200,
		},
		{
			name:           "below minimum interval",
			countInterval:  10, // below MinCountInterval
			expectedResult: MinCountInterval,
		},
		{
			name:           "zero interval",
			countInterval:  0,
			expectedResult: MinCountInterval,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txCounter := NewTxCounter(nil, tt.countInterval, 3)

			assert.NotNil(t, txCounter)
			assert.Equal(t, tt.expectedResult, txCounter.countInterval)
			assert.Equal(t, uint64(0), txCounter.result)
			assert.NotNil(t, txCounter.stop)
		})
	}
}

func TestTxCounter_Get(t *testing.T) {
	txCounter := NewTxCounter(nil, 3600, 3)

	// Test initial value
	assert.Equal(t, uint64(0), txCounter.Get())

	// Test after setting value
	txCounter.mu.Lock()
	txCounter.result = 100
	txCounter.mu.Unlock()

	assert.Equal(t, uint64(100), txCounter.Get())
}

func TestTxCounter_GetEstimatedDailyTxCount(t *testing.T) {
	txCounter := NewTxCounter(nil, 3600, 3)
	assert.Equal(t, uint64(0), txCounter.GetEstimatedDailyTxCount())

	txCounter.mu.Lock()
	txCounter.result = 100
	txCounter.mu.Unlock()

	assert.Equal(t, uint64(100*86400/3600), txCounter.GetEstimatedDailyTxCount())
}

func TestTxCounter_StartStop(t *testing.T) {
	rpcClient, err := ethclient.Dial("https://rpc.sepolia.mantle.xyz")
	if err != nil {
		t.Skip("No RPC endpoint available")
	}
	txCounter := NewTxCounter(rpcClient, 3600, 3)

	// Test Start
	txCounter.Start()
	time.Sleep(100 * time.Millisecond) // Give time for goroutine to start

	// Test Stop
	txCounter.Stop()
	time.Sleep(100 * time.Millisecond) // Give time for goroutine to stop

	// Verify stop channel is closed
	select {
	case <-txCounter.stop:
		// Expected - channel should be closed
	default:
		t.Error("Stop channel should be closed")
	}
}

func TestTxCounter_UpdateTxCount(t *testing.T) {
	if !*runSpecialTest {
		t.Skip("Skipping special test cases")
	}

	oplog.SetGlobalLogHandler(log.NewTerminalHandlerWithLevel(os.Stdout, log.LvlDebug, true))
	rpcClient, err := ethclient.Dial("https://rpc.sepolia.mantle.xyz")
	if err != nil {
		t.Skip("No RPC endpoint available")
	}
	txCounter := NewTxCounter(rpcClient, 600, 3)
	txCounter.updateTxCount()
	t.Logf("txCounter.Get(): %d", txCounter.Get())
}
