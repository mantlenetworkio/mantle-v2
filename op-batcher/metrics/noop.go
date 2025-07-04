package metrics

import (
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	txmetrics "github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
	"github.com/ethereum/go-ethereum/core/types"
)

type noopMetrics struct {
	opmetrics.NoopRefMetrics
	txmetrics.NoopTxMetrics
}

var NoopMetrics Metricer = new(noopMetrics)

func (*noopMetrics) Document() []opmetrics.DocumentedMetric { return nil }

func (*noopMetrics) RecordInfo(version string) {}
func (*noopMetrics) RecordUp()                 {}

func (*noopMetrics) RecordLatestL1Block(l1ref eth.L1BlockRef)               {}
func (*noopMetrics) RecordL2BlocksLoaded(eth.L2BlockRef)                    {}
func (*noopMetrics) RecordChannelOpened(derive.ChannelID, int)              {}
func (*noopMetrics) RecordL2BlocksAdded(eth.L2BlockRef, int, int, int, int) {}
func (*noopMetrics) RecordL2BlockInPendingQueue(*types.Block)               {}
func (*noopMetrics) RecordL2BlockInChannel(*types.Block)                    {}

func (*noopMetrics) RecordChannelClosed(derive.ChannelID, int, int, int, int, error) {}

func (*noopMetrics) RecordChannelFullySubmitted(derive.ChannelID) {}
func (*noopMetrics) RecordChannelTimedOut(derive.ChannelID)       {}

func (*noopMetrics) RecordBatchTxSubmitted() {}
func (*noopMetrics) RecordBatchTxSuccess()   {}
func (*noopMetrics) RecordBatchTxFailed()    {}

func (*noopMetrics) RecordBatchTxInitDataSubmitted() {}
func (*noopMetrics) RecordBatchTxInitDataSuccess()   {}
func (*noopMetrics) RecordBatchTxInitDataFailed()    {}

func (*noopMetrics) RecordBatchTxConfirmDataSubmitted() {}
func (*noopMetrics) RecordBatchTxConfirmDataSuccess()   {}
func (*noopMetrics) RecordBatchTxConfirmDataFailed()    {}

func (*noopMetrics) RecordRollupRetry(time int32) {}
func (*noopMetrics) RecordDaRetry(time int32)     {}

func (m *noopMetrics) RecordInitReferenceBlockNumber(dataStoreId uint32) {}
func (m *noopMetrics) RecordConfirmedDataStoreId(dataStoreId uint32)     {}

func (*noopMetrics) RecordTxOverMaxLimit() {}

func (*noopMetrics) RecordDaNonSignerPubkeys(num int) {

}

func (*noopMetrics) RecordEigenDAFailback(txs int) {}

func (*noopMetrics) RecordInterval(method string) func(error) {
	return func(error) {}
}
