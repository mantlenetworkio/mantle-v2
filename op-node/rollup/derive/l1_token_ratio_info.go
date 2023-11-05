package derive

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-service/solabi"
)

const (
	TokenRatioInfoFuncSignature = "setTokenRatio(uint256)"
	TokenRatioInfoArguments     = 1
	TokenRatioInfoLen           = 4 + 32*TokenRatioInfoArguments
)

var (
	TokenRatioInfoFuncBytes4    = crypto.Keccak256([]byte(TokenRatioInfoFuncSignature))[:4]
	TokenRatioInfoSenderAddress = common.HexToAddress("0xdeaddeaddeaddeaddeaddeaddeaddeaddead0001")
	TokenRatioInfoAddress       = predeploys.L1BlockAddr
)

// TokenRatioInfo presents the information stored in a L1Block.setL1BlockValues call
type TokenRatioInfo struct {
	TokenRatio *big.Int
}

// Binary Format
// +---------+--------------------------+
// | Bytes   | Field                    |
// +---------+--------------------------+
// | 4       | Function signature       |
// | 32      | Token ratio                   |
// +---------+--------------------------+

func (info *TokenRatioInfo) MarshalBinary() ([]byte, error) {
	w := bytes.NewBuffer(make([]byte, 0, TokenRatioInfoArguments))
	if err := solabi.WriteSignature(w, TokenRatioInfoFuncBytes4); err != nil {
		return nil, err
	}
	if err := solabi.WriteUint256(w, info.TokenRatio); err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

func (info *TokenRatioInfo) UnmarshalBinary(data []byte) error {
	if len(data) != TokenRatioInfoLen {
		return fmt.Errorf("data is unexpected length: %d", len(data))
	}
	reader := bytes.NewReader(data)

	var err error
	if _, err := solabi.ReadAndValidateSignature(reader, L1InfoFuncBytes4); err != nil {
		return err
	}
	if info.TokenRatio, err = solabi.ReadUint256(reader); err != nil {
		return err
	}
	if !solabi.EmptyReader(reader) {
		return errors.New("too many bytes")
	}
	return nil
}

// TokenRatioTxData is the inverse of tokenRatioInfo, to see where the L2 chain is derived from
func TokenRatioTxData(data []byte) (TokenRatioInfo, error) {
	var info TokenRatioInfo
	err := info.UnmarshalBinary(data)
	return info, err
}

// TokenRatioInfoTx creates a Token Ratio Info deposit transaction based on the ETH/MNT ratio,
// and the L2 block-height difference with the start of the epoch.
func TokenRatioInfoTx(seqNumber uint64, block eth.BlockInfo, regolith bool, tokenRatio *big.Int) (*types.DepositTx, error) {
	infoDat := TokenRatioInfo{
		TokenRatio: tokenRatio,
	}
	data, err := infoDat.MarshalBinary()
	if err != nil {
		return nil, err
	}

	source := L1InfoDepositSource{
		L1BlockHash: block.Hash(),
		SeqNumber:   seqNumber,
	}
	// Set a very large gas limit with `IsSystemTransaction` to ensure
	// that the L1 Attributes Transaction does not run out of gas.
	out := &types.DepositTx{
		SourceHash:             source.SourceHash(),
		From:                   L1InfoDepositerAddress,
		To:                     &L1BlockAddress,
		Mint:                   nil,
		EthValue:               nil,
		Value:                  big.NewInt(0),
		Gas:                    150_000_000,
		IsSystemTransaction:    true,
		IsBroadcastTransaction: true,
		Data:                   data,
	}
	// With the regolith fork we disable the IsSystemTx functionality, and allocate real gas
	if regolith {
		out.IsSystemTransaction = false
		out.Gas = RegolithSystemTxGas
	}
	return out, nil
}

// TokenRatioBytes returns a serialized token_ratio attributes transaction.
func TokenRatioBytes(seqNumber uint64, l1Info eth.BlockInfo, regolith bool, tokenRatio *big.Int) ([]byte, error) {
	dep, err := TokenRatioInfoTx(seqNumber, l1Info, regolith, tokenRatio)
	if err != nil {
		return nil, fmt.Errorf("failed to create L1 info tx: %w", err)
	}
	l1Tx := types.NewTx(dep)
	opaqueL1Tx, err := l1Tx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode L1 info tx: %w", err)
	}
	return opaqueL1Tx, nil
}
