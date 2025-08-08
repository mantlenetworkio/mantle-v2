package batcher

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"google.golang.org/protobuf/proto"

	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	se "github.com/ethereum-optimism/optimism/op-service/eth"

	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

const (
	EigenRollupMaxSize  = 1024 * 1024 * 300
	DaLoopRetryNum      = 10
	EigenRPCRetryNum    = 3
	BytesPerCoefficient = 31
	MaxblobNum          = 4 // max number of blobs, the bigger the more possible of timeout
)

var ErrInitDataStore = errors.New("init data store transaction failed")

func (l *BatchSubmitter) mantleDALoop() {
	defer l.wg.Done()
	ticker := time.NewTicker(l.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := l.loadBlocksIntoState(l.shutdownCtx)
			if errors.Is(err, ErrReorg) {
				err := l.state.Close()
				if err != nil {
					l.log.Error("error closing the channel manager to handle a L2 reorg", "err", err)
				}
				l.state.Clear()
				l.state.clearMantleDAStatus()
				continue
			} else if err != nil {
				l.log.Error("load block into state err", "err", err)
				continue
			}
			l.publishStateToMantleDA()

		case <-l.shutdownCtx.Done():
			err := l.state.Close()
			if err != nil {
				l.log.Error("error closing the channel manager", "err", err)
			}
			return
		}
	}
}

// publishStateToMantleDA loops through the block data loaded into `state` and
// submits the associated data to the MantleDA in the form of channel frames.
// batch frames in one rollup transaction to MantleDA
func (l *BatchSubmitter) publishStateToMantleDA() {

	for {
		isFull, err := l.appendNextRollupData(l.killCtx)
		if err != nil && err != io.EOF {
			l.log.Error("failed to  get next tx data", "err", err)
			return
		}
		if isFull {
			var (
				done bool
				err  error
			)

			done, err = l.loopEigenDa()
			if err != nil {
				l.log.Error("failed to rollup da to da", "err", err)
				l.log.Warn("reset state in channel manager")
				l.reset()
				return
			}
			if done {
				return
			}
		}
		if err == io.EOF {
			return
		}

	}

}

// appendNextRollupData get next txData and cache the data in pendingTransactions, and determine whether the channel is full
func (l *BatchSubmitter) appendNextRollupData(ctx context.Context) (bool, error) {
	// send all available transactions
	l1tip, err := l.l1Tip(ctx)
	if err != nil {
		l.log.Error("failed to query L1 tip", "error", err)
		return false, err
	}
	l.recordL1Tip(l1tip)

	// Collect next transaction data
	txdata, err := l.state.TxData(l1tip.ID())
	if err != nil && err != io.EOF {
		l.log.Error("unable to get tx data", "err", err)
		return false, err
	}
	if err == nil {
		l.state.daPendingTxData[txdata.ID()] = txdata
	}

	if l.state.pendingChannel != nil && l.state.pendingChannel.IsFull() && !l.state.pendingChannel.HasFrame() {
		return true, err
	}
	return false, err
}

