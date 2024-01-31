package genesis

import (
	"context"
	"github.com/ethereum-optimism/optimism/l2geth/common"
	"github.com/ethereum-optimism/optimism/l2geth/common/hexutil"
	"github.com/ethereum-optimism/optimism/l2geth/rpc"
	"github.com/ethereum-optimism/optimism/op-chain-ops/state"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/core"
)

// BuildL2DeveloperGenesis will build the developer Optimism Genesis
// Block. Suitable for devnets.
func BuildL2DeveloperGenesis(config *DeployConfig, l1StartBlock *types.Block, rpcBlock *RpcBlock) (*core.Genesis, error) {
	genspec, err := NewL2Genesis(config, l1StartBlock)
	if err != nil {
		return nil, err
	}

	db := state.NewMemoryStateDB(genspec)

	if config.FundDevAccounts {
		FundDevAccounts(db)
	}
	SetPrecompileBalances(db)

	storage, err := NewL2StorageConfig(config, l1StartBlock, rpcBlock)
	if err != nil {
		return nil, err
	}

	immutable, err := NewL2ImmutableConfig(config)
	if err != nil {
		return nil, err
	}

	if err := SetL2Proxies(db); err != nil {
		return nil, err
	}

	if err := SetImplementations(db, storage, immutable); err != nil {
		return nil, err
	}

	if err := SetDevOnlyL2Implementations(db, storage, immutable); err != nil {
		return nil, err
	}

	return db.Genesis(), nil
}

func HeaderCall(ctx context.Context, rawurl string, method string, id int64) (*RpcHeader, error) {
	var header *RpcHeader
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	err = c.CallContext(ctx, &header, method, id, false) // headers are just blocks without txs
	if err != nil {
		return nil, err
	}
	if header == nil {
		return nil, errors.Errorf("Not Found")
	}

	return header, nil
}

func BlockCall(ctx context.Context, rawurl string, method string, id int64) (*RpcBlock, error) {
	var block *RpcBlock
	c, err := rpc.DialContext(ctx, rawurl)
	if err != nil {
		return nil, err
	}
	err = c.CallContext(ctx, &block, method, id, true)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, errors.Errorf("Not Found")
	}

	return block, nil
}

type RpcHeader struct {
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

	// untrusted info included by RPC, may have to be checked
	Hash common.Hash `json:"hash"`
}

type RpcBlock struct {
	RpcHeader
	Transactions []*types.Transaction `json:"transactions"`
}
