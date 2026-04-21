package backend

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/config"
	"github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/store"
	ftypes "github.com/ethereum-optimism/optimism/op-faucet/faucet/backend/types"
	"github.com/ethereum-optimism/optimism/op-faucet/faucet/frontend"
	"github.com/ethereum-optimism/optimism/op-faucet/metrics"
	"github.com/ethereum-optimism/optimism/op-service/apis"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/txmgr"
)

type Faucet struct {
	mu sync.RWMutex

	log log.Logger
	m   metrics.Metricer

	id       ftypes.FaucetID
	chainID  eth.ChainID
	txMgr    txmgr.TxManager
	elClient apis.EthBalance

	// true when the faucet is disabled and may not serve any new faucet requests
	disabled bool

	// balance monitoring
	alertThreshold       eth.ETH
	balanceCheckInterval time.Duration
	stopMonitor          chan struct{}
	larkWebhookURL       string
	chainName            string
	explorerURL          string
	wasLowBalance        bool // tracks last state for edge-triggered alerts

	// rate limiting
	store      *store.Store
	dailyLimit *big.Int
}

var _ frontend.FaucetBackend = (*Faucet)(nil)

// MantleGasPriceEstimatorFn is a custom gas price estimator for Mantle.
// Mantle does not support eth_blobBaseFee, so we skip it and return 0 for blob fee.
func MantleGasPriceEstimatorFn(ctx context.Context, backend txmgr.ETHBackend) (*big.Int, *big.Int, *big.Int, error) {
	tip, err := backend.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, nil, nil, err
	}

	head, err := backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, nil, nil, err
	}
	if head.BaseFee == nil {
		return nil, nil, nil, errors.New("txmgr does not support pre-london blocks that do not have a base fee")
	}

	return tip, head.BaseFee, big.NewInt(0), nil
}

func FaucetFromConfig(logger log.Logger, m metrics.Metricer, fID ftypes.FaucetID, fCfg *config.FaucetEntry) (*Faucet, error) {
	logger = logger.New("faucet", fID, "chain", fCfg.ChainID)
	txCfg, err := fCfg.TxManagerConfig(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to setup tx manager config: %w", err)
	}
	txCfg.GasPriceEstimatorFn = MantleGasPriceEstimatorFn
	txMgr, err := txmgr.NewSimpleTxManagerFromConfig(string(fID), logger, m, txCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to start tx manager: %w", err)
	}
	elClient, err := ethclient.Dial(fCfg.ELRPC.Value.RPC())
	if err != nil {
		return nil, fmt.Errorf("failed to dial EL client: %w", err)
	}

	checkInterval := 60 * time.Second
	if fCfg.BalanceCheckInterval > 0 {
		checkInterval = time.Duration(fCfg.BalanceCheckInterval) * time.Second
	}

	f := faucetWithTxManager(logger, m, fID, txMgr, elClient)
	f.alertThreshold = fCfg.AlertThreshold
	f.balanceCheckInterval = checkInterval
	f.stopMonitor = make(chan struct{})
	f.larkWebhookURL = fCfg.LarkWebhookURL
	f.chainName = fCfg.ChainName
	f.explorerURL = fCfg.ExplorerURL

	if f.alertThreshold.ToBig().Sign() > 0 {
		go f.monitorBalance()
	}

	return f, nil
}

func faucetWithTxManager(logger log.Logger, m metrics.Metricer, fID ftypes.FaucetID, txMgr txmgr.TxManager, elClient apis.EthBalance) *Faucet {
	return &Faucet{
		log:      logger,
		m:        m,
		id:       fID,
		chainID:  txMgr.ChainID(),
		txMgr:    txMgr,
		elClient: elClient,
		disabled: false,
	}
}

func (f *Faucet) Enable() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.log.Info("Enabling faucet")
	f.disabled = false
}

func (f *Faucet) Disable() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.log.Info("Disabling faucet")
	f.disabled = true
}

func (f *Faucet) Close() {
	f.log.Info("Closing faucet")
	if f.stopMonitor != nil {
		close(f.stopMonitor)
	}
	f.Disable()
	f.txMgr.Close()
}

