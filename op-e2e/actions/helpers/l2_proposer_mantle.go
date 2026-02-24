package helpers

import (
	"crypto/ecdsa"
	"math/big"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-e2e/bindings"
	"github.com/ethereum-optimism/optimism/op-e2e/config"
	mantlebindings "github.com/ethereum-optimism/optimism/op-e2e/mantlebindings/bindings"
	"github.com/ethereum-optimism/optimism/op-proposer/metrics"
	"github.com/ethereum-optimism/optimism/op-proposer/proposer"
	"github.com/ethereum-optimism/optimism/op-proposer/proposer/source"
	"github.com/ethereum-optimism/optimism/op-service/dial"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/sources"
	"github.com/ethereum-optimism/optimism/op-service/sources/batching"
)

type MantleL2Proposer struct {
	log                    log.Logger
	l1                     *ethclient.Client
	driver                 *proposer.L2OutputSubmitter
	disputeGameFactory     *bindings.DisputeGameFactoryCaller
	l2OutputOracle         *mantlebindings.L2OutputOracleCaller
	l2OutputOracleAddr     *common.Address
	disputeGameFactoryAddr *common.Address
	address                common.Address
	privKey                *ecdsa.PrivateKey
	lastTx                 common.Hash
	allocType              config.AllocType
}
type MantleProposerCfg struct {
	DisputeGameFactoryAddr *common.Address
	ProposalInterval       time.Duration
	ProposalRetryInterval  time.Duration
	DisputeGameType        uint32
	OutputOracleAddr       *common.Address
	ProposerKey            *ecdsa.PrivateKey
	AllowNonFinalized      bool
	AllocType              config.AllocType
	ChainID                eth.ChainID
}

func NewMantleL2Proposer(t Testing, log log.Logger, cfg *MantleProposerCfg, l1 *ethclient.Client, rollupCl *sources.RollupClient) *MantleL2Proposer {
	proposerConfig := proposer.ProposerConfig{
		PollInterval:           time.Second,
		NetworkTimeout:         time.Second,
		ProposalInterval:       cfg.ProposalInterval,
		L2OutputOracleAddr:     cfg.OutputOracleAddr,
		DisputeGameFactoryAddr: cfg.DisputeGameFactoryAddr,
		DisputeGameType:        cfg.DisputeGameType,
		AllowNonFinalized:      cfg.AllowNonFinalized,
	}
	rollupProvider, err := dial.NewStaticL2RollupProviderFromExistingRollup(rollupCl)
	require.NoError(t, err)

	driverSetup := proposer.DriverSetup{
		Log:            log,
		Metr:           metrics.NoopMetrics,
		Cfg:            proposerConfig,
		Txmgr:          fakeTxMgr{from: crypto.PubkeyToAddress(cfg.ProposerKey.PublicKey), chainID: cfg.ChainID},
		L1Client:       l1,
		Multicaller:    batching.NewMultiCaller(l1.Client(), batching.DefaultBatchSize),
		ProposalSource: source.NewRollupProposalSource(rollupProvider),
	}

	dr, err := proposer.NewL2OutputSubmitter(driverSetup)
	require.NoError(t, err)

	address := crypto.PubkeyToAddress(cfg.ProposerKey.PublicKey)

	var l2OutputOracle *mantlebindings.L2OutputOracleCaller
	var disputeGameFactory *bindings.DisputeGameFactoryCaller

	l2OutputOracle, err = mantlebindings.NewL2OutputOracleCaller(*cfg.OutputOracleAddr, l1)
	require.NoError(t, err)
	proposer, err := l2OutputOracle.PROPOSER(&bind.CallOpts{})
	require.NoError(t, err)
	require.Equal(t, proposer, address, "PROPOSER must be the proposer's address")

	return &MantleL2Proposer{
		log:                    log,
		l1:                     l1,
		driver:                 dr,
		l2OutputOracle:         l2OutputOracle,
		l2OutputOracleAddr:     cfg.OutputOracleAddr,
		disputeGameFactory:     disputeGameFactory,
		disputeGameFactoryAddr: cfg.DisputeGameFactoryAddr,
		address:                address,
		privKey:                cfg.ProposerKey,
		allocType:              cfg.AllocType,
	}
}

func (p *MantleL2Proposer) fetchMantleNextOutput(t Testing) (source.Proposal, bool, error) {

	return p.driver.FetchL2OOOutput(t.Ctx())
}

func (p *MantleL2Proposer) CanMantlePropose(t Testing) bool {
	_, shouldPropose, err := p.fetchMantleNextOutput(t)
	require.NoError(t, err)
	return shouldPropose
}

func (p *MantleL2Proposer) ActMantleMakeProposalTx(t Testing) {
	output, shouldPropose, err := p.fetchMantleNextOutput(t)
	require.NoError(t, err)

	if !shouldPropose {
		return
	}

	var txData []byte

	txData, err = p.driver.ProposeL2OutputTxData(output)
	require.NoError(t, err)

	// Note: Use L1 instead of the output submitter's transaction manager because
	// this is non-blocking while the txmgr is blocking & deadlocks the tests
	p.sendMantleTx(t, txData)
}

func (p *MantleL2Proposer) LastMantleProposalTx() common.Hash {
	return p.lastTx
}

// sendTx reimplements creating & sending transactions because we need to do the final send as async in
// the action tests while we do it synchronously in the real system.
func (p *MantleL2Proposer) sendMantleTx(t Testing, data []byte) {
	gasTipCap := big.NewInt(2 * params.GWei)
	pendingHeader, err := p.l1.HeaderByNumber(t.Ctx(), big.NewInt(-1))
	require.NoError(t, err, "need l1 pending header for gas price estimation")
	gasFeeCap := new(big.Int).Add(gasTipCap, new(big.Int).Mul(pendingHeader.BaseFee, big.NewInt(2)))
	chainID, err := p.l1.ChainID(t.Ctx())
	require.NoError(t, err)
	nonce, err := p.l1.NonceAt(t.Ctx(), p.address, nil)
	require.NoError(t, err)

	var addr common.Address

	addr = *p.l2OutputOracleAddr

	gasLimit, err := estimateGasPending(t.Ctx(), p.l1, ethereum.CallMsg{
		From:      p.address,
		To:        &addr,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Data:      data,
	})
	require.NoError(t, err)

	rawTx := &types.DynamicFeeTx{
		Nonce:     nonce,
		To:        &addr,
		Data:      data,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasTipCap,
		Gas:       gasLimit,
		ChainID:   chainID,
	}

	tx, err := types.SignNewTx(p.privKey, types.LatestSignerForChainID(chainID), rawTx)
	require.NoError(t, err, "need to sign tx")

	err = p.l1.SendTransaction(t.Ctx(), tx)
	require.NoError(t, err, "need to send tx")

	p.lastTx = tx.Hash()
}
