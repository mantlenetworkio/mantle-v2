package trafficGen

import(
	"time"
	"sync"
	"github.com/Layr-Labs/datalayr/common/logging"
)

// need mutex protection if multi-thread, make everything array
type Metrics struct {
	numSource       uint64 // total number of generators
	target          string // which disperser socket
	numSuccesses    []uint64
	numTotals       []uint64
	numBytes        []uint64
	highestStoreIds []uint32
	StartTime       time.Time
	Logger          *logging.Logger
	mu              *sync.Mutex
}

func NewMetrics(n uint64, target string, logger *logging.Logger) *Metrics {
	metrics := &Metrics {
		numSource: n,
		target: target,
		numSuccesses: make([]uint64, n),
		numTotals: make([]uint64, n),
		numBytes: make([]uint64, n),
		highestStoreIds: make([]uint32, n),
		StartTime: time.Now(),
		Logger: logger,
		mu: &sync.Mutex{},
	}

	return metrics
}

func (m *Metrics) Update(id int, storeId uint32, numBytes uint64, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.numTotals[id] += 1
	if err == nil {
		m.numSuccesses[id] += 1
		m.highestStoreIds[id] = storeId
		m.numBytes[id] += numBytes
	}
}

func (m *Metrics) Log() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	duration := now.Sub(m.StartTime)
	m.Logger.Info().Msg("------------------------------------------------------")
	m.Logger.Info().Msgf("                 Summary  %v           ", duration)

	sumBytes := uint64(0)
	sumNumSuccess := uint64(0)
	sumNumTotal := uint64(0)
	for i := 0 ; i < int(m.numSource) ; i++ {
		throughput := float64(m.numBytes[i]) / duration.Seconds()
		m.Logger.Info().Msgf(
			"src %v. tgt %v. numSuccess/numTotal (%v/%v) %v. HighestStore Id %v. numBytes %v. Thx %v B/sec",
			i,
			m.target,
			m.numSuccesses[i],
			m.numTotals[i],
			float64(m.numSuccesses[i])/float64(m.numTotals[i]),
			m.highestStoreIds[i],
			m.numBytes[i],
			throughput,
		)

		sumBytes += m.numBytes[i]
		sumNumSuccess += m.numSuccesses[i]
		sumNumTotal += m.numTotals[i]
	}

	aggregatedThx := float64(sumBytes) / float64(duration.Seconds())
	aggregatedPercentage := float64(sumNumSuccess) / float64(sumNumTotal)
	m.Logger.Info().Msgf(
		"All: numSuccess/Percentage (%v/%v) %v. Thx %v B/sec",
		sumNumSuccess,
		sumNumTotal,
		aggregatedPercentage,
		aggregatedThx,
	)
}
