package batcher

import (
	"context"
	"errors"
	"github.com/ethereum-optimism/optimism/l2geth/rlp"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"

	"github.com/Layr-Labs/datalayr/common/interfaces/interfaceDL"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"time"
)

func (l *BatchSubmitter) mantleDALoop() {
	defer l.wg.Done()
	ticker := time.NewTicker(l.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := l.loadBlocksIntoState(l.shutdownCtx); errors.Is(err, ErrReorg) {
				err := l.state.Close()
				if err != nil {
					l.log.Error("error closing the channel manager to handle a L2 reorg", "err", err)
				}
				l.publishStateToMantleDA()
				l.state.Clear()
				continue
			}
			l.publishStateToMantleDA()
		case <-l.shutdownCtx.Done():
			err := l.state.Close()
			if err != nil {
				l.log.Error("error closing the channel manager", "err", err)
			}
			l.publishStateToMantleDA()
			return
		}
	}
}

// publishStateToMantleDA loops through the block data loaded into `state` and
// submits the associated data to the MantleDA in the form of channel frames.
// batch frames in one rollup transaction to MantleDA
func (l *BatchSubmitter) publishStateToMantleDA() {
	txDone := make(chan struct{})
	// send/wait and receipt reading must be on a separate goroutines to avoid deadlocks
	go func() {
		defer func() {
			close(txDone)
		}()
		for {
			err := l.publishTxsToMantleDA(l.killCtx)
			if err != nil {
				if err != io.EOF {
					l.log.Error("error sending tx while draining state", "err", err)
				}
				return
			}
		}
	}()
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
				txsdata = append(txsdata,v.Bytes())
			}
			err := l.DisperseStoreData(txsdata)
			return err
		} else {
			l.log.Error("there is no frame in the current channel")
			return errors.New("there is no frame in the current channel")
		}
	}
	return nil
}

func (l *BatchSubmitter) handleInitDataStoreReceipt(r txmgr.TxReceipt[txsData]) {

}

func (l *BatchSubmitter) handleConfirmDataStoreReceipy()  {

}

func (l *BatchSubmitter) DisperseStoreData(txsdata [][]byte) error {
	txnBufBytes, err := rlp.EncodeToBytes(txsdata)
	if err != nil {
		l.log.Error("rlp unable to encode txn","err",err)
		return err
	}



	params, err := l.callEncode(txnBufBytes)
	if err != nil {
		return params, nil, err
	}
	uploadHeader, err := common2.CreateUploadHeader(params)
	if err != nil {
		return params, nil, err
	}
	log.Info("Operator Info", "NumSys", params.NumSys, "NumPar", params.NumPar, "TotalOperatorsIndex", params.TotalOperatorsIndex, "NumTotal", params.NumTotal)
	tx, err := d.StoreData(
		d.Ctx, uploadHeader, uint8(params.Duration), params.ReferenceBlockNumber,, params.TotalOperatorsIndex, isReRollup,
	)
	if err != nil {
		log.Error("MtBatcher StoreData tx", "err", err)
		return params, nil, err
	} else if tx == nil {
		return params, nil, errors.New("tx is nil")
	}
}

func (l *BatchSubmitter) callEncode(data []byte) (StoreParams, error)  {
	conn, err := grpc.Dial(l.DisperserSocket, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.log.Error("op-batcher disperser cannot connect to", "DisperserSocket", l.DisperserSocket)
		return StoreParams{}, err
	}
	defer conn.Close()
	c := pb.NewDataDispersalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(d.Cfg.DataStoreTimeout))
	defer cancel()
	request := &pb.EncodeStoreRequest{
		Duration: d.Cfg.DataStoreDuration,
		Data:     data,
	}
	opt := grpc.MaxCallSendMsgSize(1024 * 1024 * 300)
	reply, err := c.EncodeStore(ctx, request, opt)
	log.Info("MtBatcher get store", "reply", reply)
	if err != nil {
		log.Error("MtBatcher get store err", err)
		return common2.StoreParams{}, err
	}
	log.Info("MtBatcher get store end")
	g := reply.GetStore()
	feeBigInt := new(big.Int).SetBytes(g.Fee)
	params := common2.StoreParams{
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

