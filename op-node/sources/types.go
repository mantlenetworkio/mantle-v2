package sources

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/holiman/uint256"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	opseth "github.com/ethereum-optimism/optimism/op-service/eth"
)

type BatchCallContextFn func(ctx context.Context, b []rpc.BatchElem) error

type CallContextFn func(ctx context.Context, result any, method string, args ...any) error

// Note: these types are used, instead of the geth types, to enable:
// - batched calls of many block requests (standard bindings do extra uncle-header fetches, cannot be batched nicely)
// - ignore uncle data (does not even exist anymore post-Merge)
// - use cached block hash, if we trust the RPC.
// - verify transactions list matches tx-root, to ensure consistency with block-hash, if we do not trust the RPC
// - verify block contents are compatible with Post-Merge ExecutionPayload format
//
// Transaction-sender data from the RPC is not cached, since ethclient.setSenderFromServer is private,
// and we only need to compute the sender for transactions into the inbox.
//
// This way we minimize RPC calls, enable batching, and can choose to verify what the RPC gives us.

// headerInfo is a conversion type of types.Header turning it into a
// BlockInfo, but using a cached hash value.
type headerInfo struct {
	hash common.Hash
	*types.Header
}

var _ eth.BlockInfo = (*headerInfo)(nil)

func (h headerInfo) Hash() common.Hash {
	return h.hash
}

func (h headerInfo) ParentHash() common.Hash {
	return h.Header.ParentHash
}

func (h headerInfo) Coinbase() common.Address {
	return h.Header.Coinbase
}

func (h headerInfo) Root() common.Hash {
	return h.Header.Root
}

func (h headerInfo) NumberU64() uint64 {
	return h.Header.Number.Uint64()
}

func (h headerInfo) Time() uint64 {
	return h.Header.Time
}

func (h headerInfo) MixDigest() common.Hash {
	return h.Header.MixDigest
}

func (h headerInfo) BaseFee() *big.Int {
	return h.Header.BaseFee
}

func (h headerInfo) BlobBaseFee() *big.Int {
	if h.Header.ExcessBlobGas == nil {
		return nil
	}
	return opseth.CalcBlobFeeDefault(h.Header)
}

func (h headerInfo) ExcessBlobGas() *uint64 {
	return h.Header.ExcessBlobGas
}

func (h headerInfo) ReceiptHash() common.Hash {
	return h.Header.ReceiptHash
}

func (h headerInfo) GasUsed() uint64 {
	return h.Header.GasUsed
}

func (h headerInfo) GasLimit() uint64 {
	return h.Header.GasLimit
}

func (h headerInfo) ParentBeaconRoot() *common.Hash {
	return h.Header.ParentBeaconRoot
}

func (h headerInfo) HeaderRLP() ([]byte, error) {
	return rlp.EncodeToBytes(h.Header)
}

func (h headerInfo) WithdrawalsRoot() *common.Hash {
	return h.Header.WithdrawalsHash
}

func (h headerInfo) MarshalJSON() ([]byte, error) {
	return h.Header.MarshalJSON()
}

func (h headerInfo) UnmarshalJSON(input []byte) error {
	return h.Header.UnmarshalJSON(input)
}

type rpcHeader struct {
	ParentHash  common.Hash      `json:"parentHash"`
	UncleHash   common.Hash      `json:"sha3Uncles"`
	Coinbase    common.Address   `json:"miner"`
	Root        common.Hash      `json:"stateRoot"`
	TxHash      common.Hash      `json:"transactionsRoot"`
	ReceiptHash common.Hash      `json:"receiptsRoot"`
	Bloom       eth.Bytes256     `json:"logsBloom"`
	Difficulty  hexutil.Big      `json:"difficulty"`
	Number      hexutil.Uint64   `json:"number"`
	GasLimit    hexutil.Uint64   `json:"gasLimit"`
	GasUsed     hexutil.Uint64   `json:"gasUsed"`
	Time        hexutil.Uint64   `json:"timestamp"`
	Extra       hexutil.Bytes    `json:"extraData"`
	MixDigest   common.Hash      `json:"mixHash"`
	Nonce       types.BlockNonce `json:"nonce"`

	// BaseFee was added by EIP-1559 and is ignored in legacy headers.
	BaseFee *hexutil.Big `json:"baseFeePerGas"`

	// WithdrawalsRoot was added by EIP-4895 and is ignored in legacy headers.
	WithdrawalsRoot *common.Hash `json:"withdrawalsRoot"`

	// BlobGasUsed was added by EIP-4844 and is ignored in legacy headers.
	BlobGasUsed *hexutil.Uint64 `json:"blobGasUsed" rlp:"optional"`

	// ExcessBlobGas was added by EIP-4844 and is ignored in legacy headers.
	ExcessBlobGas *hexutil.Uint64 `json:"excessBlobGas" rlp:"optional"`

	// ParentBeaconRoot was added by EIP-4788 and is ignored in legacy headers.
	ParentBeaconRoot *common.Hash `json:"parentBeaconBlockRoot" rlp:"optional"`

	// RequestsHash was added by EIP-7685 and is ignored in legacy headers.
	RequestsHash *common.Hash `json:"requestsHash" rlp:"optional"`

	// untrusted info included by RPC, may have to be checked
	Hash common.Hash `json:"hash"`
}

