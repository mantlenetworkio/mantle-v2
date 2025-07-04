package metrics

import (
	"github.com/ethereum-optimism/optimism/op-service/eth"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	txmetrics "github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
)

type noopMetrics struct {
	opmetrics.NoopRefMetrics
	txmetrics.NoopTxMetrics
}

var NoopMetrics Metricer = new(noopMetrics)

func (*noopMetrics) RecordInfo(version string) {}
func (*noopMetrics) RecordUp()                 {}

func (*noopMetrics) RecordValidOutput(l2ref eth.L2BlockRef)      {}
func (*noopMetrics) RecordInvalidOutput(l2ref eth.L2BlockRef)    {}
func (*noopMetrics) RecordOutputChallenged(l2ref eth.L2BlockRef) {}
