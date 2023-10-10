package batcher

import (
	"context"
	"errors"
	"io"
	"math/big"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"

	"github.com/ethereum-optimism/optimism/l2geth/rlp"
	"github.com/ethereum-optimism/optimism/op-batcher/common"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const RollupMaxSize = 1024 * 1024 * 300

var ErrUploadDataFinished = errors.New("data has been upload to MantleDA nodes")
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
				continue
			} else if err != nil {
				l.log.Error("load block into state err,", "err", err)
				continue
			}
			l.publishStateToMantleDA()

			if l.state.params != nil {
				//start to publish transaction
				cCtx, cancel := context.WithTimeout(l.killCtx, 2*time.Minute)
				r, err := l.sendInitDataStoreTransaction(cCtx)
				if err != nil {
					l.log.Error("Failed to send init datastore transaction", "err", err)
					cancel()
					continue
				}
				receipt, err := l.handleInitDataStoreReceipt(r, cCtx)
				if err != nil {
					cancel()
					continue
				}
				l.handleConfirmDataStoreReceipt(receipt)
				cancel()
			}

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
		err := l.publishTxsToMantleDA(l.killCtx)
		if err != nil {
			if errors.Is(err, ErrUploadDataFinished) {
				l.log.Info("init data store transaction has been published")

			} else if err != io.EOF {
				l.log.Error("error sending tx while draining state", "err", err)
			}
			return
		}
	}

}

func (l *BatchSubmitter) publishTxsToMantleDA(ctx context.Context) error {
	// send all available transactions
	l1tip, err := l.l1Tip(ctx)
	if err != nil {
		l.log.Error("Failed to query L1 tip", "error", err)
		return err
	}
	l.recordL1Tip(l1tip)

	// Collect next transaction data
	_, err = l.state.TxData(l1tip.ID())

	if l.state.pendingChannel != nil && l.state.pendingChannel.IsFull() && !l.state.pendingChannel.HasFrame() {
		if len(l.state.pendingTransactions) > 0 {
			var txsdata [][]byte
			for _, v := range l.state.pendingTransactions {
				txsdata = append(txsdata, v.Bytes())
			}
			err := l.DisperseStoreData(txsdata)
			return err
		} else {
			l.log.Error("there is no frame in the current channel")
			return errors.New("there is no frame in the current channel")
		}
	}
	return err
}

func (l *BatchSubmitter) DisperseStoreData(txsdata [][]byte) error {

	//if txsdata has been successfully upload to MantleDA, we don't need to re-upload.
	if l.state.params == nil {
		txnBufBytes, err := rlp.EncodeToBytes(txsdata)
		if err != nil {
			l.log.Error("rlp unable to encode txn", "err", err)
			return err
		}
		l.log.Info("start to upload data to MantleDA node, ", "len", len(txnBufBytes))

		params, err := l.callEncode(txnBufBytes)
		if err != nil {
			return err
		}
		l.log.Info("Operator Info", "NumSys", params.NumSys, "NumPar", params.NumPar, "TotalOperatorsIndex", params.TotalOperatorsIndex, "NumTotal", params.NumTotal)
		//cache params
		l.state.params = &params
	}

	return ErrUploadDataFinished
}

func (l *BatchSubmitter) sendInitDataStoreTransaction(ctx context.Context) (*types.Receipt, error) {
	//if initStoreData transaction has been successfully executed.We don't need to re-execute .
	if l.state.initStoreDataReceipt != nil {
		l.log.Info("init store data transaction has been published successfully, skip to send transaction again")
		return l.state.initStoreDataReceipt, nil
	}
	uploadHeader, err := common.CreateUploadHeader(l.state.params)
	if err != nil {
		return nil, err
	}

	txdata, err := l.DataStoreTxData(
		l.DataLayrServiceManagerABI, uploadHeader, uint8(l.state.params.Duration), l.state.params.ReferenceBlockNumber, l.state.params.TotalOperatorsIndex,
	)
	if err != nil {
		return nil, err
	}

	candiddate := txmgr.TxCandidate{
		To:     &l.DataLayrServiceManagerAddr,
		TxData: txdata,
	}
	receipt, err := l.txMgr.Send(ctx, candiddate)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (l *BatchSubmitter) callEncode(data []byte) (common.StoreParams, error) {
	conn, err := grpc.Dial(l.DisperserSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.log.Error("op-batcher disperser cannot connect to", "DisperserSocket", l.DisperserSocket)
		return common.StoreParams{}, err
	}
	defer conn.Close()
	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), l.DisperserTimeout)
	defer cancel()
	request := &pb.EncodeStoreRequest{
		Duration: l.DataStoreDuration,
		Data:     data,
	}
	opt := grpc.MaxCallSendMsgSize(RollupMaxSize)
	reply, err := c.EncodeStore(ctx, request, opt)
	l.log.Info("op-batcher get store", "reply", reply)
	if err != nil {
		l.log.Error("op-batcher get store err", err)
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

func (l *BatchSubmitter) DataStoreTxData(abi *abi.ABI, uploadHeader []byte, duration uint8, blockNumber uint32, totalOperatorsIndex uint32) ([]byte, error) {
	l.log.Info("encode initDataStore", "feePayer", l.txMgr.From(), "confirmer", l.txMgr.From(), "duration", duration, "referenceBlockNumber", blockNumber, "totalOperatorsIndex", totalOperatorsIndex)

	return abi.Pack(
		"initDataStore",
		l.txMgr.From(),
		l.txMgr.From(),
		duration,
		blockNumber,
		totalOperatorsIndex,
		uploadHeader)
}

func (l *BatchSubmitter) callDisperse(headerHash []byte, messageHash []byte) (common.DisperseMeta, error) {
	conn, err := grpc.Dial(l.DisperserSocket, grpc.WithInsecure())
	if err != nil {
		l.log.Error("op-batcher Dial DisperserSocket", "err", err)
		return common.DisperseMeta{}, err
	}
	defer conn.Close()
	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), l.DisperserTimeout)
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

