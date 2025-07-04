package txmgr

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/txpool"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/log"
	"github.com/holiman/uint256"

	"github.com/ethereum-optimism/optimism/op-service/txmgr/metrics"
)

const (
	// geth requires a minimum fee bump of 10% for regular tx resubmission
	priceBump int64 = 10
	// geth requires a minimum fee bump of 100% for blob tx resubmission
	blobPriceBump int64 = 100

	Wei   = 1
	GWei  = 1e9
	Ether = 1e18
)

// geth enforces a 1 gwei minimum for blob tx fee
var (
	priceBumpPercent     = big.NewInt(100 + priceBump)
	blobPriceBumpPercent = big.NewInt(100 + blobPriceBump)

	// geth enforces a 1 gwei minimum for blob tx fee
	minBlobTxFee = big.NewInt(GWei)

	oneHundred = big.NewInt(100)
	ninetyNine = big.NewInt(99)
	two        = big.NewInt(2)

	ErrBlobFeeLimit = errors.New("blob fee limit reached")
)

// TxManager is an interface that allows callers to reliably publish txs,
// bumping the gas price if needed, and obtain the receipt of the resulting tx.
//
//go:generate mockery --name TxManager --output ./mocks
type TxManager interface {
	// Send is used to create & send a transaction. It will handle increasing
	// the gas price & ensuring that the transaction remains in the transaction pool.
	// It can be stopped by cancelling the provided context; however, the transaction
	// may be included on L1 even if the context is cancelled.
	//
	// NOTE: Send can be called concurrently, the nonce will be managed internally.
	Send(ctx context.Context, candidate TxCandidate) (*types.Receipt, error)

	// From returns the sending address associated with the instance of the transaction manager.
	// It is static for a single instance of a TxManager.
	From() common.Address
}

// ETHBackend is the set of methods that the transaction manager uses to resubmit gas & determine
// when transactions are included on L1.
type ETHBackend interface {
	// BlockNumber returns the most recent block number.
	BlockNumber(ctx context.Context) (uint64, error)

	// TransactionReceipt queries the backend for a receipt associated with
	// txHash. If lookup does not fail, but the transaction is not found,
	// nil should be returned for both values.
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)

	// SendTransaction submits a signed transaction to L1.
	SendTransaction(ctx context.Context, tx *types.Transaction) error

	// These functions are used to estimate what the basefee & priority fee should be set to.
	// TODO(CLI-3318): Maybe need a generic interface to support different RPC providers
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	// NonceAt returns the account nonce of the given account.
	// The block number can be nil, in which case the nonce is taken from the latest known block.
	NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error)
	// PendingNonceAt returns the pending nonce.
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	// EstimateGas returns an estimate of the amount of gas needed to execute the given
	// transaction against the current pending block.
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
}

// SimpleTxManager is a implementation of TxManager that performs linear fee
// bumping of a tx until it confirms.
type SimpleTxManager struct {
	cfg     Config // embed the config directly
	name    string
	chainID *big.Int

	backend ETHBackend
	l       log.Logger
	metr    metrics.TxMetricer

	nonce     *uint64
	nonceLock sync.RWMutex

	pending atomic.Int64
}

// NewSimpleTxManager initializes a new SimpleTxManager with the passed Config.
func NewSimpleTxManager(name string, l log.Logger, m metrics.TxMetricer, cfg CLIConfig) (*SimpleTxManager, error) {
	conf, err := NewConfig(cfg, l)
	if err != nil {
		return nil, err
	}

	return &SimpleTxManager{
		chainID: conf.ChainID,
		name:    name,
		cfg:     conf,
		backend: conf.Backend,
		l:       l.New("service", name),
		metr:    m,
	}, nil
}

func (m *SimpleTxManager) From() common.Address {
	return m.cfg.From
}

// TxCandidate is a transaction candidate that can be submitted to ask the
// [TxManager] to construct a transaction with gas price bounds.
type TxCandidate struct {
	// TxData is the transaction data to be used in the constructed tx.
	TxData []byte
	// To is the recipient of the constructed tx. Nil means contract creation.
	To *common.Address
	// GasLimit is the gas limit to be used in the constructed tx.
	GasLimit uint64
	// Blobs to send along in the tx (optional). If len(Blobs) > 0 then a blob tx
	// will be sent instead of a DynamicFeeTx.
	Blobs []*eth.Blob
}

