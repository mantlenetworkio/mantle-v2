package disperser_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/Layr-Labs/datalayr/common/logging"
	disperser "github.com/Layr-Labs/datalayr/dl-disperser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	logger  *logging.Logger
	metrics *disperser.Metrics

	maxCacheSize   = uint64(10000)
	expireDuration = int64(100000)
	cleanPeriod    = 10 * time.Second
)

func setupSuite(t *testing.T) func(t *testing.T) {

	logger = logging.GetNoopLogger()
	metrics = disperser.NewMetrics(logger)

	return func(t *testing.T) {
		fmt.Println("Tearing down suite")
	}
}

func TestCodedDataCacheAddGet(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	cache := disperser.NewCodedDataCache(
		maxCacheSize,
		expireDuration,
		cleanPeriod,
		metrics,
		logger,
	)

	var h1 [32]byte
	store1 := disperser.Store{
		HeaderBytes: h1[:],
		TotalSize:   300,
	}

	err := cache.Add(&store1)
	require.Nil(t, err)

	store1Get, err := cache.Get(h1)
	require.Nil(t, err)

	assert.Equal(t, store1Get.TotalSize, store1.TotalSize, "retrieved wrong data")
}

func TestCodedDataCacheAddDelete(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	cache := disperser.NewCodedDataCache(
		maxCacheSize,
		expireDuration,
		cleanPeriod,
		metrics,
		logger,
	)

	var h1 [32]byte
	store1 := disperser.Store{
		HeaderBytes: h1[:],
		TotalSize:   300,
	}

	err := cache.Add(&store1)
	require.Nil(t, err)

	err = cache.Delete(h1)
	require.Nil(t, err)

	var h2 [32]byte
	_, err = cache.Get(h2)
	require.NotNil(t, err)
}

func TestCodedDataCacheAddTwice(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	cache := disperser.NewCodedDataCache(
		maxCacheSize,
		expireDuration,
		cleanPeriod,
		metrics,
		logger,
	)

	var h1 [32]byte
	store1 := disperser.Store{
		HeaderBytes: h1[:],
		TotalSize:   300,
	}

	err := cache.Add(&store1)
	require.Nil(t, err)

	err = cache.Add(&store1)
	require.NotNil(t, err)
}

func TestCodedDataCacheDeleteEmpty(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	cache := disperser.NewCodedDataCache(
		maxCacheSize,
		expireDuration,
		cleanPeriod,
		metrics,
		logger,
	)

	var h1 [32]byte
	err := cache.Delete(h1)
	require.NotNil(t, err)
}

func TestCodedDataCacheFullCache(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	cache := disperser.NewCodedDataCache(
		0,
		expireDuration,
		cleanPeriod,
		metrics,
		logger,
	)

	var h1 [32]byte
	store1 := disperser.Store{
		HeaderBytes: h1[:],
		TotalSize:   300,
	}

	err := cache.Add(&store1)
	require.NotNil(t, err)
}