func (f *Faucet) monitorBalance() {
	f.log.Info("Starting balance monitor",
		"threshold", f.alertThreshold.String(),
		"interval", f.balanceCheckInterval)

	ticker := time.NewTicker(f.balanceCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-f.stopMonitor:
			f.log.Info("Stopping balance monitor")
			return
		case <-ticker.C:
			balance, err := f.Balance()
			if err != nil {
				f.log.Warn("Balance monitor failed to get balance", "err", err)
				continue
			}

			f.m.RecordBalance(f.id, f.chainID, balance.WeiFloat())

			isLow := balance.ToBig().Cmp(f.alertThreshold.ToBig()) < 0

			// Edge-triggered: only alert on state transitions
			if isLow && !f.wasLowBalance {
				f.m.RecordLowBalance(f.id, f.chainID)
				f.log.Error("ALERT: Faucet balance is below threshold, please replenish",
					"balance", balance.String(),
					"threshold", f.alertThreshold.String(),
					"wallet", f.txMgr.From())
				f.sendLarkAlert(f.buildAlertMessage(true, balance))
			} else if !isLow && f.wasLowBalance {
				f.log.Info("Faucet balance recovered above threshold",
					"balance", balance.String(),
					"threshold", f.alertThreshold.String())
				f.sendLarkAlert(f.buildAlertMessage(false, balance))
			}
			f.wasLowBalance = isLow
		}
	}
}

func (f *Faucet) buildAlertMessage(isLow bool, balance eth.ETH) string {
	chainName := f.chainName
	if chainName == "" {
		chainName = f.chainID.String()
	}

	wallet := f.txMgr.From().Hex()
	walletLine := wallet
	if f.explorerURL != "" {
		walletLine = fmt.Sprintf("%s (%s/address/%s)", wallet, f.explorerURL, wallet)
	}

	// Convert wei to human-readable MNT (divide by 10^18)
	balanceMNT := weiToMNT(balance.ToBig())
	thresholdMNT := weiToMNT(f.alertThreshold.ToBig())

	if isLow {
		return fmt.Sprintf(
			"🚨 Faucet Low Balance Alert\n"+
				"Faucet: %s\n"+
				"Chain: %s (chainID=%s)\n"+
				"Wallet: %s\n"+
				"Balance: %s MNT\n"+
				"Threshold: %s MNT\n"+
				"Please replenish the faucet wallet!",
			f.id, chainName, f.chainID, walletLine, balanceMNT, thresholdMNT)
	}
	return fmt.Sprintf(
		"✅ Faucet Balance Recovered\n"+
			"Faucet: %s\n"+
			"Chain: %s (chainID=%s)\n"+
			"Wallet: %s\n"+
			"Balance: %s MNT",
		f.id, chainName, f.chainID, walletLine, balanceMNT)
}

func weiToMNT(wei *big.Int) string {
	f := new(big.Float).SetInt(wei)
	f.Quo(f, new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)))
	return f.Text('f', 4)
}

