package sysext

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum-optimism/optimism/devnet-sdk/descriptors"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/shim"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-service/apis"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
)

// directEthFaucet implements apis.Faucet by sending ETH transfers directly
// via ethclient, bypassing txmgr to avoid nonce conflict issues in concurrent tests.
type directEthFaucet struct {
	client  *ethclient.Client
	privKey *ecdsa.PrivateKey
	chainID *big.Int
	from    common.Address
	log     log.Logger
}

func (f *directEthFaucet) ChainID(_ context.Context) (eth.ChainID, error) {
	return eth.ChainIDFromBig(f.chainID), nil
}

func (f *directEthFaucet) Balance(ctx context.Context) (eth.ETH, error) {
	bal, err := f.client.BalanceAt(ctx, f.from, nil)
	if err != nil {
		return eth.ETH{}, err
	}
	return eth.WeiBig(bal), nil
}

func (f *directEthFaucet) RequestETH(ctx context.Context, addr common.Address, amount eth.ETH) error {
	const maxRetries = 15

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := f.trySend(ctx, addr, amount.ToBig())
		if err == nil {
			return nil
		}
		if strings.Contains(err.Error(), "nonce too low") ||
			strings.Contains(err.Error(), "replacement transaction underpriced") ||
			strings.Contains(err.Error(), "already known") {
			f.log.Info("Tx conflict, retrying", "attempt", attempt+1, "err", err)
			continue
		}
		return err
	}
	return fmt.Errorf("failed to send ETH after %d retries", maxRetries)
}

func (f *directEthFaucet) trySend(ctx context.Context, to common.Address, value *big.Int) error {
	// Fresh nonce each attempt
	nonce, err := f.client.PendingNonceAt(ctx, f.from)
	if err != nil {
		return fmt.Errorf("get nonce: %w", err)
	}

	gasPrice, err := f.client.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("suggest gas price: %w", err)
	}
	// Bump gas price 20% to outbid concurrent txs
	gasPrice = new(big.Int).Mul(gasPrice, big.NewInt(120))
	gasPrice = new(big.Int).Div(gasPrice, big.NewInt(100))

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    value,
		Gas:      21000,
		GasPrice: gasPrice,
	})

	signer := types.NewEIP155Signer(f.chainID)
	signedTx, err := types.SignTx(tx, signer, f.privKey)
	if err != nil {
		return fmt.Errorf("sign tx: %w", err)
	}

	err = f.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	f.log.Info("Tx sent, waiting for receipt", "tx", signedTx.Hash(), "nonce", nonce, "to", to, "value", value)

	// Wait for receipt with dedicated timeout (not the caller's 30s context)
	receiptCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	for {
		receipt, err := f.client.TransactionReceipt(receiptCtx, signedTx.Hash())
		if err == nil {
			if receipt.Status == types.ReceiptStatusSuccessful {
				f.log.Info("Funded account", "tx", signedTx.Hash(), "to", to, "block", receipt.BlockNumber)
				return nil
			}
			return fmt.Errorf("tx %s reverted", signedTx.Hash())
		}
		select {
		case <-receiptCtx.Done():
			return fmt.Errorf("timeout waiting for receipt of tx %s", signedTx.Hash())
		case <-time.After(2 * time.Second):
		}
	}
}

// directFaucetWrapper wraps directEthFaucet as a stack.Faucet
type directFaucetWrapper struct {
	t      devtest.T
	logger log.Logger
	id     stack.FaucetID
	api    apis.Faucet
	labels map[string]string
}

var _ stack.Faucet = (*directFaucetWrapper)(nil)

func (d *directFaucetWrapper) T() devtest.T          { return d.t }
func (d *directFaucetWrapper) Logger() log.Logger     { return d.logger }
func (d *directFaucetWrapper) ID() stack.FaucetID     { return d.id }
func (d *directFaucetWrapper) API() apis.Faucet       { return d.api }
func (d *directFaucetWrapper) Label(key string) string { return d.labels[key] }
func (d *directFaucetWrapper) SetLabel(key, value string) {
	if d.labels == nil {
		d.labels = make(map[string]string)
	}
	d.labels[key] = value
}

