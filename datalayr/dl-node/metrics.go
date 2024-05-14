package dln

import (
	"github.com/Layr-Labs/datalayr/common/logging"

	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	Namespace = "dl_node"
)

type Metrics struct {
	Registered    prometheus.Gauge
	AccNumRequest prometheus.Counter

	AccStore  *prometheus.CounterVec
	CurrStore *prometheus.GaugeVec

	logger   *logging.Logger
	registry *prometheus.Registry
}

func NewMetrics(logger *logging.Logger) *Metrics {
	logger.Trace().Msg("Entering NewMetrics function...")
	defer logger.Trace().Msg("Exiting NewMetrics function...")

	reg := prometheus.NewRegistry()

	// Add Go module collectors
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())

	metrics := &Metrics{
		Registered: promauto.With(reg).NewGauge(
			prometheus.GaugeOpts{
				Namespace: Namespace,
				Name:      "registered",
				Help:      "indicator about if dl node is registered",
			},
		),
		AccNumRequest: promauto.With(reg).NewCounter(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "accumulative_number_request",
				Help:      "the total number of data dispersal requests has been handled by the dl node",
			},
		),
		AccStore: promauto.With(reg).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Name:      "accumulative_store",
				Help:      "the accumulative data dispersal handled by the dl node",
			},
			[]string{"type"},
		),
		CurrStore: promauto.With(reg).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: Namespace,
				Name:      "current_store",
				Help:      "the current unexpired data store handled by the dl node",
			},
			[]string{"type"},
		),
		logger:   logger,
		registry: reg,
	}

	return metrics
}

func (g *Metrics) Start(httpPort string) {
	g.logger.Trace().Msg("Entering Start function...")
	defer g.logger.Trace().Msg("Exiting Start function...")

	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(
			g.registry,
			promhttp.HandlerOpts{},
		))
		err := http.ListenAndServe(httpPort, nil)
		g.logger.Error().Err(err).Msg("Prometheus server failed")
	}()
}

func (g *Metrics) AcceptNewStore(lenWithOverhead int) {
	g.AccStore.WithLabelValues("number").Inc()
	g.AccStore.WithLabelValues("size").Add(float64(lenWithOverhead))
	g.CurrStore.WithLabelValues("number").Inc()
	g.CurrStore.WithLabelValues("size").Add(float64(lenWithOverhead))
}

func (g *Metrics) RemoveExpiredStore(lenWithOverhead int) {
	g.CurrStore.WithLabelValues("number").Dec()
	g.CurrStore.WithLabelValues("size").Sub(float64(lenWithOverhead))
}