func (l *BatchSubmitter) ConfirmStoredData(txHash []byte, ctx context.Context) (*types.Receipt, error) {
	event, ok := l.GraphClient.PollingInitDataStore(
		ctx,
		txHash[:],
		l.GraphPollingDuration,
	)
	if !ok {
		l.log.Error("op-batcher could not get initDataStore", "ok", ok)
		return nil, errors.New("op-batcher could not get initDataStore")
	}
	l.log.Info("PollingInitDataStore", "MsgHash", event.MsgHash, "StoreNumber", event.StoreNumber)
	meta, err := l.callDisperse(
		l.state.params.HeaderHash,
		event.MsgHash[:],
	)
	if err != nil {
		l.log.Error("op-batcher call Disperse fail", "err", err)
		return nil, err
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

	txdata, err := l.ConfirmDataTxData(l.DataLayrServiceManagerABI, callData, searchData)
	if err != nil {
		return nil, err
	}

	candiddate := txmgr.TxCandidate{
		To:     &l.DataLayrServiceManagerAddr,
		TxData: txdata,
	}
	return l.txMgr.Send(ctx, candiddate)
}

func (l *BatchSubmitter) ConfirmDataTxData(abi *abi.ABI, callData []byte, searchData bindings.IDataLayrServiceManagerDataStoreSearchData) ([]byte, error) {
	return abi.Pack(
		"confirmDataStore",
		callData,
		searchData)
}

func (l *BatchSubmitter) handleInitDataStoreReceipt(r *types.Receipt, ctx context.Context) (*types.Receipt, error) {
	if r.Status == types.ReceiptStatusFailed {
		l.log.Error("init datastore tx successfully published but reverted", "tx_hash", r.TxHash)
		l.recordFailedEigenDATx()
		return nil, ErrInitDataStore
	} else {
		l.log.Info("initDataStore tx successfully published", "tx_hash", r.TxHash)
		l.state.initStoreDataReceipt = r
		//start to confirmData
		r, err := l.ConfirmStoredData(r.TxHash.Bytes(), ctx)
		if err != nil {
			l.log.Error("failed to confirm data", "err", err)
			return nil, err
		}
		return r, nil
	}
}

func (l *BatchSubmitter) handleConfirmDataStoreReceipt(r *types.Receipt) {
	if r.Status == types.ReceiptStatusFailed {
		l.log.Error("unable to publish confirm data store tx", "tx_hash", r.TxHash)
		l.recordFailedEigenDATx()
	} else {
		l.log.Info("Transaction confirmed", "tx_hash", r.TxHash, "status", r.Status, "block_hash", r.BlockHash, "block_number", r.BlockNumber)
		l.recordConfirmedEigenDATx(r)
	}
}
func (l *BatchSubmitter) recordFailedEigenDATx() {
	for k, _ := range l.state.pendingTransactions {
		l.state.TxFailed(k)
	}
}

func (l *BatchSubmitter) recordConfirmedEigenDATx(receipt *types.Receipt) {
	l1block := eth.BlockID{Number: receipt.BlockNumber.Uint64(), Hash: receipt.BlockHash}

	for k, _ := range l.state.pendingTransactions {
		l.state.TxConfirmed(k, l1block)
	}
	l.state.clearMantleDAStatus()
}
