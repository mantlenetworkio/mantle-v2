package engineapi

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/consensus/misc/eip1559"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

var (
	ErrExceedsGasLimit  = errors.New("tx gas exceeds block gas limit")
	ErrUsesTooMuchGas   = errors.New("action takes too much gas")
	errInvalidGasLimit  = errors.New("invalid gas limit")
	errInvalidTimestamp = errors.New("invalid timestamp")
)

type BlockDataProvider interface {
	StateAt(root common.Hash) (*state.StateDB, error)
	GetHeader(common.Hash, uint64) *types.Header
	Engine() consensus.Engine
	GetVMConfig() *vm.Config
	Config() *params.ChainConfig
	consensus.ChainHeaderReader
}

type BlockProcessor struct {
	header       *types.Header
	state        *state.StateDB
	receipts     types.Receipts
	transactions types.Transactions
	gasPool      *core.GasPool
	dataProvider BlockDataProvider
	evm          *vm.EVM
}

func NewBlockProcessorFromPayloadAttributes(provider BlockDataProvider, parent common.Hash, attrs *eth.PayloadAttributes) (*BlockProcessor, error) {
	header := &types.Header{
		ParentHash:       parent,
		Coinbase:         attrs.SuggestedFeeRecipient,
		Difficulty:       common.Big0,
		GasLimit:         uint64(*attrs.GasLimit),
		Time:             uint64(attrs.Timestamp),
		Extra:            nil,
		MixDigest:        common.Hash(attrs.PrevRandao),
		Nonce:            types.EncodeNonce(0),
		ParentBeaconRoot: attrs.ParentBeaconBlockRoot,
	}

	return NewBlockProcessorFromHeader(provider, header)
}

func NewBlockProcessorFromHeader(provider BlockDataProvider, h *types.Header) (*BlockProcessor, error) {
	header := types.CopyHeader(h) // Copy to avoid mutating the original header

	if header.GasLimit > params.MaxGasLimit {
		return nil, fmt.Errorf("%w: have %v, max %v", errInvalidGasLimit, header.GasLimit, params.MaxGasLimit)
	}
	parentHeader := provider.GetHeaderByHash(header.ParentHash)
	if header.Time <= parentHeader.Time {
		return nil, errInvalidTimestamp
	}
	statedb, err := provider.StateAt(parentHeader.Root)
	if err != nil {
		return nil, fmt.Errorf("get parent state: %w", err)
	}
	header.Number = new(big.Int).Add(parentHeader.Number, common.Big1)
	header.BaseFee = eip1559.CalcBaseFee(provider.Config(), parentHeader)
	header.GasUsed = 0
	gasPool := new(core.GasPool).AddGas(header.GasLimit)
	mkEVM := func() *vm.EVM {
		// Unfortunately this is not part of any Geth environment setup,
		// we just have to apply it, like how the Geth block-builder worker does.
		context := core.NewEVMBlockContext(header, provider, nil, provider.Config(), statedb)
		vmenv := vm.NewEVM(context, statedb, provider.Config(), vm.Config{})
		return vmenv
	}
	var vmenv *vm.EVM
	if h.ParentBeaconRoot != nil {
		if provider.Config().IsCancun(header.Number, header.Time) {
			// Blob tx not supported on optimism chains but fields must be set when Cancun is active.
			zero := uint64(0)
			header.BlobGasUsed = &zero
			header.ExcessBlobGas = &zero
		}
		// core.NewEVMBlockContext need to be called after the blob gas fields are set
		vmenv = mkEVM()
		core.ProcessBeaconBlockRoot(*header.ParentBeaconRoot, vmenv)
	} else {
		vmenv = mkEVM()
	}
	if provider.Config().IsPrague(header.Number, header.Time) {
		core.ProcessParentBlockHash(header.ParentHash, vmenv)
	}
	if provider.Config().IsMantleSkadi(header.Time) {
		// set the header withdrawals root for Isthmus blocks
		mpHash := statedb.GetStorageRoot(predeploys.L2ToL1MessagePasserAddr)
		header.WithdrawalsHash = &mpHash

		// set the header requests root to empty hash for Isthmus blocks
		header.RequestsHash = &types.EmptyRequestsHash
	}

	return &BlockProcessor{
		header:       header,
		state:        statedb,
		gasPool:      gasPool,
		dataProvider: provider,
		evm:          vmenv,
	}, nil
}

func (b *BlockProcessor) CheckTxWithinGasLimit(tx *types.Transaction) error {
	if tx.Gas() > b.header.GasLimit {
		return fmt.Errorf("%w tx gas: %d, block gas limit: %d", ErrExceedsGasLimit, tx.Gas(), b.header.GasLimit)
	}
	if tx.Gas() > b.gasPool.Gas() {
		return fmt.Errorf("%w: %d, only have %d", ErrUsesTooMuchGas, tx.Gas(), b.gasPool.Gas())
	}
	return nil
}

func (b *BlockProcessor) AddTx(tx *types.Transaction) (*types.Receipt, error) {
	txIndex := len(b.transactions)
	b.state.SetTxContext(tx.Hash(), txIndex)
	receipt, err := core.ApplyTransaction(b.evm, b.gasPool, b.state, b.header, tx, &b.header.GasUsed)
	if err != nil {
		return nil, fmt.Errorf("failed to apply transaction to L2 block (tx %d): %w", txIndex, err)
	}
	b.receipts = append(b.receipts, receipt)
	b.transactions = append(b.transactions, tx)
	return receipt, nil
}

func (b *BlockProcessor) Assemble() (*types.Block, types.Receipts, error) {
	body := types.Body{
		Transactions: b.transactions,
	}

	// Processing for EIP-7685 requests would happen here, but is skipped on OP.
	// Kept here to minimize diff.
	if b.dataProvider.Config().IsPrague(b.header.Number, b.header.Time) && !b.dataProvider.Config().IsMantleSkadi(b.header.Time) {
		_requests := [][]byte{}
		// EIP-6110 - no-op because we just ignore all deposit requests, so no need to parse logs
		// EIP-7002
		core.ProcessWithdrawalQueue(&_requests, b.evm)
		// EIP-7251
		core.ProcessConsolidationQueue(&_requests, b.evm)
	}

	block, err := b.dataProvider.Engine().FinalizeAndAssemble(b.dataProvider, b.header, b.state, &body, b.receipts)
	if err != nil {
		return nil, nil, err
	}
	return block, b.receipts, nil
}

func (b *BlockProcessor) Commit() error {
	isCancun := b.dataProvider.Config().IsCancun(b.header.Number, b.header.Time)
	root, err := b.state.Commit(b.header.Number.Uint64(), b.dataProvider.Config().IsEIP158(b.header.Number), isCancun)
	if err != nil {
		return fmt.Errorf("state write error: %w", err)
	}
	if err := b.state.Database().TrieDB().Commit(root, false); err != nil {
		return fmt.Errorf("trie write error: %w", err)
	}
	return nil
}
