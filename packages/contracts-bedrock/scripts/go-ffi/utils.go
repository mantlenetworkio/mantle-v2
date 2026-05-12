package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum-optimism/optimism/op-chain-ops/crossdomain"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum-optimism/optimism/op-service/bigs"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

type OutputRootWithChainId struct {
	ChainId *big.Int
	Root    common.Hash
}

// Define a proper type for SuperRootProof
type SuperRootProof struct {
	Version     uint8
	Timestamp   uint64
	OutputRoots []OutputRootWithChainId
}

var UnknownNonceVersion = errors.New("Unknown nonce version")

// Mantle-specific V1 relay message ABI with dual values (mntValue, ethValue)
var mantleRelayMessage1ABI = `[{"inputs":[{"internalType":"uint256","name":"_nonce","type":"uint256"},{"internalType":"address","name":"_sender","type":"address"},{"internalType":"address","name":"_target","type":"address"},{"internalType":"uint256","name":"_mntValue","type":"uint256"},{"internalType":"uint256","name":"_ethValue","type":"uint256"},{"internalType":"uint256","name":"_minGasLimit","type":"uint256"},{"internalType":"bytes","name":"_message","type":"bytes"}],"name":"relayMessage","outputs":[],"stateMutability":"payable","type":"function"}]`
var mantleRelayMessage1 abi.ABI

func init() {
	var err error
	mantleRelayMessage1, err = abi.JSON(strings.NewReader(mantleRelayMessage1ABI))
	if err != nil {
		panic(err)
	}
}

// checkOk checks if ok is false, and panics if so.
// Shorthand to ease go's god awful error handling
func checkOk(ok bool) {
	if !ok {
		panic(fmt.Errorf("checkOk failed"))
	}
}

// checkErr checks if err is not nil, and throws if so.
// Shorthand to ease go's god awful error handling
func checkErr(err error, failReason string) {
	if err != nil {
		panic(fmt.Errorf("%s: %w", failReason, err))
	}
}

// encodeCrossDomainMessage encodes a versioned cross domain message into a byte array.
// Mantle uses dual values (mntValue, ethValue) in V1 encoding.
func encodeCrossDomainMessage(nonce *big.Int, sender common.Address, target common.Address, mntValue *big.Int, ethValue *big.Int, gasLimit *big.Int, data []byte) ([]byte, error) {
	_, version := crossdomain.DecodeVersionedNonce(nonce)

	var encoded []byte
	var err error
	if version.Cmp(big.NewInt(0)) == 0 {
		// Encode cross domain message V0
		encoded, err = crossdomain.EncodeCrossDomainMessageV0(target, sender, data, nonce)
	} else if version.Cmp(big.NewInt(1)) == 0 {
		// Encode cross domain message V1 with Mantle dual values
		// relayMessage(uint256,address,address,uint256,uint256,uint256,bytes)
		encoded, err = mantleRelayMessage1.Pack("relayMessage", nonce, sender, target, mntValue, ethValue, gasLimit, data)
	} else {
		return nil, UnknownNonceVersion
	}

	return encoded, err
}

// parseSuperRootProof parses an abi encoded super root proof into a SuperRootProof struct.
func parseSuperRootProof(abiEncodedProof []byte) (*SuperRootProof, error) {
	// Parse the input as hex data
	unpacked, err := superRootProofArgs.Unpack(abiEncodedProof)
	if err != nil {
		return nil, err
	}

	// The Unpack method returns a slice of interface{}, so we need to get the first element
	if len(unpacked) != 1 {
		return nil, errors.New("unexpected number of values after unpacking super root proof")
	}

	// Use an anonymous struct matching the tuple’s layout.
	tmp := unpacked[0].(struct {
		Version     [1]uint8 `json:"version"`
		Timestamp   uint64   `json:"timestamp"`
		OutputRoots []struct {
			ChainId *big.Int `json:"chainId"`
			Root    [32]byte `json:"root"`
		} `json:"outputRoots"`
	})

	// Convert into our desired SuperRootProof type.
	proof := SuperRootProof{
		Version:   tmp.Version[0],
		Timestamp: tmp.Timestamp,
	}
	for _, o := range tmp.OutputRoots {
		proof.OutputRoots = append(proof.OutputRoots, OutputRootWithChainId{
			ChainId: o.ChainId,
			Root:    common.BytesToHash(o.Root[:]),
		})
	}

	return &proof, nil
}