// Send is used to publish a transaction with incrementally higher gas prices
// until the transaction eventually confirms. This method blocks until an
// invocation of sendTx returns (called with differing gas prices). The method
// may be canceled using the passed context.
//
// The transaction manager handles all signing. If and only if the gas limit is 0, the
// transaction manager will do a gas estimation.
//
// NOTE: Send can be called concurrently, the nonce will be managed internally.
func (m *SimpleTxManager) Send(ctx context.Context, candidate TxCandidate) (*types.Receipt, error) {
	m.metr.RecordPendingTx(m.pending.Add(1))
	defer func() {
		m.metr.RecordPendingTx(m.pending.Add(-1))
	}()
	receipt, err := m.send(ctx, candidate)
	if err != nil {
		m.resetNonce()
	}
	return receipt, err
}

// send performs the actual transaction creation and sending.
func (m *SimpleTxManager) send(ctx context.Context, candidate TxCandidate) (*types.Receipt, error) {
	if m.cfg.TxSendTimeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, m.cfg.TxSendTimeout)
		defer cancel()
	}
	tx, err := m.craftTx(ctx, candidate)
	if err != nil {
		return nil, fmt.Errorf("failed to create the tx: %w", err)
	}
	return m.sendTx(ctx, tx)
}

// craftTx creates the signed transaction
// It queries L1 for the current fee market conditions as well as for the nonce.
// NOTE: This method SHOULD NOT publish the resulting transaction.
// NOTE: If the [TxCandidate.GasLimit] is non-zero, it will be used as the transaction's gas.
// NOTE: Otherwise, the [SimpleTxManager] will query the specified backend for an estimate.
func (m *SimpleTxManager) craftTx(ctx context.Context, candidate TxCandidate) (*types.Transaction, error) {
	gasTipCap, basefee, blobBaseFee, err := m.suggestGasPriceCaps(ctx)
	if err != nil {
		m.metr.RPCError()
		return nil, fmt.Errorf("failed to get gas price info: %w", err)
	}
	gasFeeCap := calcGasFeeCap(basefee, gasTipCap)

	nonce, err := m.nextNonce(ctx)
	if err != nil {
		return nil, err
	}

	rawTx := &types.DynamicFeeTx{
		ChainID:   m.chainID,
		Nonce:     nonce,
		To:        candidate.To,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Data:      candidate.TxData,
	}

	m.l.Info("creating tx", "to", rawTx.To, "from", m.cfg.From)

	// If the gas limit is set, we can use that as the gas
	if candidate.GasLimit != 0 {
		rawTx.Gas = candidate.GasLimit
	} else {
		// Calculate the intrinsic gas for the transaction
		gas, err := m.backend.EstimateGas(ctx, ethereum.CallMsg{
			From:      m.cfg.From,
			To:        candidate.To,
			GasFeeCap: gasFeeCap,
			GasTipCap: gasTipCap,
			Data:      rawTx.Data,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas: %w", err)
		}
		rawTx.Gas = gas
	}

	var sidecar *types.BlobTxSidecar
	var blobHashes []common.Hash
	if len(candidate.Blobs) > 0 {
		if candidate.To == nil {
			return nil, errors.New("blob txs cannot deploy contracts")
		}
		if sidecar, blobHashes, err = MakeSidecar(candidate.Blobs); err != nil {
			return nil, fmt.Errorf("failed to make sidecar: %w", err)
		}
	}

	var txMessage types.TxData
	if sidecar != nil {
		if blobBaseFee == nil {
			return nil, fmt.Errorf("expected non-nil blobBaseFee")
		}
		blobFeeCap := calcBlobFeeCap(blobBaseFee)
		message := &types.BlobTx{
			To:         *candidate.To,
			Data:       candidate.TxData,
			Gas:        rawTx.Gas,
			BlobHashes: blobHashes,
			Sidecar:    sidecar,
			Nonce:      nonce,
		}
		if err := finishBlobTx(message, m.chainID, gasTipCap, gasFeeCap, blobFeeCap, nil); err != nil {
			return nil, fmt.Errorf("failed to create blob transaction: %w", err)
		}
		txMessage = message
	} else {
		txMessage = rawTx
	}

	ctx, cancel := context.WithTimeout(ctx, m.cfg.NetworkTimeout)
	defer cancel()
	return m.cfg.Signer(ctx, m.cfg.From, types.NewTx(txMessage))
}

// nextNonce returns a nonce to use for the next transaction. It uses
// eth_getTransactionCount with "latest" once, and then subsequent calls simply
// increment this number. If the transaction manager is reset, it will query the
// eth_getTransactionCount nonce again.
func (m *SimpleTxManager) nextNonce(ctx context.Context) (uint64, error) {
	m.nonceLock.Lock()
	defer m.nonceLock.Unlock()

	if m.nonce == nil {
		// Fetch the sender's nonce from the latest known block (nil `blockNumber`)
		childCtx, cancel := context.WithTimeout(ctx, m.cfg.NetworkTimeout)
		defer cancel()
		nonce, err := m.backend.NonceAt(childCtx, m.cfg.From, nil)
		if err != nil {
			m.metr.RPCError()
			return 0, fmt.Errorf("failed to get nonce: %w", err)
		}
		m.nonce = &nonce
	} else {
		*m.nonce++
	}

	m.metr.RecordNonce(*m.nonce)
	return *m.nonce, nil
}

// resetNonce resets the internal nonce tracking. This is called if any pending send
// returns an error.
func (m *SimpleTxManager) resetNonce() {
	m.nonceLock.Lock()
	defer m.nonceLock.Unlock()
	m.nonce = nil
}

// send submits the same transaction several times with increasing gas prices as necessary.
// It waits for the transaction to be confirmed on chain.
func (m *SimpleTxManager) sendTx(ctx context.Context, tx *types.Transaction) (*types.Receipt, error) {
	var wg sync.WaitGroup
	defer wg.Wait()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sendState := NewSendState(m.cfg.SafeAbortNonceTooLowCount, m.cfg.TxNotInMempoolTimeout)
	receiptChan := make(chan *types.Receipt, 1)
	sendTxAsync := func(tx *types.Transaction) {
		defer wg.Done()
		m.publishAndWaitForTx(ctx, tx, sendState, receiptChan)
	}

	// Immediately publish a transaction before starting the resumbission loop
	wg.Add(1)
	go sendTxAsync(tx)

	ticker := time.NewTicker(m.cfg.ResubmissionTimeout)
	defer ticker.Stop()

	bumpCounter := 0
	for {
		select {
		case <-ticker.C:
			// Don't resubmit a transaction if it has been mined, but we are waiting for the conf depth.
			if sendState.IsWaitingForConfirmation() {
				continue
			}
			// If we see lots of unrecoverable errors (and no pending transactions) abort sending the transaction.
			if sendState.ShouldAbortImmediately() {
				m.l.Warn("Aborting transaction submission")
				return nil, errors.New("aborted transaction sending")
			}
			// Increase the gas price & submit the new transaction
			tx = m.increaseGasPrice(ctx, tx)
			wg.Add(1)
			bumpCounter += 1
			go sendTxAsync(tx)

		case <-ctx.Done():
			return nil, ctx.Err()

		case receipt := <-receiptChan:
			m.metr.RecordGasBumpCount(bumpCounter)
			m.metr.TxConfirmed(receipt)
			return receipt, nil
		}
	}
}

// publishAndWaitForTx publishes the transaction to the transaction pool and then waits for it with [waitMined].
// It should be called in a new go-routine. It will send the receipt to receiptChan in a non-blocking way if a receipt is found
// for the transaction.
func (m *SimpleTxManager) publishAndWaitForTx(ctx context.Context, tx *types.Transaction, sendState *SendState, receiptChan chan *types.Receipt) {
	log := m.l.New("hash", tx.Hash(), "nonce", tx.Nonce(), "gasTipCap", tx.GasTipCap(), "gasFeeCap", tx.GasFeeCap())
	log.Info("publishing transaction")

	cCtx, cancel := context.WithTimeout(ctx, m.cfg.NetworkTimeout)
	defer cancel()
	t := time.Now()
	err := m.backend.SendTransaction(cCtx, tx)
	sendState.ProcessSendError(err)

	// Properly log & exit if there is an error
	if err != nil {
		switch {
		case errStringMatch(err, core.ErrNonceTooLow):
			log.Warn("nonce too low", "err", err)
			m.metr.TxPublished("nonce_to_low")
		case errStringMatch(err, context.Canceled):
			m.metr.RPCError()
			log.Warn("transaction send cancelled", "err", err)
			m.metr.TxPublished("context_cancelled")
		case errStringMatch(err, txpool.ErrAlreadyKnown):
			log.Warn("resubmitted already known transaction", "err", err)
			m.metr.TxPublished("tx_already_known")
		case errStringMatch(err, txpool.ErrReplaceUnderpriced):
			log.Warn("transaction replacement is underpriced", "err", err)
			m.metr.TxPublished("tx_replacement_underpriced")
		case errStringMatch(err, txpool.ErrUnderpriced):
			log.Warn("transaction is underpriced", "err", err)
			m.metr.TxPublished("tx_underpriced")
		default:
			m.metr.RPCError()
			log.Error("unable to publish transaction", "err", err)
			m.metr.TxPublished("unknown_error")
		}
		return
	}
	m.metr.TxPublished("")

	log.Info("Transaction successfully published")
	// Poll for the transaction to be ready & then send the result to receiptChan
	receipt, err := m.waitMined(ctx, tx, sendState)
	if err != nil {
		log.Warn("Transaction receipt not found", "err", err)
		return
	}
	select {
	case receiptChan <- receipt:
		m.metr.RecordTxConfirmationLatency(time.Since(t).Milliseconds())
	default:
	}
}

// waitMined waits for the transaction to be mined or for the context to be cancelled.
func (m *SimpleTxManager) waitMined(ctx context.Context, tx *types.Transaction, sendState *SendState) (*types.Receipt, error) {
	txHash := tx.Hash()
	queryTicker := time.NewTicker(m.cfg.ReceiptQueryInterval)
	defer queryTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-queryTicker.C:
			if receipt := m.queryReceipt(ctx, txHash, sendState); receipt != nil {
				return receipt, nil
			}
		}
	}
}

