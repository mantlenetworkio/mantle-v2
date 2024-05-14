package disperser

import(
	"github.com/Layr-Labs/datalayr/common/logging"

	"time"
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const (
	Namespace = "dl_disperser"
)

type Metrics struct {
	NumRegistered          prometheus.Gauge
	// request check api hits, request is whole sequence
	Request             *prometheus.CounterVec
	// 
	Encode              *prometheus.CounterVec
	Disperse            *prometheus.CounterVec
	// 
	CodedDataCache      *prometheus.GaugeVec
	DispersalLatency     prometheus.Summary

	SizeSuccessTraffic     prometheus.Gauge
	//NumFailedByOperator   *prometheus.CounterVec
	logger                 *logging.Logger
	registry               *prometheus.Registry
}

func NewMetrics(logger *logging.Logger) *Metrics {
	reg := prometheus.NewRegistry()

	// Add Go module collectors
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	reg.MustRegister(collectors.NewGoCollector())

	metrics := &Metrics {
		NumRegistered: promauto.With(reg).NewGauge(
				prometheus.GaugeOpts{
					Namespace: Namespace,
					Name:    "num_registered",
					Help:    "number registered operators on smart contract",
				},
			),
		Request: promauto.With(reg).NewCounterVec(
				prometheus.CounterOpts{
					Namespace: Namespace,
					Name:    "request",
					Help:    "the number and size of total dispersal request",
				},
				[]string{"outcome", "type"},
			),
		Encode: promauto.With(reg).NewCounterVec(
				prometheus.CounterOpts{
					Namespace: Namespace,
					Name:    "encode",
					Help:    "the encode requests",
				},
				[]string{"outcome", "type"},
			),
		Disperse: promauto.With(reg).NewCounterVec(
				prometheus.CounterOpts{
					Namespace: Namespace,
					Name:    "disperse",
					Help:    "the disperse requests",
				},
				[]string{"outcome", "type", "latms"},
			),
		CodedDataCache: promauto.With(reg).NewGaugeVec(
				prometheus.GaugeOpts{
					Namespace: Namespace,
					Name:    "codedDataCache",
					Help:    "display metrics about the cache that stored encoded data",
				},
				[]string{"type"},
			),
		SizeSuccessTraffic: promauto.With(reg).NewGauge(
				prometheus.GaugeOpts{
					Namespace: Namespace,
					Name:    "size_success_traffic",
					Help:    "total size of all successful traffic regardless of api",
				},
			),
		DispersalLatency:  promauto.With(reg).NewSummary(
				prometheus.SummaryOpts{
					Namespace: Namespace,
					Name:    "dispersal_latency_ms",
					Help:    "dispersal latency summary in milliseconds",
					Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.01, 0.99: 0.001},
				},
			),
		//NumFailedByOperator: promauto.With(reg).NewCounterVec(
				//prometheus.CounterOpts{
					//Namespace: Namespace,
					//Name:    "num_handled_request",
					//Help:    "the number handled requests by disperser",
				//},
				//[]string{"outcome"},
			//),
		logger: logger,
		registry: reg,
	}
	logger.Info().Msgf("return metrics")
	return metrics
}


func (g *Metrics) Start(httpPort string){
	g.logger.Info().Msgf("http %v", httpPort)
	go func() {
		http.Handle("/metrics", promhttp.HandlerFor(
			g.registry,
			promhttp.HandlerOpts{},
		))
		err := http.ListenAndServe(httpPort, nil)
		g.logger.Error().Err(err).Msg("Prometheus server failed")
	}()
}

func (g *Metrics) RecordNewRequest(isSuccess bool, origLen int, lat time.Duration) {
	if isSuccess {
		g.Request.WithLabelValues("success", "number").Inc()
		g.Request.WithLabelValues("success", "size").Add(float64(origLen))
		g.Request.WithLabelValues("success", "latms").Add(float64(lat.Milliseconds()))
		g.SizeSuccessTraffic.Add(float64(origLen))
		g.DispersalLatency.Observe(float64(lat.Milliseconds()))
	} else {
		g.Request.WithLabelValues("failure", "number").Inc()
		//g.Request.WithLabelValues("failure", "size").Add(float64(origLen))
	}
}

//func (g *Metrics) RecordCodedDataCache() {
	//g.CodedDataCache.WithLabelValues("success").Inc()
//}