// hydrateDirectFaucet creates a direct ethclient-based faucet that bypasses
// op-faucet/txmgr entirely, avoiding nonce conflict issues in concurrent tests.
func (o *Orchestrator) hydrateDirectFaucet(
	t devtest.T,
	net *descriptors.L2Chain,
	l2 stack.ExtensibleL2Network,
	commonConfig shim.CommonConfig,
	_ []client.RPCOption,
) {
	require := t.Require()

	// Try faucet wallet keys in priority order
	faucetWalletNames := []string{"l1MNTFaucet", "l1Faucet", "l2Faucet"}
	var faucetWalletKey string
	var faucetWalletName string
	for _, name := range faucetWalletNames {
		if wallet, ok := net.L1Wallets[name]; ok && wallet.PrivateKey != "" {
			faucetWalletKey = wallet.PrivateKey
			faucetWalletName = name
			break
		}
	}
	if faucetWalletKey == "" {
		t.Logger().Warn("No faucet wallet key found, skipping direct faucet setup")
		return
	}

	// Parse private key
	keyStr := strings.TrimPrefix(faucetWalletKey, "0x")
	keyBytes, err := hex.DecodeString(keyStr)
	require.NoError(err, "invalid faucet private key hex")
	privKey, err := crypto.ToECDSA(keyBytes)
	require.NoError(err, "invalid faucet private key")
	fromAddr := crypto.PubkeyToAddress(privKey.PublicKey)

	// Get L2 EL RPC endpoint
	elRPCURL := o.getFirstL2ELRPC(net)
	require.NotEmpty(elRPCURL, "no L2 EL RPC endpoint found for direct faucet")

	logger := t.Logger()
	logger.Info("Creating direct ethclient faucet",
		"wallet", faucetWalletName,
		"from", fromAddr,
		"elRPC", elRPCURL,
		"chainID", net.Config.ChainID)

	ethCl, err := ethclient.Dial(elRPCURL)
	require.NoError(err, "must connect to L2 EL for direct faucet")
	t.Cleanup(ethCl.Close)

	chainID := net.Config.ChainID
	faucetID := stack.NewFaucetID(fmt.Sprintf("direct-faucet-%s", chainID), eth.ChainIDFromBig(chainID))

	api := &directEthFaucet{
		client:  ethCl,
		privKey: privKey,
		chainID: chainID,
		from:    fromAddr,
		log:     logger,
	}

	l2.AddFaucet(&directFaucetWrapper{
		t:      t,
		logger: logger,
		id:     faucetID,
		api:    api,
	})

	logger.Info("Direct faucet ready", "faucetID", faucetID, "from", fromAddr)
}

// hasDirectFaucetKey checks if any known faucet wallet key exists in the descriptor.
func (o *Orchestrator) hasDirectFaucetKey(net *descriptors.L2Chain) bool {
	for _, name := range []string{"l1MNTFaucet", "l1Faucet", "l2Faucet"} {
		if wallet, ok := net.L1Wallets[name]; ok && wallet.PrivateKey != "" {
			return true
		}
	}
	return false
}

// getFirstL1ELRPC extracts the RPC URL of the first L1 EL node from the descriptor.
func (o *Orchestrator) getFirstL1ELRPC() string {
	if o.env == nil || o.env.Env == nil || o.env.Env.L1 == nil {
		return ""
	}
	for _, node := range o.env.Env.L1.Nodes {
		elService, ok := node.Services[ELServiceName]
		if !ok {
			continue
		}
		for proto, ep := range elService.Endpoints {
			if proto == RPCProtocol {
				port := ep.Port
				if o.usePrivatePorts {
					port = ep.PrivatePort
				}
				scheme := ep.Scheme
				if scheme == "" {
					scheme = HTTPProtocol
				}
				host := ep.Host
				path := ""
				if strings.Contains(host, "/") {
					parts := strings.SplitN(host, "/", 2)
					host = parts[0]
					path = "/" + parts[1]
				}
				if port != 0 {
					return fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
				}
				return fmt.Sprintf("%s://%s%s", scheme, host, path)
			}
		}
	}
	return ""
}

// getFirstL2ELRPC extracts the RPC URL of the first L2 EL node from the descriptor.
func (o *Orchestrator) getFirstL2ELRPC(net *descriptors.L2Chain) string {
	for _, node := range net.Nodes {
		elService, ok := node.Services[ELServiceName]
		if !ok {
			continue
		}
		for proto, ep := range elService.Endpoints {
			if proto == RPCProtocol {
				port := ep.Port
				if o.usePrivatePorts {
					port = ep.PrivatePort
				}
				scheme := ep.Scheme
				if scheme == "" {
					scheme = HTTPProtocol
				}
				host := ep.Host
				path := ""
				if strings.Contains(host, "/") {
					parts := strings.SplitN(host, "/", 2)
					host = parts[0]
					path = "/" + parts[1]
				}
				if port != 0 {
					return fmt.Sprintf("%s://%s:%d%s", scheme, host, port, path)
				}
				return fmt.Sprintf("%s://%s%s", scheme, host, path)
			}
		}
	}
	return ""
}