// queryReceipt queries for the receipt and returns the receipt if it has passed the confirmation depth
func (m *SimpleTxManager) queryReceipt(ctx context.Context, txHash common.Hash, sendState *SendState) *types.Receipt {
	ctx, cancel := context.WithTimeout(ctx, m.cfg.NetworkTimeout)
	defer cancel()
	receipt, err := m.backend.TransactionReceipt(ctx, txHash)
	if errors.Is(err, ethereum.NotFound) {
		sendState.TxNotMined(txHash)
		m.l.Trace("Transaction not yet mined", "hash", txHash)
		return nil
	} else if err != nil {
		m.metr.RPCError()
		m.l.Info("Receipt retrieval failed", "hash", txHash, "err", err)
		return nil
	} else if receipt == nil {
		m.metr.RPCError()
		m.l.Warn("Receipt and error are both nil", "hash", txHash)
		return nil
	}

	// Receipt is confirmed to be valid from this point on
	sendState.TxMined(txHash)

	txHeight := receipt.BlockNumber.Uint64()
	tipHeight, err := m.backend.BlockNumber(ctx)
	if err != nil {
		m.l.Error("Unable to fetch block number", "err", err)
		return nil
	}

	m.l.Debug("Transaction mined, checking confirmations", "hash", txHash, "txHeight", txHeight,
		"tipHeight", tipHeight, "numConfirmations", m.cfg.NumConfirmations)

	// The transaction is considered confirmed when
	// txHeight+numConfirmations-1 <= tipHeight. Note that the -1 is
	// needed to account for the fact that confirmations have an
	// inherent off-by-one, i.e. when using 1 confirmation the
	// transaction should be confirmed when txHeight is equal to
	// tipHeight. The equation is rewritten in this form to avoid
	// underflows.
	if txHeight+m.cfg.NumConfirmations <= tipHeight+1 {
		m.l.Info("Transaction confirmed", "hash", txHash)
		return receipt
	}

	// Safe to subtract since we know the LHS above is greater.
	confsRemaining := (txHeight + m.cfg.NumConfirmations) - (tipHeight + 1)
	m.l.Debug("Transaction not yet confirmed", "hash", txHash, "confsRemaining", confsRemaining)
	return nil
}