// encodeSuperRootProof encodes a super root proof into a byte array.
func encodeSuperRootProof(superRootProof *SuperRootProof) ([]byte, error) {
	// Version must match the expected version (0x01)
	if superRootProof.Version != 0x01 {
		return nil, errors.New("invalid super root version")
	}

	// Output roots must not be empty
	if len(superRootProof.OutputRoots) == 0 {
		return nil, errors.New("empty super root")
	}

	// Start with version byte and timestamp
	encoded := []byte{superRootProof.Version}

	// Add timestamp as bytes8 (uint64)
	timestampBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBytes, superRootProof.Timestamp)
	encoded = append(encoded, timestampBytes...)

	// Add each output root (chainId + root)
	for _, outputRoot := range superRootProof.OutputRoots {
		// Append chainId bytes (padded to 32 bytes)
		chainIdBytes := make([]byte, 32)
		outputRoot.ChainId.FillBytes(chainIdBytes)
		encoded = append(encoded, chainIdBytes...)

		// Append root hash (already 32 bytes)
		encoded = append(encoded, outputRoot.Root.Bytes()...)
	}

	return encoded, nil
}

// hashWithdrawal hashes a withdrawal transaction with Mantle dual values.
// Matches Solidity: keccak256(abi.encode(nonce, sender, target, mntValue, ethValue, gasLimit, data))
func hashWithdrawal(nonce *big.Int, sender common.Address, target common.Address, mntValue *big.Int, ethValue *big.Int, gasLimit *big.Int, data []byte) (common.Hash, error) {
	wdArgs := abi.Arguments{
		{Name: "nonce", Type: uint256ABI},
		{Name: "sender", Type: addressABI},
		{Name: "target", Type: addressABI},
		{Name: "mntValue", Type: uint256ABI},
		{Name: "ethValue", Type: uint256ABI},
		{Name: "gasLimit", Type: uint256ABI},
		{Name: "data", Type: bytesABI},
	}
	enc, err := wdArgs.Pack(nonce, sender, target, mntValue, ethValue, gasLimit, data)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(enc), nil
}

var (
	uint256ABI, _ = abi.NewType("uint256", "", nil)
	addressABI, _ = abi.NewType("address", "", nil)
	bytesABI, _   = abi.NewType("bytes", "", nil)
)

// hashOutputRootProof hashes an output root proof.
// Matches Solidity: keccak256(abi.encode(version, stateRoot, messagePasserStorageRoot, latestBlockhash))
func hashOutputRootProof(version common.Hash, stateRoot common.Hash, messagePasserStorageRoot common.Hash, latestBlockHash common.Hash) (common.Hash, error) {
	hashArgs := abi.Arguments{
		{Name: "version", Type: fixedBytes},
		{Name: "stateRoot", Type: fixedBytes},
		{Name: "messagePasserStorageRoot", Type: fixedBytes},
		{Name: "latestBlockhash", Type: fixedBytes},
	}
	enc, err := hashArgs.Pack(version, stateRoot, messagePasserStorageRoot, latestBlockHash)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(enc), nil
}

// makeDepositTx creates a deposit transaction type with Mantle dual values.
func makeDepositTx(
	from common.Address,
	to common.Address,
	mntValue *big.Int,
	mntTxValue *big.Int,
	ethValue *big.Int,
	ethTxValue *big.Int,
	gasLimit *big.Int,
	isCreate bool,
	data []byte,
	l1BlockHash common.Hash,
	logIndex *big.Int,
) types.DepositTx {
	// Create deposit transaction source
	udp := derive.UserDepositSource{
		L1BlockHash: l1BlockHash,
		LogIndex:    bigs.Uint64Strict(logIndex),
	}

	// Create deposit transaction
	depositTx := types.DepositTx{
		SourceHash:          udp.SourceHash(),
		From:                from,
		Value:               mntTxValue,
		Gas:                 bigs.Uint64Strict(gasLimit),
		IsSystemTransaction: false, // This will never be a system transaction in the tests.
		Data:                data,
	}

	// Fill optional fields - Mint is mntValue (L2 MNT mint)
	if mntValue.Cmp(big.NewInt(0)) == 1 {
		depositTx.Mint = mntValue
	}
	// EthValue is L2 BVM_ETH mint
	if ethValue.Cmp(big.NewInt(0)) == 1 {
		depositTx.EthValue = ethValue
	}
	// EthTxValue is L2 BVM_ETH transfer
	if ethTxValue.Cmp(big.NewInt(0)) >= 0 {
		depositTx.EthTxValue = ethTxValue
	}
	if !isCreate {
		depositTx.To = &to
	}

	return depositTx
}

// Custom type to write the generated proof to
type proofList [][]byte

func (n *proofList) Put(key []byte, value []byte) error {
	*n = append(*n, value)
	return nil
}

func (n *proofList) Delete(key []byte) error {
	panic("not supported")
}