// checkPostMerge checks that the block header meets all criteria to be a valid ExecutionPayloadHeader,
// see EIP-3675 (block header changes) and EIP-4399 (mixHash usage for prev-randao)
func (hdr *rpcHeader) checkPostMerge() error {
	// TODO: the genesis block has a non-zero difficulty number value.
	// Either this block needs to change, or we special case it. This is not valid w.r.t. EIP-3675.
	if hdr.Number != 0 && (*big.Int)(&hdr.Difficulty).Cmp(common.Big0) != 0 {
		return fmt.Errorf("post-merge block header requires zeroed difficulty field, but got: %s", &hdr.Difficulty)
	}
	if hdr.Nonce != (types.BlockNonce{}) {
		return fmt.Errorf("post-merge block header requires zeroed block nonce field, but got: %s", hdr.Nonce)
	}
	if hdr.BaseFee == nil {
		return fmt.Errorf("post-merge block header requires EIP-1559 basefee field, but got %s", hdr.BaseFee)
	}
	if len(hdr.Extra) > 32 {
		return fmt.Errorf("post-merge block header requires 32 or less bytes of extra data, but got %d", len(hdr.Extra))
	}
	if hdr.UncleHash != types.EmptyUncleHash {
		return fmt.Errorf("post-merge block header requires uncle hash to be of empty uncle list, but got %s", hdr.UncleHash)
	}
	return nil
}

func (hdr *rpcHeader) computeBlockHash() common.Hash {
	gethHeader := hdr.createGethHeader()
	return gethHeader.Hash()
}

func (hdr *rpcHeader) createGethHeader() *types.Header {
	var blobGasUsed *uint64
	if hdr.BlobGasUsed != nil {
		blobGasUsedTemp := uint64(*hdr.BlobGasUsed)
		blobGasUsed = &blobGasUsedTemp
	}
	var excessBlobGas *uint64
	if hdr.BlobGasUsed != nil {
		excessBlobGasTemp := uint64(*hdr.ExcessBlobGas)
		excessBlobGas = &excessBlobGasTemp
	}

	return &types.Header{
		ParentHash:      hdr.ParentHash,
		UncleHash:       hdr.UncleHash,
		Coinbase:        hdr.Coinbase,
		Root:            hdr.Root,
		TxHash:          hdr.TxHash,
		ReceiptHash:     hdr.ReceiptHash,
		Bloom:           types.Bloom(hdr.Bloom),
		Difficulty:      (*big.Int)(&hdr.Difficulty),
		Number:          new(big.Int).SetUint64(uint64(hdr.Number)),
		GasLimit:        uint64(hdr.GasLimit),
		GasUsed:         uint64(hdr.GasUsed),
		Time:            uint64(hdr.Time),
		Extra:           hdr.Extra,
		MixDigest:       hdr.MixDigest,
		Nonce:           hdr.Nonce,
		BaseFee:         (*big.Int)(hdr.BaseFee),
		WithdrawalsHash: hdr.WithdrawalsRoot,

		BlobGasUsed:      blobGasUsed,
		ExcessBlobGas:    excessBlobGas,
		ParentBeaconRoot: hdr.ParentBeaconRoot,
		RequestsHash:     hdr.RequestsHash,
	}
}

func (hdr *rpcHeader) Info(trustCache bool, mustBePostMerge bool) (eth.BlockInfo, error) {
	if mustBePostMerge {
		if err := hdr.checkPostMerge(); err != nil {
			return nil, err
		}
	}
	if !trustCache {
		if computed := hdr.computeBlockHash(); computed != hdr.Hash {
			return nil, fmt.Errorf("failed to verify block hash: computed %s but RPC said %s", computed, hdr.Hash)
		}
	}
	return &headerInfo{hdr.Hash, hdr.createGethHeader()}, nil
}

type rpcBlock struct {
	rpcHeader
	Transactions []*types.Transaction `json:"transactions"`
	Withdrawals  *types.Withdrawals   `json:"withdrawals,omitempty"`
}

func (block *rpcBlock) verify() error {
	if computed := block.computeBlockHash(); computed != block.Hash {
		return fmt.Errorf("failed to verify block hash: computed %s but RPC said %s", computed, block.Hash)
	}
	for i, tx := range block.Transactions {
		if tx == nil {
			return fmt.Errorf("block tx %d is nil", i)
		}
	}
	if computed := types.DeriveSha(types.Transactions(block.Transactions), trie.NewStackTrie(nil)); block.TxHash != computed {
		return fmt.Errorf("failed to verify transactions list: computed %s but RPC said %s", computed, block.TxHash)
	}

	// Withdrawals validation is different between L1 and L2.
	// It is possible to determine that it is an L2 block if the first transaction is a deposit.
	// The genesis block does not have transactions, but does have a known fee-recipient predeploy address.
	isL2 := (len(block.Transactions) > 0 && block.Transactions[0].IsDepositTx()) ||
		(block.Number == 0 && block.Coinbase == predeploys.SequencerFeeVaultAddr)
	if isL2 {
		if err := block.validateL2Withdrawals(block.Withdrawals, block.WithdrawalsRoot); err != nil {
			return err
		}
	} else {
		if err := block.validateL1Withdrawals(block.Withdrawals, block.WithdrawalsRoot); err != nil {
			return err
		}
	}
	return nil
}