func (m *SimpleTxManager) checkBlobFeeLimits(blobBaseFee, bumpedBlobFee *big.Int) error {
	// If below threshold, don't apply multiplier limit. Note we use same threshold parameter here
	// used for non-blob fee limiting.
	if thr := m.cfg.FeeLimitThreshold; thr != nil && thr.Cmp(bumpedBlobFee) == 1 {
		return nil
	}
	maxBlobFee := new(big.Int).Mul(calcBlobFeeCap(blobBaseFee), big.NewInt(int64(m.cfg.FeeLimitMultiplier)))
	if bumpedBlobFee.Cmp(maxBlobFee) > 0 {
		return fmt.Errorf(
			"bumped blob fee %v is over %dx multiple of the suggested value: %w",
			bumpedBlobFee, m.cfg.FeeLimitMultiplier, ErrBlobFeeLimit)
	}
	return nil
}

// increaseGasPrice takes the previous transaction & potentially clones then signs it with a higher tip.
// If the tip + basefee suggested by the network are not greater than the previous values, the same transaction
// will be returned. If they are greater, this function will ensure that they are at least greater by 15% than
// the previous transaction's value to ensure that the price bump is large enough.
//
// We do not re-estimate the amount of gas used because for some stateful transactions (like output proposals) the
// act of including the transaction renders the repeat of the transaction invalid.
//
// If it encounters an error with creating the new transaction, it will return the old transaction.
func (m *SimpleTxManager) increaseGasPrice(ctx context.Context, tx *types.Transaction) *types.Transaction {
	tip, basefee, blobBaseFee, err := m.suggestGasPriceCaps(ctx)
	if err != nil {
		m.l.Warn("failed to get suggested gas tip and basefee", "err", err)
		return tx
	}
	gasTipCap, gasFeeCap := updateFees(tx.GasTipCap(), tx.GasFeeCap(), tip, basefee, tx.Type() == types.BlobTxType, m.l)

	if tx.GasTipCapIntCmp(gasTipCap) == 0 && tx.GasFeeCapIntCmp(gasFeeCap) == 0 {
		return tx
	}

	// Re-estimate gaslimit in case things have changed or a previous gaslimit estimate was wrong
	gas, err := m.backend.EstimateGas(ctx, ethereum.CallMsg{
		From:      m.cfg.From,
		To:        tx.To(),
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Data:      tx.Data(),
		Value:     tx.Value(),
	})
	if err != nil {
		// If this is a transaction resubmission, we sometimes see this outcome because the
		// original tx can get included in a block just before the above call. In this case the
		// error is due to the tx reverting with message "block number must be equal to next
		// expected block number"
		m.l.Warn("failed to re-estimate gas", "err", err, "tx", tx.Hash(), "gaslimit", tx.Gas(),
			"gasFeeCap", gasFeeCap, "gasTipCap", gasTipCap)
		return tx
	}

	var newTx *types.Transaction
	if tx.Type() == types.BlobTxType {
		// Blob transactions have an additional blob gas price we must specify, so we must make sure it is
		// getting bumped appropriately.
		bumpedBlobFee := calcThresholdValue(tx.BlobGasFeeCap(), true)
		if bumpedBlobFee.Cmp(blobBaseFee) < 0 {
			bumpedBlobFee = blobBaseFee
		}
		if err := m.checkBlobFeeLimits(blobBaseFee, bumpedBlobFee); err != nil {
			m.l.Warn("failed to check blob fee limits", "err", err)
			return tx
		}
		message := &types.BlobTx{
			Nonce:      tx.Nonce(),
			To:         *tx.To(),
			Data:       tx.Data(),
			Gas:        gas,
			BlobHashes: tx.BlobHashes(),
			Sidecar:    tx.BlobTxSidecar(),
		}
		if err := finishBlobTx(message, tx.ChainId(), gasTipCap, gasFeeCap, bumpedBlobFee, tx.Value()); err != nil {
			m.l.Warn("failed to finish blob tx", "err", err)
			return tx
		}
		newTx = types.NewTx(message)
	} else {
		newTx = types.NewTx(&types.DynamicFeeTx{
			ChainID:    tx.ChainId(),
			Nonce:      tx.Nonce(),
			GasTipCap:  gasTipCap,
			GasFeeCap:  gasFeeCap,
			Gas:        gas,
			To:         tx.To(),
			Value:      tx.Value(),
			Data:       tx.Data(),
			AccessList: tx.AccessList(),
		})
	}

	ctx, cancel := context.WithTimeout(ctx, m.cfg.NetworkTimeout)
	defer cancel()
	signedTx, err := m.cfg.Signer(ctx, m.cfg.From, newTx)
	if err != nil {
		m.l.Warn("failed to sign new transaction", "err", err)
		return tx
	}
	return signedTx
}