func (l *BatchSubmitter) loopEigenDa() (bool, error) {
	//it means that all the txData has been rollup
	if len(l.state.daPendingTxData) == 0 {
		l.log.Info("all txData of current channel have been rollup")
		return true, nil
	}

	var err error
	var wrappedData []byte

	daData, err := l.txAggregatorForEigenDa()
	if err != nil {
		l.log.Error("loopEigenDa txAggregatorForEigenDa err", "err", err)
		return false, err
	}

	currentL1, err := l.l1Tip(l.killCtx)
	if err != nil {
		l.log.Warn("loopEigenDa l1Tip", "err", err)
		return false, err
	}

	l.log.Info("disperseEigenDaData", "skip da rpc", l.Config.SkipEigenDaRpc)

	eigendaSuccess := false
	if !l.Config.SkipEigenDaRpc {
		timeoutTime := time.Now().Add(l.EigenDA.DisperseBlobTimeout)
		for retry := 0; retry < EigenRPCRetryNum; retry++ {
			l.metr.RecordDaRetry(int32(retry))
			if time.Now().After(timeoutTime) {
				l.log.Warn("loopEigenDa disperseEigenDaData timeout", "retry time", retry, "err", err)
				break
			}

			wrappedData, err = l.disperseEigenDaData(daData)
			if err == nil && len(wrappedData) > 0 {
				eigendaSuccess = true
				break
			}

			if err != nil && !errors.Is(err, eigenda.ErrNetwork) {
				l.log.Warn("unrecoverable error in disperseEigenDaData", "retry time", retry, "err", err)
				break
			}

			l.log.Warn("loopEigenDa disperseEigenDaData err,need to try again", "retry time", retry, "err", err)
			time.Sleep(5 * time.Second)
		}
	}

	var candidates []*txmgr.TxCandidate
	if eigendaSuccess {
		candidate := l.calldataTxCandidate(wrappedData)
		candidates = append(candidates, candidate)
	} else {
		if blobCandidates, err := l.blobTxCandidates(daData); err != nil {
			l.log.Warn("failed to create blob tx candidate", "err", err)
			return false, err
		} else {
			candidates = append(candidates, blobCandidates...)
			l.metr.RecordEigenDAFailback(len(blobCandidates))
		}
	}

	var lastReceipt *types.Receipt
	var successTxs []ecommon.Hash
	for loopRetry := 0; loopRetry < DaLoopRetryNum; loopRetry++ {
		l.metr.RecordRollupRetry(int32(loopRetry))
		failedIdx := 0
		for idx, tx := range candidates {
			lastReceipt, err = l.txMgr.Send(l.killCtx, *tx)
			if err != nil || lastReceipt.Status == types.ReceiptStatusFailed {
				l.log.Warn("failed to send tx candidate", "err", err)
				break
			}
			successTxs = append(successTxs, lastReceipt.TxHash)
			failedIdx = idx + 1
		}
		candidates = candidates[failedIdx:]

		if len(candidates) > 0 {
			l.log.Warn("failed to rollup", "err", err, "retry time", loopRetry)
		} else {
			l.log.Info("rollup success", "success txs", successTxs)
			break
		}
	}

	if len(candidates) > 0 {
		err = fmt.Errorf("failed to rollup %d tx candidates", len(candidates))
		l.log.Error("failed to rollup", "err", err)
		l.metr.RecordBatchTxConfirmDataFailed()
		return false, err
	}

	l.metr.RecordBatchTxConfirmDataSuccess()
	l.recordConfirmedEigenDATx(lastReceipt)

	//create a new channel now for reducing the disperseEigenDaData latency time
	if err = l.state.ensurePendingChannel(currentL1.ID()); err != nil {
		l.log.Error("failed to ensurePendingChannel", "err", err)
		return false, err
	}
	l.state.registerL1Block(currentL1.ID())

	return true, nil

}