func (f *Faucet) sendLarkAlert(message string) {
	if f.larkWebhookURL == "" {
		return
	}

	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": message,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		f.log.Warn("Failed to marshal Lark payload", "err", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", f.larkWebhookURL, bytes.NewReader(body))
	if err != nil {
		f.log.Warn("Failed to build Lark request", "err", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		f.log.Warn("Failed to send Lark alert", "err", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		f.log.Warn("Lark webhook returned non-2xx", "status", resp.StatusCode)
	}
}

func (f *Faucet) Balance() (eth.ETH, error) {
	wallet := f.txMgr.From()
	balance, err := f.elClient.BalanceAt(context.Background(), wallet, nil)

	var ethBalance eth.ETH
	if err != nil {
		f.log.Error("Failed to get balance", "err", err)
		return ethBalance, err
	}
	return eth.WeiBig(balance), nil
}

func (f *Faucet) ChainID() eth.ChainID {
	return f.chainID
}

// Register registers a new user. Returns true if newly registered, false if already exists.
func (f *Faucet) Register(addr common.Address) (bool, error) {
	if f.store == nil {
		return false, errors.New("store not configured")
	}
	return f.store.RegisterUser(addr)
}

// Eligibility checks whether the user can claim and returns their remaining quota.
func (f *Faucet) Eligibility(addr common.Address) (*ftypes.EligibilityResult, error) {
	result := &ftypes.EligibilityResult{}

	if f.store == nil || f.dailyLimit == nil {
		result.Eligible = true
		result.Registered = true
		result.DailyLimit = "0"
		result.DailyClaimed = "0"
		result.Remaining = "0"
		return result, nil
	}

	registered, err := f.store.IsRegistered(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to check registration: %w", err)
	}
	result.Registered = registered

	if !registered {
		result.Eligible = false
		result.DailyLimit = f.dailyLimit.String()
		result.DailyClaimed = "0"
		result.Remaining = "0"
		return result, nil
	}

	claimed, err := f.store.DailyClaimedAmount(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to check daily claims: %w", err)
	}

	remaining := new(big.Int).Sub(f.dailyLimit, claimed)
	if remaining.Sign() < 0 {
		remaining = big.NewInt(0)
	}

	result.DailyLimit = f.dailyLimit.String()
	result.DailyClaimed = claimed.String()
	result.Remaining = remaining.String()
	result.Eligible = remaining.Sign() > 0

	return result, nil
}

func (f *Faucet) RequestMNT(ctx context.Context, request *ftypes.FaucetRequest) (result error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	logger := f.log.New("to", request.Target, "amount", request.Amount)
	if f.disabled {
		logger.Info("Cannot serve request, faucet is disabled")
		return errors.New("faucet is disabled")
	}

	logger.Info("Sending funds")

	// Check user registration and daily limit
	if f.store != nil && f.dailyLimit != nil {
		registered, err := f.store.IsRegistered(request.Target)
		if err != nil {
			logger.Error("Failed to check user registration", "err", err)
			return fmt.Errorf("failed to check registration: %w", err)
		}
		if !registered {
			logger.Info("Rejecting unregistered user")
			return errors.New("user not registered, please register first via faucet_register")
		}

		claimed, err := f.store.DailyClaimedAmount(request.Target)
		if err != nil {
			logger.Error("Failed to check daily claims", "err", err)
			return fmt.Errorf("failed to check daily claims: %w", err)
		}
		newTotal := new(big.Int).Add(claimed, request.Amount.ToBig())
		if newTotal.Cmp(f.dailyLimit) > 0 {
			logger.Info("Daily limit exceeded",
				"claimed", claimed.String(),
				"requested", request.Amount.String(),
				"limit", f.dailyLimit.String())
			return fmt.Errorf("daily limit exceeded: already claimed %s wei today, limit is %s wei",
				claimed.String(), f.dailyLimit.String())
		}
	}

	balance, err := f.Balance()
	if err != nil {
		logger.Warn("Failed to get balance, optimistically continuing the request")
	} else {
		if balance.ToBig().Cmp(request.Amount.ToBig()) < 0 {
			logger.Error("Insufficient balance", "balance", balance.String(), "amount", request.Amount)
			return errors.New("insufficient balance")
		}
	}

	onDone := f.m.RecordFundAction(f.id, f.chainID, request.Amount)
	defer func() {
		onDone(result)
	}()

	// We execute this tiny special EVM program,
	// such that we can move ETH into the target account,
	// without executing the code of the target account.
	// Since we don't want to accidentally trigger untrusted EVM functionality as the funder EOA.

	// This code self-destructs the ephemeral contract-creation-scope,
	// and assigns (no execution!) all the value of this scope to the target address.
	// These types of ephemeral self-destructs are still allowed post-Cancun.
	var out []byte
	out = append(out, byte(vm.PUSH20))
	out = append(out, request.Target[:]...)
	out = append(out, byte(vm.SELFDESTRUCT))

	candidate := txmgr.TxCandidate{
		TxData:   out,
		Blobs:    nil,
		To:       nil, // contract-creation, see above
		GasLimit: 0,   // estimate gas dynamically
		Value:    request.Amount.ToBig(),
	}
	rec, err := f.txMgr.Send(ctx, candidate)
	if err != nil {
		logger.Error("failed to send funds", "err", err)
		return fmt.Errorf("failed to send funds: %w", err)
	}
	if rec.Status == types.ReceiptStatusFailed {
		logger.Error("funding tx reverted", "tx", rec.TxHash)
		return fmt.Errorf("failed to fund, tx %s reverted", rec.TxHash)
	}
	logger.Info("Successfully funded account",
		"tx", rec.TxHash,
		"included_hash", rec.BlockHash,
		"included_num", rec.BlockNumber)

	// Record claim after successful transfer
	if f.store != nil {
		if err := f.store.RecordClaim(request.Target, request.Amount.ToBig()); err != nil {
			logger.Warn("Failed to record claim in store", "err", err)
		}
	}

	return nil
}