// suggestGasPriceCaps suggests what the new tip & new basefee should be based on the current L1 conditions
func (m *SimpleTxManager) suggestGasPriceCaps(ctx context.Context) (*big.Int, *big.Int, *big.Int, error) {
	cCtx, cancel := context.WithTimeout(ctx, m.cfg.NetworkTimeout)
	defer cancel()
	tip, err := m.backend.SuggestGasTipCap(cCtx)
	if err != nil {
		m.metr.RPCError()
		return nil, nil, nil, fmt.Errorf("failed to fetch the suggested gas tip cap: %w", err)
	} else if tip == nil {
		return nil, nil, nil, errors.New("the suggested tip was nil")
	}
	cCtx, cancel = context.WithTimeout(ctx, m.cfg.NetworkTimeout)
	defer cancel()
	head, err := m.backend.HeaderByNumber(cCtx, nil)
	if err != nil {
		m.metr.RPCError()
		return nil, nil, nil, fmt.Errorf("failed to fetch the suggested basefee: %w", err)
	} else if head.BaseFee == nil {
		return nil, nil, nil, errors.New("txmgr does not support pre-london blocks that do not have a basefee")
	}

	var blobFee *big.Int
	if head.ExcessBlobGas != nil {
		blobFee = eth.CalcBlobFeeDefault(head)
	}

	return tip, head.BaseFee, blobFee, nil
}

