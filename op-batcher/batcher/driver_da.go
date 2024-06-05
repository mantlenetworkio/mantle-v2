package batcher

import (
	"context"
	"errors"
	"io"
	"math/big"
	"sort"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	pb "github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"

	"github.com/ethereum/go-ethereum/accounts/abi"
	ecommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-batcher/common"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-node/eth"
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
			if l.Config.DaUpgradeChainConfig != nil && l.state.lastProcessedBlock.Number().Cmp(l.Config.DaUpgradeChainConfig.EigenDaUpgradeHeight) >= 0 {
				done, err = l.loopEigenDa()
			} else {
				done, err = l.loopRollupDa()
			}
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

// When the channel is full, it starts fetching data from the channel's cache(pendingTransactions) and rollup to MantleDA.
// A rollup consists of two transactions, InitDataStore and confirmStoredData, both of which are successful, meaning that the rollup is successful.
// A rollup has a RollupMaxSize limit, if a channel's corresponding txData is larger than this limit, you need to commit it in several times.
func (l *BatchSubmitter) loopRollupDa() (bool, error) {
	var retry int32
	l.metr.RecordRollupRetry(0)
	l.metr.RecordDaRetry(0)
	for {
		//it means that all the txData has been rollup
		if len(l.state.daPendingTxData) == 0 {
			l.log.Info("all txData of current channel have been rollup")
			return true, nil
		}
		//If current txsData has already been successfully uploaded to MantleDA(params is not nil), we don't need to re-upload.
		if l.state.params == nil {
			transactionData, err := l.txAggregator()
			if err != nil {
				l.log.Error("loopRollupDa txAggregator err,need to try again", "retry time", retry, "err", err)
				if l.isRetry(&retry) {
					continue
				}
				return false, err
			}
			err = l.disperseStoreData(transactionData)
			if err != nil {
				l.log.Error("loopRollupDa disperseStoreData err,need to try again", "retry time", retry, "err", err, "retry time", retry)
				if l.isRetry(&retry) {
					continue
				}
				return false, err
			}
		}

		sendTx := func() error {
			//start to publish transaction
			cCtx, cancel := context.WithTimeout(l.killCtx, 2*time.Minute)
			defer cancel()
			r, err := l.sendInitDataStoreTransaction(cCtx)
			if err != nil {
				l.log.Error("failed to send init datastore transaction,need to try again", "retry time", retry, "err", err)
				return err
			}

			receipt, err := l.handleInitDataStoreReceipt(cCtx, r)
			if err != nil {
				l.log.Error("failed to send confirm data transaction,need to try again", "retry time", retry, "err", err)
				return err
			}
			err = l.handleConfirmDataStoreReceipt(receipt)
			if err != nil {
				l.log.Error("failed to handle confirm data transaction receipt,need to try again", "retry time", retry, "err", err)
				return err
			}
			return nil
		}
		err := sendTx()
		if err != nil {
			if l.isRetry(&retry) {
				continue
			}
			return false, err
		}

	}
}

func (l *BatchSubmitter) loopEigenDa() (bool, error) {
	//it means that all the txData has been rollup
	if len(l.state.daPendingTxData) == 0 {
		l.log.Info("all txData of current channel have been rollup")
		return true, nil
	}

	var err error
	var candidate *txmgr.TxCandidate
	var receipt *types.Receipt
	var data []byte
	eigendaSuccess := false

	data, err = l.txAggregatorForEigenDa()
	if err != nil {
		l.log.Error("loopEigenDa txAggregatorForEigenDa err", "err", err)
		return false, err
	}

	currentL1, err := l.l1Tip(l.killCtx)
	if err != nil {
		l.log.Warn("loopEigenDa l1Tip", "err", err)
		return false, err
	}

	for loopRetry := 0; loopRetry < DaLoopRetryNum; loopRetry++ {
		err = func() error {
			l.metr.RecordRollupRetry(int32(loopRetry))
			if candidate == nil {
				//try 3 times
				for retry := 0; retry < EigenRPCRetryNum; retry++ {
					l.metr.RecordDaRetry(int32(retry))
					wrappedData, err := l.disperseEigenDaData(data)
					if err != nil {
						l.log.Warn("loopEigenDa disperseEigenDaData err,need to try again", "retry time", retry, "err", err)
						time.Sleep(5 * time.Second)
						continue
					}

					candidate = l.calldataTxCandidate(wrappedData)
					eigendaSuccess = true
					break
				}

				if !eigendaSuccess {
					if candidate, err = l.blobTxCandidate(data); err != nil {
						l.log.Warn("failed to create blob tx candidate", "err", err)
						return err
					}
				}
			}

			receipt, err = l.txMgr.Send(l.killCtx, *candidate)
			if err != nil {
				l.log.Warn("failed to send tx candidate", "err", err)
				return err
			}

			return nil
		}()
		if err != nil {
			l.log.Warn("failed to rollup", "err", err, "retry time", loopRetry)
		} else {
			break
		}
	}

	if err != nil {
		l.log.Error("failed to rollup", "err", err)
		return false, err
	}

	err = l.handleConfirmDataStoreReceipt(receipt)
	if err != nil {
		l.log.Error("failed to handleConfirmDataStoreReceipt", "err", err)
		return false, err
	}

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

func (l *BatchSubmitter) blobTxCandidate(data []byte) (*txmgr.TxCandidate, error) {
	l.log.Info("building Blob transaction candidate", "size", len(data))
	blobs := []*se.Blob{}
	for idx := 0; idx < len(data); idx += se.MaxBlobDataSize {
		blobData := data[idx : idx+minInt(len(data)-idx, se.MaxBlobDataSize)]
		var blob se.Blob
		if err := blob.FromData(blobData); err != nil {
			return nil, err
		}
		blobs = append(blobs, &blob)
	}

	return &txmgr.TxCandidate{
		To:    &l.Rollup.BatchInboxAddress,
		Blobs: blobs,
	}, nil
}

func (l *BatchSubmitter) disperseEigenDaData(data []byte) ([]byte, error) {
	blobInfo, requestId, err := l.eigenDA.DisperseBlob(l.shutdownCtx, data)
	if err != nil {
		l.log.Error("Unable to publish batch frameset to EigenDA", "err", err)
		return nil, err

	}

	quorumIDs := make([]uint32, len(blobInfo.BlobHeader.BlobQuorumParams))
	for i := range quorumIDs {
		quorumIDs[i] = blobInfo.BlobHeader.BlobQuorumParams[i].QuorumNumber
	}
	calldataFrame := &op_service.CalldataFrame{
		Value: &op_service.CalldataFrame_FrameRef{
			FrameRef: &op_service.FrameRef{
				BatchHeaderHash:      blobInfo.BlobVerificationProof.BatchMetadata.BatchHeaderHash,
				BlobIndex:            blobInfo.BlobVerificationProof.BlobIndex,
				ReferenceBlockNumber: blobInfo.BlobVerificationProof.BatchMetadata.BatchHeader.ReferenceBlockNumber,
				QuorumIds:            quorumIDs,
				BlobLength:           uint32(len(data)),
				RequestId:            requestId,
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

func (l *BatchSubmitter) txAggregator() ([]byte, error) {
	var txsData [][]byte
	var transactionByte []byte
	sortTxIds := make([]txID, 0, len(l.state.daPendingTxData))
	l.state.daUnConfirmedTxID = l.state.daUnConfirmedTxID[:0]
	for k, _ := range l.state.daPendingTxData {
		sortTxIds = append(sortTxIds, k)
	}
	sort.Slice(sortTxIds, func(i, j int) bool {
		return sortTxIds[i].frameNumber < sortTxIds[j].frameNumber
	})
	for _, v := range sortTxIds {
		txData, _ := l.state.daPendingTxData[v]
		txsData = append(txsData, txData.Bytes())
		txnBufBytes, err := rlp.EncodeToBytes(txsData)
		if err != nil {
			l.log.Error("op-batcher unable to encode txn", "err", err)
			return nil, err
		}
		if uint64(len(txnBufBytes)) >= l.RollupMaxSize {
			l.log.Info("op-batcher transactionByte size is more than RollupMaxSize", "rollupMaxSize", l.RollupMaxSize, "txnBufBytes", len(txnBufBytes), "transactionByte", len(transactionByte))
			l.metr.RecordTxOverMaxLimit()
			break
		}
		transactionByte = txnBufBytes
		l.state.daUnConfirmedTxID = append(l.state.daUnConfirmedTxID, v)
		l.log.Info("added frame to daUnConfirmedTxID", "id", v.String())
	}
	nodesNumber, err := l.getMantleDANodesNumber()
	if err != nil {
		l.log.Warn("op-batcher get nodes number failed", "err", err)
		nodesNumber = l.MantleDaNodes
	}
	l.log.Info("op-batcher transactionByte", "size", len(transactionByte))
	if len(transactionByte) <= BytesPerCoefficient*nodesNumber {
		paddingBytes := make([]byte, (BytesPerCoefficient*nodesNumber)-len(transactionByte))
		transactionByte = append(transactionByte, paddingBytes...)
	}
	return transactionByte, nil
}

func (l *BatchSubmitter) txAggregatorForEigenDa() ([]byte, error) {
	var txsData [][]byte
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
		txsData = append(txsData, txData.Bytes())
		txnBufBytes, err := rlp.EncodeToBytes(txsData)
		if err != nil {
			l.log.Error("op-batcher unable to encode txn", "err", err)
			return nil, err
		}
		if uint64(len(txnBufBytes)) >= l.RollupMaxSize {
			l.log.Info("op-batcher transactionByte size is more than RollupMaxSize", "rollupMaxSize", l.RollupMaxSize, "txnBufBytes", len(txnBufBytes), "transactionByte", len(transactionByte))
			l.metr.RecordTxOverMaxLimit()
			break
		}
		transactionByte = txnBufBytes
		l.state.daUnConfirmedTxID = append(l.state.daUnConfirmedTxID, v)
		l.log.Info("added frame to daUnConfirmedTxID", "id", v.String())
	}
	return transactionByte, nil
}

func (l *BatchSubmitter) disperseStoreData(txsData []byte) error {

	params, err := l.callEncode(txsData)
	if err != nil {
		return err
	}
	l.log.Info("operator info", "numSys", params.NumSys, "numPar", params.NumPar, "totalOperatorsIndex", params.TotalOperatorsIndex, "numTotal", params.NumTotal)
	//cache params
	l.state.params = params

	return nil
}

func (l *BatchSubmitter) sendInitDataStoreTransaction(ctx context.Context) (*types.Receipt, error) {
	//If initStoreData transaction has already been successfully executed, we don't need to re-execute .
	if l.state.initStoreDataReceipt != nil {
		l.log.Info("init store data transaction has been published successfully, skip to send transaction again")
		return l.state.initStoreDataReceipt, nil
	}
	uploadHeader, err := common.CreateUploadHeader(l.state.params)
	if err != nil {
		return nil, err
	}

	dataStoreTxData, err := l.dataStoreTxData(
		l.DataLayrServiceManagerABI, uploadHeader, uint8(l.state.params.Duration), l.state.params.ReferenceBlockNumber, l.state.params.TotalOperatorsIndex,
	)
	if err != nil {
		return nil, err
	}

	candidate := txmgr.TxCandidate{
		To:     &l.DataLayrServiceManagerAddr,
		TxData: dataStoreTxData,
	}
	receipt, err := l.txMgr.Send(ctx, candidate)
	if err != nil {
		return nil, err
	}
	l.metr.RecordBatchTxInitDataSubmitted()
	l.metr.RecordInitReferenceBlockNumber(l.state.params.ReferenceBlockNumber)
	return receipt, nil
}

func (l *BatchSubmitter) callEncode(data []byte) (*common.StoreParams, error) {
	conn, err := grpc.Dial(l.DisperserSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.log.Error("op-batcher disperser cannot connect to", "disperserSocket", l.DisperserSocket)
		return nil, err
	}
	defer conn.Close()
	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), l.DisperserTimeout)
	defer cancel()
	request := &pb.EncodeStoreRequest{
		Duration: l.DataStoreDuration,
		Data:     data,
	}
	opt := grpc.MaxCallSendMsgSize(EigenRollupMaxSize)
	reply, err := c.EncodeStore(ctx, request, opt)
	l.log.Info("op-batcher get store", "reply", reply.String())
	if err != nil {
		l.log.Error("op-batcher get store", "err", err)
		return nil, err
	}
	l.log.Info("op-batcher get store end")
	g := reply.GetStore()
	feeBigInt := new(big.Int).SetBytes(g.Fee)
	params := &common.StoreParams{
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

func (l *BatchSubmitter) dataStoreTxData(abi *abi.ABI, uploadHeader []byte, duration uint8, blockNumber uint32, totalOperatorsIndex uint32) ([]byte, error) {
	l.log.Info("encode initDataStore", "feePayer", l.txMgr.From(), "confirmor", l.txMgr.From(), "duration", duration, "referenceBlockNumber", blockNumber, "totalOperatorsIndex", totalOperatorsIndex)

	return abi.Pack(
		"initDataStore",
		l.txMgr.From(),
		l.txMgr.From(),
		duration,
		blockNumber,
		totalOperatorsIndex,
		uploadHeader)
}

func (l *BatchSubmitter) callDisperse(headerHash []byte, messageHash []byte) (*common.DisperseMeta, error) {
	conn, err := grpc.Dial(l.DisperserSocket, grpc.WithInsecure())
	if err != nil {
		l.log.Error("op-batcher dial disperserSocket", "err", err)
		return nil, err
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
		return nil, err
	}
	sigs := reply.GetSigs()
	aggSig := common.AggregateSignature{
		AggSig:            sigs.AggSig,
		StoredAggPubkeyG1: sigs.StoredAggPubkeyG1,
		UsedAggPubkeyG2:   sigs.UsedAggPubkeyG2,
		NonSignerPubkeys:  sigs.NonSignerPubkeys,
	}
	meta := &common.DisperseMeta{
		Sigs:            aggSig,
		ApkIndex:        reply.GetApkIndex(),
		TotalStakeIndex: reply.GetTotalStakeIndex(),
	}
	return meta, nil
}

func (l *BatchSubmitter) confirmStoredData(txHash []byte, ctx context.Context) (*types.Receipt, error) {
	event, ok := l.GraphClient.PollingInitDataStore(
		ctx,
		txHash[:],
		l.GraphPollingDuration,
	)
	if !ok {
		l.log.Error("op-batcher could not get initDataStore", "ok", ok)
		return nil, errors.New("op-batcher could not get initDataStore")
	}
	l.log.Info("PollingInitDataStore", "storeNumber", event.StoreNumber)
	meta, err := l.callDisperse(
		l.state.params.HeaderHash,
		event.MsgHash[:],
	)
	if err != nil {
		l.log.Error("op-batcher call disperse fail", "err", err)
		return nil, err
	}
	if len(meta.Sigs.NonSignerPubkeys) != 0 {
		l.log.Error("op-batcher call disperse success. However, there are nodes that do not participate in the signature.", "number", len(meta.Sigs.NonSignerPubkeys))
		l.metr.RecordDaNonSignerPubkeys(len(meta.Sigs.NonSignerPubkeys))
		return nil, errors.New("disperse meta nonSignerPubkeys is not 0")
	}

	callData, err := common.MakeCalldata(l.state.params, *meta, event.StoreNumber, event.MsgHash)
	if err != nil {
		l.log.Error("op-batcher make call data fail", "err", err)
		return nil, err
	}
	searchData := &bindings.IDataLayrServiceManagerDataStoreSearchData{
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

	confirmTxData, err := l.confirmDataTxData(l.DataLayrServiceManagerABI, callData, searchData)
	if err != nil {
		return nil, err
	}

	candidate := txmgr.TxCandidate{
		To:     &l.DataLayrServiceManagerAddr,
		TxData: confirmTxData,
	}

	txReceipt, err := l.txMgr.Send(ctx, candidate)
	if err != nil {
		log.Error("Tx manager send transaction fail", "err", err)
		return nil, err
	}
	l.metr.RecordConfirmedDataStoreId(event.StoreNumber)
	return txReceipt, nil
}

func (l *BatchSubmitter) confirmDataTxData(abi *abi.ABI, callData []byte, searchData *bindings.IDataLayrServiceManagerDataStoreSearchData) ([]byte, error) {
	return abi.Pack(
		"confirmDataStore",
		callData,
		searchData)
}

func (l *BatchSubmitter) handleInitDataStoreReceipt(ctx context.Context, txReceiptIn *types.Receipt) (*types.Receipt, error) {
	if txReceiptIn.Status == types.ReceiptStatusFailed {
		l.log.Error("init datastore tx successfully published but reverted", "txHash", txReceiptIn.TxHash.String())
		l.metr.RecordBatchTxInitDataFailed()
		return nil, ErrInitDataStore
	}
	l.metr.RecordBatchTxInitDataSuccess()
	l.log.Info("init datastore tx successfully published", "txHash", txReceiptIn.TxHash.String())
	l.state.initStoreDataReceipt = txReceiptIn
	// start to confirmData
	txReceiptOut, err := l.confirmStoredData(txReceiptIn.TxHash.Bytes(), ctx)
	if err != nil {
		l.log.Error("failed to confirm data", "err", err)
		return nil, err
	}
	l.metr.RecordBatchTxConfirmDataSubmitted()
	return txReceiptOut, nil

}

func (l *BatchSubmitter) handleConfirmDataStoreReceipt(r *types.Receipt) error {
	if r.Status == types.ReceiptStatusFailed {
		l.log.Error("unable to publish confirm data store tx", "txHash", r.TxHash.String())
		l.metr.RecordBatchTxConfirmDataFailed()
		return errors.New("unable to publish confirm data store tx")
	}
	l.log.Info("transaction confirmed", "txHash", r.TxHash.String(), "status", r.Status, "blockHash", r.BlockHash.String(), "blockNumber", r.BlockNumber)
	l.metr.RecordBatchTxConfirmDataSuccess()
	l.recordConfirmedEigenDATx(r)
	return nil
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

func (l *BatchSubmitter) getMantleDANodesNumber() (int, error) {
	operators, err := l.GraphClient.QueryOperatorsByStatus()
	if err != nil {
		l.log.Error("op-batcher query mantle da operators fail", "err", err)
		return 0, err
	}
	return len(operators), nil
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
