package batcher

import (
	"context"
	"errors"
	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"github.com/ethereum-optimism/optimism/l2geth/rlp"
	"github.com/ethereum-optimism/optimism/op-batcher/common"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"math/big"
	"time"
)

const ROLLUP_MAX_SIZE_ = 1024 * 1024 * 300

var ErrInitDataStoreDone = errors.New("init data store transaction done")
var ErrConfirmDataStoreDone = errors.New("confirm data store transaction done")

func (l *BatchSubmitter) mantleDALoop() {
	defer l.wg.Done()
	ticker := time.NewTicker(l.PollInterval)
	defer ticker.Stop()

	dataStoreReceiptCh := make(chan txmgr.TxReceipt[string])
	dataStoreQueue := txmgr.NewQueue[string](l.killCtx, l.txMgr, 1)
	confirmDataReceiptCh := make(chan txmgr.TxReceipt[string])
	confirmDataQueue := txmgr.NewQueue[string](l.killCtx, l.txMgr, 1)

	for {
		select {
		case <-ticker.C:
			if err := l.loadBlocksIntoState(l.shutdownCtx); errors.Is(err, ErrReorg) {
				err := l.state.Close()
				if err != nil {
					l.log.Error("error closing the channel manager to handle a L2 reorg", "err", err)
				}
				l.state.Clear()
				continue
			}
			l.publishStateToMantleDA(dataStoreQueue, dataStoreReceiptCh)
		case r := <-dataStoreReceiptCh:
			l.handleInitDataStoreReceipt(r, confirmDataQueue, confirmDataReceiptCh)
		case r := <-confirmDataReceiptCh:
			l.handleConfirmDataStoreReceipt(r)
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
func (l *BatchSubmitter) publishStateToMantleDA(queue *txmgr.Queue[string], receiptCh chan txmgr.TxReceipt[string]) {

	for {
		err := l.publishTxsToMantleDA(l.killCtx, queue, receiptCh)
		if err != nil {
			if err != io.EOF || !errors.Is(err, ErrInitDataStoreDone) {
				l.log.Error("error sending tx while draining state", "err", err)
			}
			return
		}
	}

}

func (l *BatchSubmitter) publishTxsToMantleDA(ctx context.Context, queue *txmgr.Queue[string], receiptCh chan txmgr.TxReceipt[string]) error {
	// send all available transactions
	l1tip, err := l.l1Tip(ctx)
	if err != nil {
		l.log.Error("Failed to query L1 tip", "error", err)
		return err
	}
	l.recordL1Tip(l1tip)

	// Collect next transaction data
	_, err = l.state.TxData(l1tip.ID())
	if err == io.EOF && !l.state.pendingChannel.IsFull() {
		l.log.Trace("no transaction data available")
		return err
	} else if err != nil {
		l.log.Error("unable to get tx data", "err", err)
		return err
	}
	if l.state.pendingChannel != nil && l.state.pendingChannel.IsFull() && l.state.pendingChannel.NumFrames() == 0 {
		if len(l.state.pendingTransactions) > 0 {
			var txsdata [][]byte
			for _, v := range l.state.pendingTransactions {
				txsdata = append(txsdata, v.Bytes())
			}
			err := l.DisperseStoreData(txsdata, queue, receiptCh)
			return err
		} else {
			l.log.Error("there is no frame in the current channel")
			return errors.New("there is no frame in the current channel")
		}
	}
	return nil
}

func (l *BatchSubmitter) DisperseStoreData(txsdata [][]byte, queue *txmgr.Queue[string], receiptCh chan txmgr.TxReceipt[string]) error {
	txnBufBytes, err := rlp.EncodeToBytes(txsdata)
	if err != nil {
		l.log.Error("rlp unable to encode txn", "err", err)
		return err
	}

	params, err := l.callEncode(txnBufBytes)
	if err != nil {
		return err
	}
	uploadHeader, err := common.CreateUploadHeader(params)
	if err != nil {
		return err
	}
	l.log.Info("Operator Info", "NumSys", params.NumSys, "NumPar", params.NumPar, "TotalOperatorsIndex", params.TotalOperatorsIndex, "NumTotal", params.NumTotal)
	//cache params
	l.state.params = params

	txdata, err := l.DataStoreTxData(
		uploadHeader, uint8(params.Duration), params.ReferenceBlockNumber, params.TotalOperatorsIndex,
	)
	if err != nil {
		return err
	}
	intrinsicGas, err := core.IntrinsicGas(txdata, nil, false, true, true, false)
	if err != nil {
		return err
	}
	candiddate := txmgr.TxCandidate{
		To:       &l.Rollup.DataLayerChainAddress,
		TxData:   txdata,
		GasLimit: intrinsicGas,
	}
	queue.Send("initDataStore", candiddate, receiptCh)
	return ErrInitDataStoreDone
}

func (l *BatchSubmitter) callEncode(data []byte) (common.StoreParams, error) {
	conn, err := grpc.Dial(l.DisperserSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.log.Error("op-batcher disperser cannot connect to", "DisperserSocket", l.DisperserSocket)
		return common.StoreParams{}, err
	}
	defer conn.Close()
	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(l.DisperserTimeout))
	defer cancel()
	request := &pb.EncodeStoreRequest{
		Duration: l.DataStoreDuration,
		Data:     data,
	}
	opt := grpc.MaxCallSendMsgSize(1024 * 1024 * 300)
	reply, err := c.EncodeStore(ctx, request, opt)
	l.log.Info("op-batcher get store", "reply", reply)
	if err != nil {
		l.log.Error("MtBatcher get store err", err)
		return common.StoreParams{}, err
	}
	l.log.Info("op-batcher get store end")
	g := reply.GetStore()
	feeBigInt := new(big.Int).SetBytes(g.Fee)
	params := common.StoreParams{
		ReferenceBlockNumber: g.ReferenceBlockNumber,
		TotalOperatorsIndex:  g.TotalOperatorsIndex,
		OrigDataSize:         g.OrigDataSize,
		NumTotal:             g.NumTotal,
		Quorum:               g.Quorum,
		NumSys:               g.NumSys,
		NumPar:               g.NumPar,
		Duration:             g.Duration,
		KzgCommit:            g.KzgCommit,
		LowDegreeProof:       g.LowDegreeProof,
		Degree:               g.Degree,
		TotalSize:            g.TotalSize,
		Order:                g.Order,
		Fee:                  feeBigInt,
		HeaderHash:           g.HeaderHash,
		Disperser:            g.Disperser,
	}
	return params, nil
}

func (l *BatchSubmitter) DataStoreTxData(uploadHeader []byte, duration uint8, blockNumber uint32, totalOperatorsIndex uint32) ([]byte, error) {
	initDataStoreTxData, err := l.DatalayrABI.Pack(
		"initDataStore",
		l.txMgr.From(),
		l.txMgr.From(),
		duration,
		blockNumber,
		totalOperatorsIndex,
		uploadHeader,
	)
	if err != nil {
		return nil, err
	}
	return initDataStoreTxData, nil

}

func (l *BatchSubmitter) callDisperse(headerHash []byte, messageHash []byte) (common.DisperseMeta, error) {
	conn, err := grpc.Dial(l.DisperserSocket, grpc.WithInsecure())
	if err != nil {
		l.log.Error("op-batcher Dial DisperserSocket", "err", err)
		return common.DisperseMeta{}, err
	}
	defer conn.Close()
	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(l.DisperserTimeout))
	defer cancel()
	request := &pb.DisperseStoreRequest{
		HeaderHash:  headerHash,
		MessageHash: messageHash,
	}
	reply, err := c.DisperseStore(ctx, request)
	if err != nil {
		return common.DisperseMeta{}, err
	}
	sigs := reply.GetSigs()
	aggSig := common.AggregateSignature{
		AggSig:            sigs.AggSig,
		StoredAggPubkeyG1: sigs.StoredAggPubkeyG1,
		UsedAggPubkeyG2:   sigs.UsedAggPubkeyG2,
		NonSignerPubkeys:  sigs.NonSignerPubkeys,
	}
	meta := common.DisperseMeta{
		Sigs:            aggSig,
		ApkIndex:        reply.GetApkIndex(),
		TotalStakeIndex: reply.GetTotalStakeIndex(),
	}
	return meta, nil
}

func (l *BatchSubmitter) ConfirmStoredData(txHash []byte, queue *txmgr.Queue[string], receiptCh chan txmgr.TxReceipt[string]) error {
	event, ok := l.GraphClient.PollingInitDataStore(
		l.killCtx,
		txHash[:],
		l.GraphPollingDuration,
	)
	if !ok {
		l.log.Error("op-batcher could not get initDataStore", "ok", ok)
		return errors.New("op-batcher could not get initDataStore")
	}
	l.log.Info("PollingInitDataStore", "MsgHash", event.MsgHash, "StoreNumber", event.StoreNumber)
	meta, err := l.callDisperse(
		l.state.params.HeaderHash,
		event.MsgHash[:],
	)
	if err != nil {
		l.log.Error("op-batcher call Disperse fail", "err", err)
		return err
	}
	callData := common.MakeCalldata(l.state.params, meta, event.StoreNumber, event.MsgHash)
	searchData := bindings.IDataLayrServiceManagerDataStoreSearchData{
		Duration:  event.Duration,
		Timestamp: new(big.Int).SetUint64(uint64(event.InitTime)),
		Index:     event.Index,
		Metadata: bindings.IDataLayrServiceManagerDataStoreMetadata{
			HeaderHash:           event.DataCommitment,
			DurationDataStoreId:  event.DurationDataStoreId,
			GlobalDataStoreId:    event.StoreNumber,
			ReferenceBlockNumber: event.ReferenceBlockNumber,
			BlockNumber:          uint32(event.InitBlockNumber.Uint64()),
			Fee:                  event.Fee,
			Confirmer:            ecommon.HexToAddress(event.Confirmer),
			SignatoryRecordHash:  [32]byte{},
		},
	}

	txdata, err := l.ConfirmDataTxData(callData, searchData)
	if err != nil {
		return err
	}
	intrinsicGas, err := core.IntrinsicGas(txdata, nil, false, true, true, false)
	if err != nil {
		return err
	}
	candiddate := txmgr.TxCandidate{
		To:       &l.Rollup.DataLayerChainAddress,
		TxData:   txdata,
		GasLimit: intrinsicGas,
	}
	queue.Send("initDataStore", candiddate, receiptCh)
	return ErrConfirmDataStoreDone

}

func (l *BatchSubmitter) ConfirmDataTxData(callData []byte, searchData bindings.IDataLayrServiceManagerDataStoreSearchData) ([]byte, error) {
	confirmDataTxData, err := l.DatalayrABI.Pack(
		"confirmDataStore",
		callData,
		searchData,
	)
	if err != nil {
		return nil, err
	}
	return confirmDataTxData, nil

}

func (l *BatchSubmitter) handleInitDataStoreReceipt(r txmgr.TxReceipt[string], queue *txmgr.Queue[string], receiptCh chan txmgr.TxReceipt[string]) {
	if r.Err != nil {
		l.log.Warn("unable to publish init data store tx", "err", r.Err)
		l.recordFailedEigenDATx(r.Err)
	} else {
		l.log.Info("initDataStore tx successfully published", "tx_hash", r.Receipt.TxHash)
		//start to confirmData
		err := l.ConfirmStoredData(r.Receipt.TxHash.Bytes(), queue, receiptCh)
		if err != nil {
			l.log.Error("failed to confirm data", "err", err)
		}
	}
}

func (l *BatchSubmitter) handleConfirmDataStoreReceipt(r txmgr.TxReceipt[string]) {
	if r.Err != nil {
		l.log.Warn("unable to publish confirm data store tx", "err", r.Err)
		l.recordFailedEigenDATx(r.Err)
	} else {
		l.log.Info("Transaction confirmed", "tx_hash", r.Receipt.TxHash, "status", r.Receipt.Status, "block_hash", r.Receipt.BlockHash, "block_number", r.Receipt.BlockNumber)
		l.recordConfirmedEigenDATx(r.Receipt)
	}
}
func (l *BatchSubmitter) recordFailedEigenDATx(err error) {
	l.log.Warn("Failed to send transaction", "err", err)
	for k, _ := range l.state.pendingTransactions {
		l.state.TxFailed(k)
	}
}

func (l *BatchSubmitter) recordConfirmedEigenDATx(receipt *types.Receipt) {
	l1block := eth.BlockID{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash}

	for k, _ := range l.state.pendingTransactions {
		l.state.TxConfirmed(k, l1block)
	}
}