// calcThresholdValue returns ceil(x * priceBumpPercent / 100) for non-blob txs, or
// ceil(x * blobPriceBumpPercent / 100) for blob txs.
// It guarantees that x is increased by at least 1
func calcThresholdValue(x *big.Int, isBlobTx bool) *big.Int {
	threshold := new(big.Int)
	if isBlobTx {
		threshold.Set(blobPriceBumpPercent)
	} else {
		threshold.Set(priceBumpPercent)
	}
	return threshold.Mul(threshold, x).Add(threshold, ninetyNine).Div(threshold, oneHundred)
}

// updateFees takes the old tip/basefee & the new tip/basefee and then suggests
// a gasTipCap and gasFeeCap that satisfies geth's required fee bumps
// Geth: FC and Tip must be bumped if any increase
func updateFees(oldTip, oldFeeCap, newTip, newBaseFee *big.Int, isBlobTx bool, lgr log.Logger) (*big.Int, *big.Int) {
	newFeeCap := calcGasFeeCap(newBaseFee, newTip)
	lgr = lgr.New("old_tip", oldTip, "old_feecap", oldFeeCap, "new_tip", newTip, "new_feecap", newFeeCap)
	// If the new prices are less than the old price, reuse the old prices
	if oldTip.Cmp(newTip) >= 0 && oldFeeCap.Cmp(newFeeCap) >= 0 {
		lgr.Debug("Reusing old tip and feecap")
		return oldTip, oldFeeCap
	}
	// Determine if we need to increase the suggested values
	thresholdTip := calcThresholdValue(oldTip, isBlobTx)
	thresholdFeeCap := calcThresholdValue(oldFeeCap, isBlobTx)
	if newTip.Cmp(thresholdTip) >= 0 && newFeeCap.Cmp(thresholdFeeCap) >= 0 {
		lgr.Debug("Using new tip and feecap")
		return newTip, newFeeCap
	} else if newTip.Cmp(thresholdTip) >= 0 && newFeeCap.Cmp(thresholdFeeCap) < 0 {
		// Tip has gone up, but basefee is flat or down.
		// TODO(CLI-3714): Do we need to recalculate the FC here?
		lgr.Debug("Using new tip and threshold feecap")
		return newTip, thresholdFeeCap
	} else if newTip.Cmp(thresholdTip) < 0 && newFeeCap.Cmp(thresholdFeeCap) >= 0 {
		// Basefee has gone up, but the tip hasn't. Recalculate the feecap because if the tip went up a lot
		// not enough of the feecap may be dedicated to paying the basefee.
		lgr.Debug("Using threshold tip and recalculated feecap")
		return thresholdTip, calcGasFeeCap(newBaseFee, thresholdTip)

	} else {
		// TODO(CLI-3713): Should we skip the bump in this case?
		lgr.Debug("Using threshold tip and threshold feecap")
		return thresholdTip, thresholdFeeCap
	}
}