func (l *BatchSubmitter) calldataTxCandidate(data []byte) *txmgr.TxCandidate {
	l.log.Info("building Calldata transaction candidate", "size", len(data))
	return &txmgr.TxCandidate{
		To:     &l.Rollup.BatchInboxAddress,
		TxData: data,
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (l *BatchSubmitter) blobTxCandidates(data [][]byte) ([]*txmgr.TxCandidate, error) {
	l.log.Info("building Blob transaction candidate", "size", len(data))
	candidates := []*txmgr.TxCandidate{}
	dataInTx := [][]byte{}
	encodeData := []byte{}

	for _, frameData := range data {
		dataInTx = append(dataInTx, frameData)
		nextEncodeData, err := rlp.EncodeToBytes(dataInTx)
		if err != nil {
			l.log.Error("op-batcher unable to encode txn", "err", err)
			return nil, err
		}

		if len(nextEncodeData) > se.MaxBlobDataSize*MaxblobNum {
			if len(encodeData) == 0 {
				err := fmt.Errorf("single frame data size %d larger than max blob transaction maximum size", len(nextEncodeData))
				l.log.Error("empty encodeData", "err", err)
				return nil, err
			}
			blobs := []*se.Blob{}
			for idx := 0; idx < len(encodeData); idx += se.MaxBlobDataSize {
				blobData := encodeData[idx : idx+minInt(len(encodeData)-idx, se.MaxBlobDataSize)]
				var blob se.Blob
				if err := blob.FromData(blobData); err != nil {
					return nil, err
				}
				blobs = append(blobs, &blob)
			}
			candidates = append(candidates, &txmgr.TxCandidate{
				To:    &l.Rollup.BatchInboxAddress,
				Blobs: blobs,
			})
			dataInTx = [][]byte{frameData}
			encodeData, err = rlp.EncodeToBytes(dataInTx)
			if err != nil {
				l.log.Error("op-batcher unable to encode txn", "err", err)
				return nil, err
			}
		} else {
			encodeData = nextEncodeData
		}

	}

	if len(dataInTx) > 0 {
		blobs := []*se.Blob{}
		for idx := 0; idx < len(encodeData); idx += se.MaxBlobDataSize {
			blobData := encodeData[idx : idx+minInt(len(encodeData)-idx, se.MaxBlobDataSize)]
			var blob se.Blob
			if err := blob.FromData(blobData); err != nil {
				return nil, err
			}
			blobs = append(blobs, &blob)
		}
		candidates = append(candidates, &txmgr.TxCandidate{
			To:    &l.Rollup.BatchInboxAddress,
			Blobs: blobs,
		})
	}

	return candidates, nil
}

func (l *BatchSubmitter) disperseEigenDaData(data [][]byte) ([]byte, error) {
	encodeData, err := rlp.EncodeToBytes(data)
	if err != nil {
		l.log.Error("op-batcher unable to encode txn", "err", err)
		return nil, err
	}

	blobInfo, err := l.eigenDA.DisperseBlob(l.shutdownCtx, encodeData)
	if err != nil {
		l.log.Error("Unable to publish batch frameset to EigenDA", "err", err)
		return nil, err

	}
	commitment, err := eigenda.EncodeCommitment(blobInfo)
	if err != nil {
		return nil, err
	}

	quorumIDs := make([]uint32, len(blobInfo.BlobHeader.BlobQuorumParams))
	for i := range quorumIDs {
		quorumIDs[i] = blobInfo.BlobHeader.BlobQuorumParams[i].QuorumNumber
	}
	calldataFrame := &op_service.CalldataFrame{
		Value: &op_service.CalldataFrame_FrameRef{
			FrameRef: &op_service.FrameRef{
				ReferenceBlockNumber: blobInfo.BlobVerificationProof.BatchMetadata.BatchHeader.ReferenceBlockNumber,
				QuorumIds:            quorumIDs,
				BlobLength:           uint32(len(encodeData)),
				Commitment:           commitment,
			},
		},
	}
	wrappedData, err := proto.Marshal(calldataFrame)
	if err != nil {
		return nil, err
	}

	l.metr.RecordConfirmedDataStoreId(blobInfo.BlobVerificationProof.BatchId)
	l.metr.RecordInitReferenceBlockNumber(blobInfo.BlobVerificationProof.BatchMetadata.BatchHeader.ReferenceBlockNumber)
	// prepend the derivation version to the data
	l.log.Info("Prepending derivation version to calldata")
	wrappedData = append([]byte{eigenda.DerivationVersionEigenda}, wrappedData...)

	return wrappedData, nil

}

func (l *BatchSubmitter) isRetry(retry *int32) bool {
	*retry = *retry + 1
	l.metr.RecordRollupRetry(*retry)
	if *retry > DaLoopRetryNum {
		l.log.Error("rollup failed by 10 attempts, need to re-store data to mantle da")
		*retry = 0
		l.state.params = nil
		l.state.initStoreDataReceipt = nil
		l.metr.RecordDaRetry(1)
		return true
	}
	time.Sleep(5 * time.Second)
	return true
}

func (l *BatchSubmitter) txAggregatorForEigenDa() ([][]byte, error) {
	var tempTxsData, txsData [][]byte
	var transactionByte []byte
	sortTxIds := make([]txID, 0, len(l.state.daPendingTxData))
	l.state.daUnConfirmedTxID = l.state.daUnConfirmedTxID[:0]
	for k := range l.state.daPendingTxData {
		sortTxIds = append(sortTxIds, k)
	}
	sort.Slice(sortTxIds, func(i, j int) bool {
		return sortTxIds[i].frameNumber < sortTxIds[j].frameNumber
	})
	for _, v := range sortTxIds {
		txData := l.state.daPendingTxData[v]
		tempTxsData = append(tempTxsData, txData.Bytes())
		txnBufBytes, err := rlp.EncodeToBytes(tempTxsData)
		if err != nil {
			l.log.Error("op-batcher unable to encode txn", "err", err)
			return nil, err
		}
		if uint64(len(txnBufBytes)) >= l.RollupMaxSize {
			l.log.Info("op-batcher transactionByte size is more than RollupMaxSize", "rollupMaxSize", l.RollupMaxSize, "txnBufBytes", len(txnBufBytes), "transactionByte", len(transactionByte))
			l.metr.RecordTxOverMaxLimit()
			break
		}
		txsData = tempTxsData
		transactionByte = txnBufBytes
		l.state.daUnConfirmedTxID = append(l.state.daUnConfirmedTxID, v)
		l.log.Info("added frame to daUnConfirmedTxID", "id", v.String())
	}

	if len(txsData) == 0 {
		l.log.Error("txsData is empty")
		return nil, fmt.Errorf("txsData is empty")
	}

	return txsData, nil
}

func (l *BatchSubmitter) recordConfirmedEigenDATx(receipt *types.Receipt) {
	l1block := eth.BlockID{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash}

	for _, id := range l.state.daUnConfirmedTxID {
		l.state.TxConfirmed(id, l1block)
		l.daTxDataConfirmed(id)
	}
	l.state.params = nil
	l.state.initStoreDataReceipt = nil
	l.state.metr.RecordRollupRetry(0)
	l.state.metr.RecordDaRetry(0)
}

func (l *BatchSubmitter) daTxDataConfirmed(id txID) {
	if _, ok := l.state.daPendingTxData[id]; !ok {
		l.log.Warn("daConfirmed, unknown txID of txData  marked as confirmed", "id", id.String())
		return
	}
	delete(l.state.daPendingTxData, id)
}

func (l *BatchSubmitter) reset() {
	l.state.clearPendingChannel()
	l.state.clearMantleDAStatus()
	l.lastStoredBlock = eth.BlockID{}
}