func (block *rpcBlock) validateL1Withdrawals(withdrawals *types.Withdrawals, withdrawalsRoot *common.Hash) error {
	if withdrawalsRoot != nil {
		if withdrawals == nil {
			return errors.New("expected withdrawals")
		}
		for i, w := range *withdrawals {
			if w == nil {
				return fmt.Errorf("block withdrawal %d is null", i)
			}
		}
		if computed := types.DeriveSha(*withdrawals, trie.NewStackTrie(nil)); *withdrawalsRoot != computed {
			return fmt.Errorf("failed to verify withdrawals list: computed %s but RPC said %s", computed, withdrawalsRoot)
		}
	} else {
		if withdrawals != nil {
			return fmt.Errorf("expected no withdrawals due to missing withdrawals-root, but got %d", len(*withdrawals))
		}
	}
	return nil
}

func (block *rpcBlock) validateL2Withdrawals(withdrawals *types.Withdrawals, withdrawalsRoot *common.Hash) error {
	if withdrawalsRoot != nil {
		if !(withdrawals != nil && len(*withdrawals) == 0) {
			return fmt.Errorf("expected empty withdrawals, but got %d", len(*withdrawals))
		}
	}
	return nil
}

func (block *rpcBlock) Info(trustCache bool, mustBePostMerge bool) (eth.BlockInfo, types.Transactions, error) {
	if mustBePostMerge {
		if err := block.checkPostMerge(); err != nil {
			return nil, nil, err
		}
	}
	if !trustCache {
		if err := block.verify(); err != nil {
			return nil, nil, err
		}
	}

	// verify the header data
	info, err := block.rpcHeader.Info(trustCache, mustBePostMerge)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to verify block from RPC: %w", err)
	}

	return info, block.Transactions, nil
}

func (block *rpcBlock) ExecutionPayloadEnvelope(trustCache bool) (*eth.ExecutionPayloadEnvelope, error) {
	if err := block.checkPostMerge(); err != nil {
		return nil, err
	}
	if !trustCache {
		if err := block.verify(); err != nil {
			return nil, err
		}
	}
	var baseFee uint256.Int
	baseFee.SetFromBig((*big.Int)(block.BaseFee))

	// Unfortunately eth_getBlockByNumber either returns full transactions, or only tx-hashes.
	// There is no option for encoded transactions.
	opaqueTxs := make([]hexutil.Bytes, len(block.Transactions))
	for i, tx := range block.Transactions {
		data, err := tx.MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("failed to encode tx %d from RPC: %w", i, err)
		}
		opaqueTxs[i] = data
	}

	payload := &eth.ExecutionPayload{
		ParentHash:      block.ParentHash,
		FeeRecipient:    block.Coinbase,
		StateRoot:       eth.Bytes32(block.Root),
		ReceiptsRoot:    eth.Bytes32(block.ReceiptHash),
		LogsBloom:       block.Bloom,
		PrevRandao:      eth.Bytes32(block.MixDigest), // mix-digest field is used for prevRandao post-merge
		BlockNumber:     block.Number,
		GasLimit:        block.GasLimit,
		GasUsed:         block.GasUsed,
		Timestamp:       block.Time,
		ExtraData:       eth.BytesMax32(block.Extra),
		BaseFeePerGas:   eth.Uint256Quantity(baseFee),
		BlockHash:       block.Hash,
		Transactions:    opaqueTxs,
		Withdrawals:     block.Withdrawals,
		BlobGasUsed:     block.BlobGasUsed,
		ExcessBlobGas:   block.ExcessBlobGas,
		WithdrawalsRoot: block.WithdrawalsRoot,
	}

	return &eth.ExecutionPayloadEnvelope{
		ParentBeaconBlockRoot: block.ParentBeaconRoot,
		ExecutionPayload:      payload,
	}, nil
}

// blockHashParameter is used as "block parameter":
// Some Nethermind and Alchemy RPC endpoints require an object to identify a block, instead of a string.
type blockHashParameter struct {
	BlockHash common.Hash `json:"blockHash"`
}

// unusableMethod identifies if an error indicates that the RPC method cannot be used as expected:
// if it's an unknown method, or if parameters were invalid.
func unusableMethod(err error) bool {
	if rpcErr, ok := err.(rpc.Error); ok {
		code := rpcErr.ErrorCode()
		// method not found, or invalid params
		if code == -32601 || code == -32602 {
			return true
		}
	} else {
		errText := strings.ToLower(err.Error())
		if strings.Contains(errText, "unknown method") || strings.Contains(errText, "invalid param") || strings.Contains(errText, "is not available") {
			return true
		}
	}
	return false
}