// calcGasFeeCap deterministically computes the recommended gas fee cap given
// the base fee and gasTipCap. The resulting gasFeeCap is equal to:
//
//	gasTipCap + 2*baseFee.
func calcGasFeeCap(baseFee, gasTipCap *big.Int) *big.Int {
	return new(big.Int).Add(
		gasTipCap,
		new(big.Int).Mul(baseFee, big.NewInt(2)),
	)
}

// errStringMatch returns true if err.Error() is a substring in target.Error() or if both are nil.
// It can accept nil errors without issue.
func errStringMatch(err, target error) bool {
	if err == nil && target == nil {
		return true
	} else if err == nil || target == nil {
		return false
	}
	return strings.Contains(err.Error(), target.Error())
}

// MakeSidecar builds & returns the BlobTxSidecar and corresponding blob hashes from the raw blob
// data.
func MakeSidecar(blobs []*eth.Blob) (*types.BlobTxSidecar, []common.Hash, error) {
	sidecar := &types.BlobTxSidecar{}
	blobHashes := []common.Hash{}
	for i, blob := range blobs {
		rawBlob := blob.KZGBlob()
		sidecar.Blobs = append(sidecar.Blobs, *rawBlob)
		commitment, err := kzg4844.BlobToCommitment(rawBlob)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot compute KZG commitment of blob %d in tx candidate: %w", i, err)
		}
		sidecar.Commitments = append(sidecar.Commitments, commitment)
		proof, err := kzg4844.ComputeBlobProof(rawBlob, commitment)
		if err != nil {
			return nil, nil, fmt.Errorf("cannot compute KZG proof for fast commitment verification of blob %d in tx candidate: %w", i, err)
		}
		sidecar.Proofs = append(sidecar.Proofs, proof)
		blobHashes = append(blobHashes, eth.KZGToVersionedHash(commitment))
	}
	return sidecar, blobHashes, nil
}

// calcBlobFeeCap computes a suggested blob fee cap that is twice the current header's blob base fee
// value, with a minimum value of minBlobTxFee.
func calcBlobFeeCap(blobBaseFee *big.Int) *big.Int {
	cap := new(big.Int).Mul(blobBaseFee, two)
	if cap.Cmp(minBlobTxFee) < 0 {
		cap.Set(minBlobTxFee)
	}
	return cap
}

func finishBlobTx(message *types.BlobTx, chainID, tip, fee, blobFee, value *big.Int) error {
	var o bool
	if message.ChainID, o = uint256.FromBig(chainID); o {
		return fmt.Errorf("ChainID overflow")
	}
	if message.GasTipCap, o = uint256.FromBig(tip); o {
		return fmt.Errorf("GasTipCap overflow")
	}
	if message.GasFeeCap, o = uint256.FromBig(fee); o {
		return fmt.Errorf("GasFeeCap overflow")
	}
	if message.BlobFeeCap, o = uint256.FromBig(blobFee); o {
		return fmt.Errorf("BlobFeeCap overflow")
	}
	if message.Value, o = uint256.FromBig(value); o {
		return fmt.Errorf("Value overflow")
	}
	return nil
}
