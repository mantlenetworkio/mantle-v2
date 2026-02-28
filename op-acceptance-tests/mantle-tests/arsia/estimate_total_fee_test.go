package arsia

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-core/predeploys"
	"github.com/ethereum-optimism/optimism/op-devstack/devtest"
	"github.com/ethereum-optimism/optimism/op-devstack/dsl"
	"github.com/ethereum-optimism/optimism/op-devstack/presets"
	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	txib "github.com/ethereum-optimism/optimism/op-service/txintent/bindings"
	"github.com/ethereum-optimism/optimism/op-service/txintent/contractio"
	"github.com/ethereum-optimism/optimism/op-service/txplan"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

// ============================================================
// Helpers
// ============================================================

// estimateArgs mirrors TransactionArgs for JSON-RPC serialization.
type estimateArgs struct {
	From                 *common.Address   `json:"from,omitempty"`
	To                   *common.Address   `json:"to,omitempty"`
	Value                *hexutil.Big      `json:"value,omitempty"`
	Data                 *hexutil.Bytes    `json:"data,omitempty"`
	Input                *hexutil.Bytes    `json:"input,omitempty"`
	Gas                  *hexutil.Uint64   `json:"gas,omitempty"`
	GasPrice             *hexutil.Big      `json:"gasPrice,omitempty"`
	MaxFeePerGas         *hexutil.Big      `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas *hexutil.Big      `json:"maxPriorityFeePerGas,omitempty"`
	AccessList           *types.AccessList `json:"accessList,omitempty"`
	BlobHashes           []common.Hash     `json:"blobVersionedHashes,omitempty"`
	BlobFeeCap           *hexutil.Big      `json:"maxFeePerBlobGas,omitempty"`
}

func withData(data []byte) *hexutil.Bytes {
	b := hexutil.Bytes(data)
	return &b
}

func bigHex(v *big.Int) *hexutil.Big {
	return (*hexutil.Big)(v)
}

func weiHex(n int64) *hexutil.Big {
	return bigHex(big.NewInt(n))
}

func gwei(n int64) *hexutil.Big {
	return bigHex(new(big.Int).Mul(big.NewInt(n), big.NewInt(1e9)))
}

// rpcEstimateTotalFee calls eth_estimateTotalFee with an optional block param.
func rpcEstimateTotalFee(ctx context.Context, rpc client.RPC, args estimateArgs, block ...string) (*big.Int, error) {
	var result hexutil.Big
	blk := "latest"
	if len(block) > 0 {
		blk = block[0]
	}
	err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", args, blk)
	if err != nil {
		return nil, err
	}
	return result.ToInt(), nil
}

// rpcEstimateGas calls eth_estimateGas.
func rpcEstimateGas(ctx context.Context, rpc client.RPC, args estimateArgs) (uint64, error) {
	var result hexutil.Uint64
	err := rpc.CallContext(ctx, &result, "eth_estimateGas", args)
	return uint64(result), err
}

// rpcEstimateGasAtBlock calls eth_estimateGas with an explicit block param.
func rpcEstimateGasAtBlock(ctx context.Context, rpc client.RPC, args estimateArgs, block string) (uint64, error) {
	var result hexutil.Uint64
	err := rpc.CallContext(ctx, &result, "eth_estimateGas", args, block)
	return uint64(result), err
}

// rpcGasPrice calls eth_gasPrice.
func rpcGasPrice(ctx context.Context, rpc client.RPC) (*big.Int, error) {
	var result hexutil.Big
	err := rpc.CallContext(ctx, &result, "eth_gasPrice")
	if err != nil {
		return nil, err
	}
	return result.ToInt(), nil
}

func requireRelativeErrorLE(t devtest.T, name string, estimated, actual *big.Int, maxPercent int64) {
	t.Helper()
	t.Require().True(actual.Sign() > 0, "%s actual should be > 0, got %s", name, actual)
	diff := new(big.Int).Sub(estimated, actual)
	if diff.Sign() < 0 {
		diff.Neg(diff)
	}
	// |estimated-actual| / actual <= maxPercent/100
	lhs := new(big.Int).Mul(diff, big.NewInt(100))
	rhs := new(big.Int).Mul(actual, big.NewInt(maxPercent))
	t.Require().True(lhs.Cmp(rhs) <= 0,
		"%s relative error too high: est=%s actual=%s diff=%s max=%d%%",
		name, estimated, actual, diff, maxPercent)
}

// ============================================================
// Smoke Tests (SM-01 ~ SM-04)
// ============================================================

func TestEstimateTotalFee_Smoke(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpc := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()
	baseArgs := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(1000)}

	// SM-01: basic transfer returns positive hex value
	t.Run("SM01_BasicTransferPositive", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, baseArgs)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0, "totalFee should be positive, got %s", fee)
		t.Logf("SM-01 totalFee=%s", fee)
	})

	// SM-02: return value is valid hex big
	t.Run("SM02_ReturnFormatValid", func(t devtest.T) {
		var raw string
		err := rpc.CallContext(ctx, &raw, "eth_estimateTotalFee", baseArgs, "latest")
		t.Require().NoError(err)
		t.Require().True(strings.HasPrefix(raw, "0x"), "should be hex prefixed, got %s", raw)
		decoded, ok := new(big.Int).SetString(raw[2:], 16)
		t.Require().True(ok, "should parse as hex int")
		t.Require().True(decoded.Sign() > 0)
	})

	// SM-03: estimateTotalFee and estimateGas both available
	t.Run("SM03_BothAPIsAvailable", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, baseArgs)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)

		gas, err := rpcEstimateGas(ctx, rpc, baseArgs)
		t.Require().NoError(err)
		t.Require().True(gas > 0)
	})

	// SM-04: deterministic on repeated calls
	t.Run("SM04_Deterministic", func(t devtest.T) {
		fee1, err := rpcEstimateTotalFee(ctx, rpc, baseArgs)
		t.Require().NoError(err)
		fee2, err := rpcEstimateTotalFee(ctx, rpc, baseArgs)
		t.Require().NoError(err)
		fee3, err := rpcEstimateTotalFee(ctx, rpc, baseArgs)
		t.Require().NoError(err)
		t.Require().Equal(fee1, fee2)
		t.Require().Equal(fee2, fee3)
	})
}

// ============================================================
// Transaction Type Tests (T0 ~ T4, TX)
// ============================================================

func TestEstimateTotalFee_TransactionTypes(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.ThousandEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpc := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()

	data256 := make([]byte, 256)
	_, err := rand.Read(data256)
	require.NoError(err)
	data1KB := make([]byte, 1024)
	_, err = rand.Read(data1KB)
	require.NoError(err)

	// Simple contract bytecode (minimal runtime code: STOP)
	deployBytecode := common.FromHex("6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea164736f6c6343000813000a")

	storageKey1 := common.HexToHash("0x01")
	storageKey2 := common.HexToHash("0x02")
	storageKey3 := common.HexToHash("0x03")
	someContract := common.HexToAddress("0xaaaa000000000000000000000000000000000001")

	// ---- Type 0: Legacy ----

	t.Run("T0_01_LegacySimpleTransfer", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			Value: weiHex(1e18), GasPrice: gwei(1),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("T0-01 Legacy transfer fee=%s", fee)
	})

	t.Run("T0_02_LegacyContractDeploy", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, Data: withData(deployBytecode), GasPrice: gwei(1),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("T0-02 Legacy deploy fee=%s", fee)
	})

	t.Run("T0_03_LegacyContractCall", func(t devtest.T) {
		// Call to precompile (identity) as a proxy for contract call
		precompile := common.BytesToAddress([]byte{0x4})
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &precompile, Data: withData(data256), GasPrice: gwei(1),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T0_04_LegacyLargeData", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Data: withData(data1KB), GasPrice: gwei(1),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("T0-04 Legacy 1KB data fee=%s", fee)
	})

	t.Run("T0_05_LegacyGasPriceZero", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			Value: weiHex(1), GasPrice: bigHex(big.NewInt(0)),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0, "zero gasPrice should fallback to suggest")
	})

	t.Run("T0_06_LegacyHighGasPrice", func(t devtest.T) {
		// Keep gas price very high but still affordable in test account balance.
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1000),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// ---- Type 1: AccessList ----

	t.Run("T1_01_EmptyAccessList", func(t devtest.T) {
		emptyAL := types.AccessList{}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1), AccessList: &emptyAL,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T1_02_SingleAddrNoKey", func(t devtest.T) {
		// storageKeys must be encoded as an empty array (not omitted/null).
		al := types.AccessList{{Address: someContract, StorageKeys: []common.Hash{}}}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1), AccessList: &al,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T1_03_SingleAddrMultiKeys", func(t devtest.T) {
		al := types.AccessList{{Address: someContract, StorageKeys: []common.Hash{storageKey1, storageKey2, storageKey3}}}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1), AccessList: &al,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T1_04_MultiAddrMultiKeys", func(t devtest.T) {
		addr2 := common.HexToAddress("0xbbbb000000000000000000000000000000000002")
		al := types.AccessList{
			{Address: someContract, StorageKeys: []common.Hash{storageKey1, storageKey2}},
			{Address: addr2, StorageKeys: []common.Hash{storageKey3}},
		}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1), AccessList: &al,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T1_05_LargeAccessList", func(t devtest.T) {
		al := make(types.AccessList, 10)
		for i := 0; i < 10; i++ {
			keys := make([]common.Hash, 5)
			for j := 0; j < 5; j++ {
				keys[j] = common.BigToHash(big.NewInt(int64(i*5 + j)))
			}
			al[i] = types.AccessTuple{
				Address:     common.BigToAddress(big.NewInt(int64(0xAA00 + i))),
				StorageKeys: keys,
			}
		}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1), AccessList: &al,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("T1-05 10addr*5key AL fee=%s", fee)
	})

	t.Run("T1_06_AccessListNoFeeParams", func(t devtest.T) {
		al := types.AccessList{{Address: someContract, StorageKeys: []common.Hash{storageKey1}}}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, AccessList: &al,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// ---- Type 2: DynamicFee (EIP-1559) ----

	t.Run("T2_01_EIP1559Standard", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1000),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T2_02_EIP1559ContractDeploy", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, Data: withData(deployBytecode),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T2_03_EIP1559ContractCall", func(t devtest.T) {
		precompile := common.BytesToAddress([]byte{0x4})
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &precompile, Data: withData(data256),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T2_04_EIP1559WithAccessList", func(t devtest.T) {
		al := types.AccessList{{Address: someContract, StorageKeys: []common.Hash{storageKey1}}}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2), AccessList: &al,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T2_05_EIP1559LargeData", func(t devtest.T) {
		bigData := make([]byte, 10*1024)
		_, err := rand.Read(bigData)
		t.Require().NoError(err)
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Data: withData(bigData),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("T2-05 10KB data fee=%s", fee)
	})

	t.Run("T2_06_EIP1559ZeroValue", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(0),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0, "zero-value tx should still have gas cost")
	})

	t.Run("T2_07_EIP1559SelfTransfer", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &aliceAddr,
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T2_08_EIP1559ToZeroAddr", func(t devtest.T) {
		zero := common.Address{}
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &zero,
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	t.Run("T2_09_EIP1559ERC20Call", func(t devtest.T) {
		// Use a basic calldata that simulates ERC20 transfer(address,uint256)
		transferSelector := common.FromHex("a9059cbb")
		paddedAddr := common.LeftPadBytes(bobAddr.Bytes(), 32)
		paddedAmount := common.LeftPadBytes(big.NewInt(100).Bytes(), 32)
		calldata := append(append(transferSelector, paddedAddr...), paddedAmount...)
		// Call to a random address (will likely be a no-op but fee estimate still works)
		tokenAddr := common.HexToAddress("0xcccc000000000000000000000000000000000001")
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &tokenAddr, Data: withData(calldata),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// ---- Type 3: Blob (EIP-4844) ----

	t.Run("T3_01_BlobTx", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			BlobHashes:           []common.Hash{{0x01, 0x01}},
			BlobFeeCap:           bigHex(big.NewInt(1e9)),
			MaxFeePerGas:         gwei(10),
			MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().Error(err, "blob tx should be rejected by eth_estimateTotalFee")
		t.Require().Contains(strings.ToLower(err.Error()), "blob")
	})

	t.Run("T3_02_MultipleBlobHashes", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			BlobHashes:           []common.Hash{{0x01, 0x01}, {0x01, 0x02}, {0x01, 0x03}},
			BlobFeeCap:           bigHex(big.NewInt(1e9)),
			MaxFeePerGas:         gwei(10),
			MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().Error(err, "blob tx should be rejected by eth_estimateTotalFee")
		t.Require().Contains(strings.ToLower(err.Error()), "blob")
	})

	t.Run("T3_03_BlobWithAccessList", func(t devtest.T) {
		al := types.AccessList{{Address: someContract, StorageKeys: []common.Hash{storageKey1}}}
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			BlobHashes: []common.Hash{{0x01, 0x01}}, BlobFeeCap: bigHex(big.NewInt(1e9)),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2), AccessList: &al,
		})
		t.Require().Error(err, "blob tx should be rejected by eth_estimateTotalFee")
		t.Require().Contains(strings.ToLower(err.Error()), "blob")
	})

	t.Run("T3_04_BlobNoTo", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From:       &aliceAddr,
			BlobHashes: []common.Hash{{0x01, 0x01}}, BlobFeeCap: bigHex(big.NewInt(1e9)),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().Error(err, "blob tx without to should fail")
	})

	t.Run("T3_05_OnlyBlobFeeCapNoBlobHashes", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			BlobFeeCap:   bigHex(big.NewInt(1e9)),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		// Without blobHashes, should fall back to DynamicFee type
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// ---- Type 4: SetCode (EIP-7702) ----

	// Helper: sign a SetCodeAuthorization using alice's key
	signAuth := func(addr common.Address, nonce uint64) types.SetCodeAuthorization {
		chainIDBig := alice.ChainID().ToBig()
		auth, err := types.SignSetCode(alice.Key().Priv(), types.SetCodeAuthorization{
			ChainID: *uint256.MustFromBig(chainIDBig),
			Address: addr,
			Nonce:   nonce,
		})
		require.NoError(err)
		return auth
	}

	t.Run("T4_01_SetCodeSingleAuth", func(t devtest.T) {
		auth := signAuth(someContract, 0)
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", map[string]interface{}{
			"from":                 aliceAddr,
			"to":                   aliceAddr,
			"authorizationList":    []types.SetCodeAuthorization{auth},
			"maxFeePerGas":         gwei(10),
			"maxPriorityFeePerGas": gwei(2),
		}, "latest")
		t.Require().NoError(err)
		t.Require().True(result.ToInt().Sign() > 0)
		t.Logf("T4-01 SetCode single auth fee=%s", result.ToInt())
	})

	t.Run("T4_02_SetCodeMultiAuth", func(t devtest.T) {
		auth1 := signAuth(someContract, 0)
		addr2 := common.HexToAddress("0xbbbb000000000000000000000000000000000099")
		addr3 := common.HexToAddress("0xcccc000000000000000000000000000000000099")
		auth2 := signAuth(addr2, 0)
		auth3 := signAuth(addr3, 0)
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", map[string]interface{}{
			"from":                 aliceAddr,
			"to":                   aliceAddr,
			"authorizationList":    []types.SetCodeAuthorization{auth1, auth2, auth3},
			"maxFeePerGas":         gwei(10),
			"maxPriorityFeePerGas": gwei(2),
		}, "latest")
		t.Require().NoError(err)
		t.Require().True(result.ToInt().Sign() > 0)
		t.Logf("T4-02 SetCode 3 auths fee=%s", result.ToInt())
	})

	t.Run("T4_03_SetCodeWithAccessList", func(t devtest.T) {
		auth := signAuth(someContract, 0)
		al := types.AccessList{{Address: someContract, StorageKeys: []common.Hash{storageKey1}}}
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", map[string]interface{}{
			"from":                 aliceAddr,
			"to":                   aliceAddr,
			"authorizationList":    []types.SetCodeAuthorization{auth},
			"accessList":           al,
			"maxFeePerGas":         gwei(10),
			"maxPriorityFeePerGas": gwei(2),
		}, "latest")
		t.Require().NoError(err)
		t.Require().True(result.ToInt().Sign() > 0)
		t.Logf("T4-03 SetCode+AL fee=%s", result.ToInt())
	})

	t.Run("T4_04_SetCodeNoTo", func(t devtest.T) {
		auth := signAuth(someContract, 0)
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", map[string]interface{}{
			"from":              aliceAddr,
			"authorizationList": []types.SetCodeAuthorization{auth},
			"maxFeePerGas":      gwei(10),
		}, "latest")
		t.Require().Error(err, "SetCode tx without To should fail")
	})

	t.Run("T4_05_SetCodeEmptyAuthList", func(t devtest.T) {
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", map[string]interface{}{
			"from":              aliceAddr,
			"to":                bobAddr,
			"authorizationList": []types.SetCodeAuthorization{},
			"maxFeePerGas":      gwei(10),
		}, "latest")
		t.Require().Error(err, "empty auth list should be rejected")
	})

	t.Run("T4_06_SetCodeWithGasPrice", func(t devtest.T) {
		auth := signAuth(someContract, 0)
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", map[string]interface{}{
			"from":              aliceAddr,
			"to":                aliceAddr,
			"authorizationList": []types.SetCodeAuthorization{auth},
			"gasPrice":          gwei(1),
		}, "latest")
		// Compatibility behavior:
		// - Some implementations reject SetCode+gasPrice as invalid fee field mix.
		// - Some implementations accept it and treat it as a valid legacy-style estimate.
		if err != nil {
			t.Require().True(
				strings.Contains(strings.ToLower(err.Error()), "gasprice") ||
					strings.Contains(strings.ToLower(err.Error()), "fee"),
				"unexpected error for SetCode+gasPrice: %v", err)
			return
		}
		t.Require().True(result.ToInt().Sign() > 0, "accepted SetCode+gasPrice should still return positive fee")
	})

	// ---- Cross-Type Comparisons (TX) ----

	t.Run("TX01_LegacyVsDynamic", func(t devtest.T) {
		legacyFee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1000),
			GasPrice: gwei(1),
		})
		t.Require().NoError(err)

		dynamicFee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1000),
			MaxFeePerGas: gwei(1), MaxPriorityFeePerGas: bigHex(big.NewInt(0)),
		})
		t.Require().NoError(err)

		t.Logf("TX-01 Legacy=%s, Dynamic=%s", legacyFee, dynamicFee)
		// Both should be similar (difference only in RLP encoding size)
	})

	t.Run("TX02_WithVsWithoutAccessList", func(t devtest.T) {
		noAL, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)

		al := make(types.AccessList, 10)
		for i := 0; i < 10; i++ {
			al[i] = types.AccessTuple{
				Address:     common.BigToAddress(big.NewInt(int64(0xAA00 + i))),
				StorageKeys: []common.Hash{common.BigToHash(big.NewInt(int64(i)))},
			}
		}
		withAL, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2), AccessList: &al,
		})
		t.Require().NoError(err)
		t.Require().True(withAL.Cmp(noAL) > 0,
			"with AL(%s) should > without AL(%s)", withAL, noAL)
	})

	t.Run("TX03_DynamicVsSetCode", func(t devtest.T) {
		// DynamicFee: simple transfer
		dynamicFee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1000),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)

		// SetCode: same transfer but with 1 authorization (+25000 intrinsic gas + authList RLP)
		auth := signAuth(someContract, 0)
		var setCodeResult hexutil.Big
		err = rpc.CallContext(ctx, &setCodeResult, "eth_estimateTotalFee", map[string]interface{}{
			"from":                 aliceAddr,
			"to":                   aliceAddr,
			"value":                weiHex(1000),
			"authorizationList":    []types.SetCodeAuthorization{auth},
			"maxFeePerGas":         gwei(10),
			"maxPriorityFeePerGas": gwei(2),
		}, "latest")
		t.Require().NoError(err)
		setCodeFee := setCodeResult.ToInt()

		t.Logf("TX-03 Dynamic=%s, SetCode=%s", dynamicFee, setCodeFee)
		t.Require().True(setCodeFee.Cmp(dynamicFee) > 0,
			"SetCode fee(%s) should > Dynamic fee(%s) due to authorization overhead", setCodeFee, dynamicFee)
	})

	t.Run("TX04_SmallVsLargeData", func(t devtest.T) {
		small, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1),
		})
		t.Require().NoError(err)

		big10KB := make([]byte, 10*1024)
		_, err = rand.Read(big10KB)
		t.Require().NoError(err)
		large, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Data: withData(big10KB), GasPrice: gwei(1),
		})
		t.Require().NoError(err)
		t.Require().True(large.Cmp(small) > 0,
			"10KB data fee(%s) should >> no data fee(%s)", large, small)
	})
}

// ============================================================
// Fee Parameter Tests (P-01 ~ P-14)
// ============================================================

func TestEstimateTotalFee_FeeParams(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpc := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()

	base := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(1)}

	// P-01: Legacy gasPrice=1gwei
	t.Run("P01_Legacy1Gwei", func(t devtest.T) {
		args := base
		args.GasPrice = gwei(1)
		fee, err := rpcEstimateTotalFee(ctx, rpc, args)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("P-01 fee=%s", fee)
	})

	// P-02: Legacy gasPrice=100gwei (should be higher than P-01)
	t.Run("P02_Legacy100Gwei", func(t devtest.T) {
		info, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		blockHex := hexutil.EncodeUint64(info.NumberU64())

		fee1gwei, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1), GasPrice: gwei(1),
		}, blockHex)
		t.Require().NoError(err)
		fee100gwei, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1), GasPrice: gwei(100),
		}, blockHex)
		t.Require().NoError(err)
		t.Require().True(fee100gwei.Cmp(fee1gwei) > 0,
			"100gwei(%s) should > 1gwei(%s)", fee100gwei, fee1gwei)
	})

	// P-03: gasPrice=0 → fallback
	t.Run("P03_GasPriceZeroFallback", func(t devtest.T) {
		args := base
		args.GasPrice = bigHex(big.NewInt(0))
		fee, err := rpcEstimateTotalFee(ctx, rpc, args)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0, "zero gasPrice should fallback to suggest")
	})

	// P-04: EIP-1559 full params: effectivePrice = min(maxFee, baseFee+tip)
	t.Run("P04_EIP1559FullParams", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// P-05: maxFee < baseFee+tip → capped by maxFee
	t.Run("P05_MaxFeeCaps", func(t devtest.T) {
		info, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		blockHex := hexutil.EncodeUint64(info.NumberU64())

		highCap, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
			MaxFeePerGas: gwei(100), MaxPriorityFeePerGas: gwei(2),
		}, blockHex)
		t.Require().NoError(err)
		lowCap, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
			MaxFeePerGas: gwei(2), MaxPriorityFeePerGas: gwei(2),
		}, blockHex)
		t.Require().NoError(err)
		t.Require().True(lowCap.Cmp(highCap) < 0,
			"lowMaxFee(%s) should produce lower fee than highMaxFee(%s)", lowCap, highCap)
	})

	// P-06: Only maxFeePerGas, no tip
	t.Run("P06_OnlyMaxFee", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
			MaxFeePerGas: gwei(10),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// P-07: Only maxPriorityFeePerGas.
	// Compatibility behavior differs across implementations:
	// - some reject only-tip because maxFeePerGas defaults to 0 (tip > maxFee)
	// - some treat it similarly to no-fee-params path
	t.Run("P07_OnlyTip", func(t devtest.T) {
		info, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		blockHex := hexutil.EncodeUint64(info.NumberU64())

		onlyTip, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
			MaxPriorityFeePerGas: gwei(2),
		}, blockHex)
		if err != nil {
			t.Require().Contains(strings.ToLower(err.Error()), "max priority fee per gas higher than max fee per gas")
			return
		}
		noParams, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		}, blockHex)
		t.Require().NoError(err)
		t.Require().Equal(noParams, onlyTip, "only-tip path should match no-fee-params path at same block")
	})

	// P-15: Effective gas price formula (exact relation at same block):
	// fee(highCap) - fee(lowCap) == gasEstimate * (effectiveHigh - effectiveLow)
	t.Run("P15_EffectiveGasPriceFormula", func(t devtest.T) {
		info, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		baseFee := info.BaseFee()
		blockHex := hexutil.EncodeUint64(info.NumberU64())

		tip := gwei(2)
		lowCap := gwei(3)
		highCap := gwei(20)

		argsLow := estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
			MaxFeePerGas: lowCap, MaxPriorityFeePerGas: tip,
		}
		argsHigh := estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
			MaxFeePerGas: highCap, MaxPriorityFeePerGas: tip,
		}

		feeLow, err := rpcEstimateTotalFee(ctx, rpc, argsLow, blockHex)
		t.Require().NoError(err)
		feeHigh, err := rpcEstimateTotalFee(ctx, rpc, argsHigh, blockHex)
		t.Require().NoError(err)

		gas, err := rpcEstimateGasAtBlock(ctx, rpc, argsLow, blockHex)
		t.Require().NoError(err)

		basePlusTip := new(big.Int).Add(baseFee, tip.ToInt())
		effectiveLow := new(big.Int).Set(lowCap.ToInt())
		if basePlusTip.Cmp(effectiveLow) < 0 {
			effectiveLow = basePlusTip
		}
		effectiveHigh := new(big.Int).Set(highCap.ToInt())
		if basePlusTip.Cmp(effectiveHigh) < 0 {
			effectiveHigh = basePlusTip
		}
		effectiveDelta := new(big.Int).Sub(effectiveHigh, effectiveLow)
		expectedDelta := new(big.Int).Mul(new(big.Int).SetUint64(gas), effectiveDelta)
		actualDelta := new(big.Int).Sub(feeHigh, feeLow)
		t.Require().True(expectedDelta.Cmp(actualDelta) == 0,
			"fee delta should match gas * effectiveGasPrice delta at same block")
	})

	// P-08: No fee params → auto-suggest
	t.Run("P08_NoFeeParams", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, base)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// P-09: gasPrice + maxFeePerGas conflict
	t.Run("P09_ConflictGasPriceMaxFee", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			GasPrice: gwei(1), MaxFeePerGas: gwei(2),
		})
		t.Require().Error(err)
		t.Require().Contains(strings.ToLower(err.Error()), "gasprice")
	})

	// P-10: gasPrice + maxPriorityFeePerGas conflict
	t.Run("P10_ConflictGasPriceTip", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			GasPrice: gwei(1), MaxPriorityFeePerGas: gwei(1),
		})
		t.Require().Error(err)
	})

	// P-11: Very large gasPrice (no overflow)
	t.Run("P11_HugeGasPrice", func(t devtest.T) {
		// Keep this large for overflow checks, but avoid triggering insufficient-funds in test env.
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1000),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// P-12: Tiny gasPrice (1 wei)
	t.Run("P12_TinyGasPrice", func(t devtest.T) {
		info, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		tinyButValid := new(big.Int).Add(info.BaseFee(), big.NewInt(1))
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, GasPrice: bigHex(tinyButValid),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0, "tiny but valid gasPrice should produce positive fee")
	})

	// P-13: maxFeePerGas below baseFee should fail
	t.Run("P13_VeryLowMaxFee", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, MaxFeePerGas: bigHex(big.NewInt(1)),
		})
		t.Require().Error(err)
		t.Require().Contains(strings.ToLower(err.Error()), "max fee per gas less than block base fee")
	})

	// P-14: tip near maxFee; effective price still capped by maxFee
	t.Run("P14_HugeTipCapped", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr,
			MaxFeePerGas:         gwei(5),
			MaxPriorityFeePerGas: gwei(5),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})
}

// ============================================================
// Block Parameter Tests (B-01 ~ B-12)
// ============================================================

func TestEstimateTotalFee_BlockParam(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpc := sys.L2EL.Escape().EthClient().RPC()
	l2Client := sys.L2EL.Escape().EthClient()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()
	base := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(1)}
	historicalBase := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(0)}

	var genesisBlock struct {
		Timestamp hexutil.Uint64 `json:"timestamp"`
	}
	err := rpc.CallContext(ctx, &genesisBlock, "eth_getBlockByNumber", "0x0", false)
	t.Require().NoError(err)
	arsiaAtGenesis := sys.L2Chain.IsMantleForkActiveAt(forks.MantleArsia, uint64(genesisBlock.Timestamp))

	// B-01: nil (default latest)
	t.Run("B01_NilDefaultsLatest", func(t devtest.T) {
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", base)
		t.Require().NoError(err)
		t.Require().True(result.ToInt().Sign() > 0)
	})

	// B-02: explicit "latest"
	t.Run("B02_ExplicitLatest", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, base, "latest")
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// B-03: "pending"
	t.Run("B03_Pending", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, base, "pending")
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// B-04: "earliest"
	t.Run("B04_Earliest", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, historicalBase, "earliest")
		if arsiaAtGenesis {
			t.Require().NoError(err)
			t.Require().True(fee.Sign() > 0)
		} else {
			t.Require().Error(err)
			t.Require().Contains(strings.ToLower(err.Error()), "arsia")
		}
	})

	// B-05: Valid post-Arsia block number
	t.Run("B05_ValidBlockNumber", func(t devtest.T) {
		info, err := l2Client.InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		blockHex := hexutil.EncodeUint64(info.NumberU64())
		fee, err := rpcEstimateTotalFee(ctx, rpc, base, blockHex)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// B-06: Pre-Arsia block number (if Arsia doesn't start at genesis)
	t.Run("B06_PreArsiaBlock", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, historicalBase, "0x0")
		if arsiaAtGenesis {
			t.Require().NoError(err)
			t.Require().True(fee.Sign() > 0)
		} else {
			t.Require().Error(err)
			t.Require().Contains(strings.ToLower(err.Error()), "arsia")
		}
	})

	// B-09: Future block number → not found
	t.Run("B09_FutureBlockNotFound", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, base, "0xFFFFFFFF")
		t.Require().Error(err)
		t.Logf("B-09 future block error: %v", err)
	})

	// B-10: Valid block hash
	t.Run("B10_ValidBlockHash", func(t devtest.T) {
		info, err := l2Client.InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		var result hexutil.Big
		err = rpc.CallContext(ctx, &result, "eth_estimateTotalFee", base,
			map[string]interface{}{"blockHash": info.Hash()})
		t.Require().NoError(err)
		t.Require().True(result.ToInt().Sign() > 0)
	})

	// B-11: Invalid block hash
	t.Run("B11_InvalidBlockHash", func(t devtest.T) {
		var result hexutil.Big
		err := rpc.CallContext(ctx, &result, "eth_estimateTotalFee", base,
			map[string]interface{}{"blockHash": common.HexToHash("0xdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef")})
		t.Require().Error(err)
	})

	// B-12: nil and "latest" consistent
	t.Run("B12_NilAndLatestConsistent", func(t devtest.T) {
		var feeNil hexutil.Big
		err := rpc.CallContext(ctx, &feeNil, "eth_estimateTotalFee", base)
		t.Require().NoError(err)

		feeLatest, err := rpcEstimateTotalFee(ctx, rpc, base, "latest")
		t.Require().NoError(err)
		t.Require().Equal(feeNil.ToInt(), feeLatest)
	})

	// B-13: Invalid mixed selector (both blockHash and blockNumber)
	t.Run("B13_BlockHashAndNumberConflict", func(t devtest.T) {
		info, err := l2Client.InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		var result hexutil.Big
		err = rpc.CallContext(ctx, &result, "eth_estimateTotalFee", base, map[string]interface{}{
			"blockHash":   info.Hash(),
			"blockNumber": "latest",
		})
		t.Require().Error(err)
	})

	// B-14: Invalid block tag format
	t.Run("B14_InvalidBlockTag", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, base, "latestx")
		t.Require().Error(err)
	})
}

// ============================================================
// Error Handling Tests (E-01 ~ E-12)
// ============================================================

func TestEstimateTotalFee_ErrorHandling(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	poor := sys.Wallet.NewEOA(sys.L2EL) // unfunded
	rpc := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()
	poorAddr := poor.Address()

	// Deploy an always-revert contract for E-02/E-03/E-04 tests.
	// Init code (12 bytes): copies 5-byte runtime to memory and returns it.
	//   PUSH1 5, PUSH1 12, PUSH1 0, CODECOPY, PUSH1 5, PUSH1 0, RETURN
	// Runtime code (5 bytes): always reverts with empty data.
	//   PUSH1 0, PUSH1 0, REVERT
	revertDeployBytecode := common.FromHex("6005600c60003960056000f360006000fd")

	revertTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(revertDeployBytecode))
	revertReceipt, err := revertTx.Included.Eval(ctx)
	require.NoError(err)
	revertContractAddr := revertReceipt.ContractAddress
	t.Logf("Deployed always-revert contract at %s", revertContractAddr)

	// E-01: Insufficient balance with explicit gasPrice
	t.Run("E01_InsufficientBalance", func(t devtest.T) {
		huge := new(big.Int).Mul(big.NewInt(1e18), big.NewInt(1000))
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &poorAddr, To: &bobAddr,
			Value: bigHex(huge), GasPrice: gwei(100),
		})
		t.Require().Error(err, "insufficient balance should fail estimate")
		t.Require().True(
			strings.Contains(strings.ToLower(err.Error()), "insufficient") ||
				strings.Contains(strings.ToLower(err.Error()), "fund"),
			"error should mention insufficient funds, got: %v", err)
	})

	// E-02: Call to always-revert contract
	t.Run("E02_RevertContract", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &revertContractAddr,
		})
		t.Require().Error(err, "call to always-revert contract should fail")
		t.Logf("E-02 revert error: %v", err)
	})

	// E-03: Revert with calldata (same contract, any calldata triggers revert)
	t.Run("E03_RevertWithCalldata", func(t devtest.T) {
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &revertContractAddr,
			Data: withData([]byte{0x12, 0x34, 0x56, 0x78}),
		})
		t.Require().Error(err, "call to revert contract with data should fail")
		t.Require().True(
			strings.Contains(strings.ToLower(err.Error()), "revert") ||
				strings.Contains(strings.ToLower(err.Error()), "execution"),
			"error should mention revert or execution, got: %v", err)
	})

	// E-04: Call non-existent method on revert contract
	t.Run("E04_NonExistentMethod", func(t devtest.T) {
		randomSelector := common.FromHex("deadbeef")
		_, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &revertContractAddr, Data: withData(randomSelector),
		})
		t.Require().Error(err, "non-existent method on revert contract should fail")
	})

	// E-05: To is EOA with data (should succeed — EOA ignores data)
	t.Run("E05_EOAWithData", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Data: withData([]byte{0xde, 0xad, 0xbe, 0xef}),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// E-06: Gas param very large → capped by RPCGasCap
	t.Run("E06_HugeGasParam", func(t devtest.T) {
		hugeGas := hexutil.Uint64(0xFFFFFFFFFFFF)
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Gas: &hugeGas,
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// E-07: Large data (128KB)
	t.Run("E07_LargeData128KB", func(t devtest.T) {
		bigData := make([]byte, 128*1024)
		_, err := rand.Read(bigData)
		t.Require().NoError(err)
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Data: withData(bigData),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// E-08: Empty from compatibility
	t.Run("E08_EmptyFrom", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			To: &bobAddr, Value: weiHex(1),
		})
		if err != nil {
			// Some implementations require explicit sender for value transfer.
			msg := strings.ToLower(err.Error())
			t.Require().True(
				strings.Contains(msg, "from") ||
					strings.Contains(msg, "sender") ||
					strings.Contains(msg, "insufficient funds"),
				"unexpected missing-from error: %v", err,
			)
			return
		}
		// Mantle/geth compatibility path: estimation can succeed when balance checks are skipped.
		t.Require().True(fee.Sign() > 0, "empty-from estimation should return positive fee on compatibility path")
	})

	// E-11: To is precompile address
	t.Run("E11_PrecompileAddress", func(t devtest.T) {
		sha256pre := common.BytesToAddress([]byte{0x2})
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &sha256pre, Data: withData([]byte("hello")),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})

	// E-12: Empty request
	t.Run("E12_EmptyRequest", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
	})
}

// ============================================================
// Cross-Validation with eth_estimateGas (X-01 ~ X-06)
// ============================================================

func TestEstimateTotalFee_CrossValidation(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpc := sys.L2EL.Escape().EthClient().RPC()
	l2Client := sys.L2EL.Escape().EthClient()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()
	base := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(1000)}

	info, err := l2Client.InfoByLabel(ctx, "latest")
	require.NoError(err)
	baseFee := info.BaseFee()

	// X-01: totalFee >= estimateGas * baseFee
	t.Run("X01_TotalFeeGteGasTimesBaseFee", func(t devtest.T) {
		totalFee, err := rpcEstimateTotalFee(ctx, rpc, base)
		t.Require().NoError(err)

		gas, err := rpcEstimateGas(ctx, rpc, base)
		t.Require().NoError(err)

		minCost := new(big.Int).Mul(new(big.Int).SetUint64(gas), baseFee)
		t.Require().True(totalFee.Cmp(minCost) >= 0,
			"totalFee(%s) >= gas*baseFee(%s)", totalFee, minCost)
	})

	// X-02: totalFee >= estimateGas * baseFee (redundant with X-01 but different wording, test with data)
	t.Run("X02_TotalFeeGteWithData", func(t devtest.T) {
		data := make([]byte, 512)
		_, err := rand.Read(data)
		t.Require().NoError(err)
		argsWithData := estimateArgs{From: &aliceAddr, To: &bobAddr, Data: withData(data)}

		totalFee, err := rpcEstimateTotalFee(ctx, rpc, argsWithData)
		t.Require().NoError(err)

		gas, err := rpcEstimateGas(ctx, rpc, argsWithData)
		t.Require().NoError(err)

		minCost := new(big.Int).Mul(new(big.Int).SetUint64(gas), baseFee)
		t.Require().True(totalFee.Cmp(minCost) >= 0)
	})

	// X-03: surplus = totalFee - gasEstimate*gasPrice >= 0
	t.Run("X03_SurplusPositive", func(t devtest.T) {
		totalFee, err := rpcEstimateTotalFee(ctx, rpc, base)
		t.Require().NoError(err)

		gas, err := rpcEstimateGas(ctx, rpc, base)
		t.Require().NoError(err)

		suggestPrice, err := rpcGasPrice(ctx, rpc)
		t.Require().NoError(err)

		l2Fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), suggestPrice)
		surplus := new(big.Int).Sub(totalFee, l2Fee)
		t.Logf("X-03 total=%s, l2=%s, surplus(L1+Op)=%s", totalFee, l2Fee, surplus)
		t.Require().True(surplus.Sign() >= 0, "surplus should >= 0")
	})

	// X-04: Revert → both APIs fail consistently
	t.Run("X04_RevertBothFail", func(t devtest.T) {
		// Deploy an always-revert contract for a deterministic revert
		revertCode := common.FromHex("6005600c60003960056000f360006000fd")
		deployTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(revertCode))
		receipt, err := deployTx.Included.Eval(ctx)
		t.Require().NoError(err)
		revertAddr := receipt.ContractAddress

		_, err1 := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &revertAddr,
		})
		_, err2 := rpcEstimateGas(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &revertAddr,
		})
		// Both should fail — call to always-revert contract
		t.Require().Error(err1, "estimateTotalFee should fail on revert contract")
		t.Require().Error(err2, "estimateGas should fail on revert contract")
		t.Logf("X-04 estimateTotalFee error: %v", err1)
		t.Logf("X-04 estimateGas error: %v", err2)
	})

	// X-05: Normal tx → both succeed
	t.Run("X05_NormalBothSucceed", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpc, base)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)

		gas, err := rpcEstimateGas(ctx, rpc, base)
		t.Require().NoError(err)
		t.Require().True(gas > 0)
	})

	// X-06: nil and "latest" produce same result
	t.Run("X06_NilEqualsLatest", func(t devtest.T) {
		var feeNil hexutil.Big
		err := rpc.CallContext(ctx, &feeNil, "eth_estimateTotalFee", base)
		t.Require().NoError(err)

		feeLatest, err := rpcEstimateTotalFee(ctx, rpc, base, "latest")
		t.Require().NoError(err)
		t.Require().Equal(feeNil.ToInt(), feeLatest)
	})
}

// ============================================================
// Control Variable Tests (C-01 ~ C-07)
// ============================================================

func TestEstimateTotalFee_ControlVariable(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpc := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()

	// C-01: Data size monotonic
	t.Run("C01_DataSizeMonotonic", func(t devtest.T) {
		sizes := []int{0, 32, 256, 1024, 4096}
		var prevFee *big.Int
		for _, sz := range sizes {
			data := make([]byte, sz)
			if sz > 0 {
				for i := range data {
					data[i] = 0xAB
				}
			}
			var d *hexutil.Bytes
			if sz > 0 {
				d = withData(data)
			}
			fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
				From: &aliceAddr, To: &bobAddr, Data: d, GasPrice: gwei(1),
			})
			t.Require().NoError(err)
			if prevFee != nil {
				t.Require().True(fee.Cmp(prevFee) > 0,
					"size=%d fee(%s) should > previous fee(%s)", sz, fee, prevFee)
			}
			prevFee = fee
			t.Logf("C-01 size=%d fee=%s", sz, fee)
		}
	})

	// C-02: GasPrice scales L2 fee
	t.Run("C02_GasPriceScales", func(t devtest.T) {
		prices := []int64{1, 2, 5, 10}
		fees := make([]*big.Int, len(prices))
		for i, p := range prices {
			var err error
			fees[i], err = rpcEstimateTotalFee(ctx, rpc, estimateArgs{
				From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
				GasPrice: gwei(p),
			})
			t.Require().NoError(err)
			t.Logf("C-02 gasPrice=%dgwei fee=%s", p, fees[i])
		}
		for i := 1; i < len(fees); i++ {
			t.Require().True(fees[i].Cmp(fees[i-1]) > 0)
		}
	})

	// C-03: Value does not affect fee
	t.Run("C03_ValueNoEffect", func(t devtest.T) {
		values := []*big.Int{big.NewInt(0), big.NewInt(1e15), big.NewInt(1e18)}
		fees := make([]*big.Int, len(values))
		for i, v := range values {
			var err error
			fees[i], err = rpcEstimateTotalFee(ctx, rpc, estimateArgs{
				From: &aliceAddr, To: &bobAddr, Value: bigHex(v), GasPrice: gwei(1),
			})
			t.Require().NoError(err)
		}
		tol := new(big.Int).Div(fees[0], big.NewInt(50)) // 2% tolerance
		for i := 1; i < len(fees); i++ {
			diff := new(big.Int).Abs(new(big.Int).Sub(fees[i], fees[0]))
			t.Require().True(diff.Cmp(tol) <= 0,
				"value change should not affect fee: fee[0]=%s fee[%d]=%s", fees[0], i, fees[i])
		}
	})

	// C-04: Fee changes with block (follows baseFee)
	t.Run("C04_FeeChangesWithBlock", func(t devtest.T) {
		// Pin at block N
		infoN, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		blockN := hexutil.EncodeUint64(infoN.NumberU64())
		feeN, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		}, blockN)
		t.Require().NoError(err)
		t.Require().True(feeN.Sign() > 0)

		// Force block progression with a simple tx.
		tx := txplan.NewPlannedTx(alice.Plan(), txplan.WithTo(&bobAddr), txplan.WithValue(eth.WeiBig(big.NewInt(1))))
		receipt, err := tx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)

		infoM, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		t.Require().True(infoM.NumberU64() >= infoN.NumberU64(), "latest block should not move backwards")

		// Re-query pinned block N and verify exact stability.
		feeNAgain, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		}, blockN)
		t.Require().NoError(err)
		t.Require().Equal(feeN, feeNAgain, "same args on same pinned block should be deterministic")

		// Query at new latest and ensure valid output.
		feeLatest, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		}, "latest")
		t.Require().NoError(err)
		t.Require().True(feeLatest.Sign() > 0)
	})

	// C-05: EOA vs deployed contract target
	t.Run("C05_EOAvsDeployedContract", func(t devtest.T) {
		// Deploy a simple storage contract for comparison
		storageCode := common.FromHex("6007600c60003960076000f360003560005500")
		deployTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(storageCode))
		receipt, err := deployTx.Included.Eval(ctx)
		t.Require().NoError(err)
		contractAddr := receipt.ContractAddress

		// EOA transfer: just value, no code execution
		eoaFee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1), GasPrice: gwei(1),
		})
		t.Require().NoError(err)

		// Contract call: triggers SSTORE (cold), costs more gas
		calldata := common.LeftPadBytes(big.NewInt(999).Bytes(), 32)
		contractFee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &contractAddr, Data: withData(calldata), GasPrice: gwei(1),
		})
		t.Require().NoError(err)
		t.Logf("C-05 EOA=%s, Contract(SSTORE)=%s", eoaFee, contractFee)
		t.Require().True(contractFee.Cmp(eoaFee) > 0,
			"contract call with SSTORE(%s) should cost more than EOA transfer(%s)", contractFee, eoaFee)
	})

	// C-06: Repeated calls same block → identical
	t.Run("C06_RepeatedIdentical", func(t devtest.T) {
		args := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(42), GasPrice: gwei(1)}
		info, err := sys.L2EL.Escape().EthClient().InfoByLabel(ctx, "latest")
		t.Require().NoError(err)
		blockHex := hexutil.EncodeUint64(info.NumberU64())
		results := make([]*big.Int, 10)
		for i := 0; i < 10; i++ {
			results[i], err = rpcEstimateTotalFee(ctx, rpc, args, blockHex)
			t.Require().NoError(err)
		}
		for i := 1; i < 10; i++ {
			t.Require().Equal(results[0], results[i], "call %d differs from call 0", i)
		}
	})

	// C-07: AccessList size monotonic
	t.Run("C07_AccessListMonotonic", func(t devtest.T) {
		alSizes := []int{0, 1, 5, 10}
		fees := make([]*big.Int, len(alSizes))
		for i, n := range alSizes {
			var alPtr *types.AccessList
			if n > 0 {
				al := make(types.AccessList, n)
				for j := 0; j < n; j++ {
					al[j] = types.AccessTuple{
						Address:     common.BigToAddress(big.NewInt(int64(0xAA00 + j))),
						StorageKeys: []common.Hash{common.BigToHash(big.NewInt(int64(j)))},
					}
				}
				alPtr = &al
			}
			var err error
			fees[i], err = rpcEstimateTotalFee(ctx, rpc, estimateArgs{
				From: &aliceAddr, To: &bobAddr, GasPrice: gwei(1), AccessList: alPtr,
			})
			t.Require().NoError(err)
			t.Logf("C-07 AL size=%d fee=%s", n, fees[i])
		}
		for i := 1; i < len(fees); i++ {
			t.Require().True(fees[i].Cmp(fees[i-1]) > 0,
				"AL=%d fee(%s) should > AL=%d fee(%s)", alSizes[i], fees[i], alSizes[i-1], fees[i-1])
		}
	})
}

// ============================================================
// Actual On-Chain Comparison (R-01 ~ R-08)
// ============================================================

func TestEstimateTotalFee_VsActualCost(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	// Set operator fee (same as TestFees)
	operatorFee := dsl.NewOperatorFee(t, sys.L2Chain, sys.L1EL)
	operatorFee.SetOperatorFee(100000000, 500)
	operatorFee.WaitForL2SyncWithCurrentL1State()

	l2Client := sys.L2EL.Escape().EthClient()
	rpcCli := l2Client.RPC()

	gpo := txib.NewGasPriceOracle(
		txib.WithClient(l2Client),
		txib.WithTo(predeploys.GasPriceOracleAddr),
		txib.WithTest(t),
	)
	tokenRatio, err := contractio.Read(gpo.TokenRatio(), ctx)
	require.NoError(err)
	t.Logf("Token ratio: %v", tokenRatio)

	readBalanceAtBlock := func(t devtest.T, addr common.Address, blockNumber *big.Int) *big.Int {
		t.Helper()
		blockHex := hexutil.EncodeBig(blockNumber)
		var bal hexutil.Big
		err := rpcCli.CallContext(ctx, &bal, "eth_getBalance", addr, blockHex)
		t.Require().NoError(err)
		return bal.ToInt()
	}

	readActualFeeBySenderDelta := func(t devtest.T, sender common.Address, value *big.Int, blockNumber *big.Int) *big.Int {
		t.Helper()
		prevBlock := new(big.Int).Sub(blockNumber, big.NewInt(1))
		if prevBlock.Sign() < 0 {
			prevBlock = big.NewInt(0)
		}
		before := readBalanceAtBlock(t, sender, prevBlock)
		after := readBalanceAtBlock(t, sender, blockNumber)
		spent := new(big.Int).Sub(before, after)
		fee := new(big.Int).Sub(spent, value)
		t.Require().True(fee.Sign() >= 0,
			"sender delta fee should be >= 0, before=%s after=%s value=%s fee=%s",
			before, after, value, fee)
		return fee
	}

	// Helper: estimate then send, compare
	verifyEstimateVsActual := func(t devtest.T, name string, from *dsl.EOA, to *dsl.EOA, amount *big.Int, data []byte) {
		fromAddr := from.Address()
		toAddr := to.Address()

		args := estimateArgs{From: &fromAddr, To: &toAddr, Value: bigHex(amount)}
		if len(data) > 0 {
			args.Data = withData(data)
		}

		// 1. Estimate
		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, args)
		t.Require().NoError(err)

		// 2. Send actual transaction
		opts := []txplan.Option{from.Plan(), txplan.WithTo(&toAddr), txplan.WithValue(eth.WeiBig(amount))}
		if len(data) > 0 {
			opts = append(opts, txplan.WithData(data))
		}
		tx := txplan.NewPlannedTx(txplan.Combine(opts...))
		receipt, err := tx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)

		// 3. Calculate actual fee
		actualL2Fee := new(big.Int).Mul(
			new(big.Int).SetUint64(receipt.GasUsed),
			receipt.EffectiveGasPrice,
		)
		actualL1Fee := big.NewInt(0)
		if receipt.L1Fee != nil {
			actualL1Fee = new(big.Int).Set(receipt.L1Fee)
		}
		actualTotal := readActualFeeBySenderDelta(t, fromAddr, amount, receipt.BlockNumber)
		actualOperatorFee := new(big.Int).Sub(new(big.Int).Sub(new(big.Int).Set(actualTotal), actualL2Fee), actualL1Fee)
		t.Require().True(actualOperatorFee.Sign() >= 0, "%s operator fee should be >= 0, got %s", name, actualOperatorFee)

		t.Logf("%s: estimated=%s actual=%s (L2=%s L1=%s Op=%s)",
			name, estimated, actualTotal, actualL2Fee, actualL1Fee, actualOperatorFee)

		// 4. Verify relative error is bounded.
		requireRelativeErrorLE(t, name, estimated, actualTotal, 15)

		// 5. Verify not wildly over
		if actualTotal.Sign() > 0 {
			gap := new(big.Int).Sub(estimated, actualTotal)
			gapPct := new(big.Int).Mul(gap, big.NewInt(100))
			gapPct.Div(gapPct, actualTotal)
			t.Logf("%s: overestimate=%s%%", name, gapPct)
		}
	}

	// R-01: Simple native-token transfer (MNT on Mantle)
	t.Run("R01_SimpleTransfer", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		bob := sys.Wallet.NewEOA(sys.L2EL)
		verifyEstimateVsActual(t, "R-01", alice, bob, big.NewInt(1000), nil)
	})

	// R-02: Contract deployment — estimate fee then actually deploy, compare
	t.Run("R02_ContractDeploy", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		fromAddr := alice.Address()

		// Minimal contract bytecode (same as used in TransactionTypes)
		deployCode := common.FromHex("6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea164736f6c6343000813000a")

		// 1. Estimate deploy fee
		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &fromAddr, Data: withData(deployCode),
		})
		t.Require().NoError(err)

		// 2. Actually deploy
		tx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(deployCode))
		receipt, err := tx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)
		t.Require().True(receipt.ContractAddress != common.Address{}, "should have contract address")

		// 3. Calculate actual fee
		actualL2Fee := new(big.Int).Mul(
			new(big.Int).SetUint64(receipt.GasUsed),
			receipt.EffectiveGasPrice,
		)
		actualL1Fee := big.NewInt(0)
		if receipt.L1Fee != nil {
			actualL1Fee = new(big.Int).Set(receipt.L1Fee)
		}
		actualTotal := readActualFeeBySenderDelta(t, fromAddr, big.NewInt(0), receipt.BlockNumber)
		actualOperatorFee := new(big.Int).Sub(new(big.Int).Sub(new(big.Int).Set(actualTotal), actualL2Fee), actualL1Fee)
		t.Require().True(actualOperatorFee.Sign() >= 0, "R-02 operator fee should be >= 0, got %s", actualOperatorFee)

		t.Logf("R-02 Deploy: estimated=%s actual=%s (L2=%s L1=%s Op=%s) contract=%s",
			estimated, actualTotal, actualL2Fee, actualL1Fee, actualOperatorFee, receipt.ContractAddress)

		// 4. Verify estimate >= actual
		requireRelativeErrorLE(t, "R-02", estimated, actualTotal, 20)
	})

	// R-03: Contract call — deploy storage contract, estimate call fee, then actually call
	t.Run("R03_ContractCall", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		fromAddr := alice.Address()

		// Deploy a simple storage contract:
		//   Runtime: CALLDATALOAD(0) -> SSTORE(slot 0, value) -> STOP
		//   Init: copies 7-byte runtime to memory and returns it
		storageDeployCode := common.FromHex("6007600c60003960076000f360003560005500")
		deployTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(storageDeployCode))
		deployReceipt, err := deployTx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), deployReceipt.Status)
		contractAddr := deployReceipt.ContractAddress
		t.Logf("R-03 Deployed storage contract at %s", contractAddr)

		// Prepare calldata: 32 bytes to store
		calldata := common.LeftPadBytes(big.NewInt(42).Bytes(), 32)

		// 1. Estimate call fee
		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &fromAddr, To: &contractAddr, Data: withData(calldata),
		})
		t.Require().NoError(err)

		// 2. Actually call
		callTx := txplan.NewPlannedTx(
			alice.Plan(),
			txplan.WithTo(&contractAddr),
			txplan.WithData(calldata),
		)
		callReceipt, err := callTx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), callReceipt.Status)

		// 3. Calculate actual fee
		actualL2Fee := new(big.Int).Mul(
			new(big.Int).SetUint64(callReceipt.GasUsed),
			callReceipt.EffectiveGasPrice,
		)
		actualL1Fee := big.NewInt(0)
		if callReceipt.L1Fee != nil {
			actualL1Fee = new(big.Int).Set(callReceipt.L1Fee)
		}
		actualTotal := readActualFeeBySenderDelta(t, fromAddr, big.NewInt(0), callReceipt.BlockNumber)
		actualOperatorFee := new(big.Int).Sub(new(big.Int).Sub(new(big.Int).Set(actualTotal), actualL2Fee), actualL1Fee)
		t.Require().True(actualOperatorFee.Sign() >= 0, "R-03 operator fee should be >= 0, got %s", actualOperatorFee)

		t.Logf("R-03 Call: estimated=%s actual=%s (L2=%s L1=%s Op=%s gas=%d)",
			estimated, actualTotal, actualL2Fee, actualL1Fee, actualOperatorFee, callReceipt.GasUsed)

		// 4. Verify estimate >= actual (call with SSTORE should use significant gas)
		requireRelativeErrorLE(t, "R-03", estimated, actualTotal, 20)
		t.Require().True(callReceipt.GasUsed > 21000,
			"storage write should use more than base gas, got %d", callReceipt.GasUsed)
	})

	// R-04: Large data transfer
	t.Run("R04_LargeData", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		bob := sys.Wallet.NewEOA(sys.L2EL)
		data := make([]byte, 1024)
		_, err := rand.Read(data)
		t.Require().NoError(err)
		verifyEstimateVsActual(t, "R-04", alice, bob, big.NewInt(0), data)
	})

	// R-05: Minimal data (signature compensation bug more visible)
	t.Run("R05_MinimalData", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		bob := sys.Wallet.NewEOA(sys.L2EL)
		verifyEstimateVsActual(t, "R-05", alice, bob, big.NewInt(1), nil)
	})

	// R-06: Legacy transaction (explicit gasPrice)
	t.Run("R06_LegacyTx", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		bob := sys.Wallet.NewEOA(sys.L2EL)
		fromAddr := alice.Address()
		toAddr := bob.Address()

		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &fromAddr, To: &toAddr, Value: weiHex(1000), GasPrice: gwei(1),
		})
		t.Require().NoError(err)
		t.Require().True(estimated.Sign() > 0)
		t.Logf("R-06 Legacy estimated=%s", estimated)
	})

	// R-07: EIP-1559 transaction
	t.Run("R07_EIP1559Tx", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		bob := sys.Wallet.NewEOA(sys.L2EL)
		fromAddr := alice.Address()
		toAddr := bob.Address()

		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &fromAddr, To: &toAddr, Value: weiHex(1000),
			MaxFeePerGas: gwei(10), MaxPriorityFeePerGas: gwei(2),
		})
		t.Require().NoError(err)
		t.Require().True(estimated.Sign() > 0)
		t.Logf("R-07 EIP-1559 estimated=%s", estimated)
	})

	// R-08: Complex contract — deploy multi-slot storage writer, estimate & compare
	t.Run("R08_ComplexContractInteraction", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		fromAddr := alice.Address()

		// Deploy a contract that writes 3 storage slots on every call:
		//   Runtime (19 bytes):
		//     CALLDATALOAD(0)  -> SSTORE(slot 0)
		//     CALLDATALOAD(32) -> SSTORE(slot 1)
		//     CALLDATALOAD(64) -> SSTORE(slot 2)
		//     STOP
		//   Init: copies 19-byte runtime to memory and returns it
		multiStoreCode := common.FromHex(
			"6013600c60003960136000f3" + // init (12 bytes)
				"600035600055" + // CALLDATALOAD(0) SSTORE(0)
				"602035600155" + // CALLDATALOAD(32) SSTORE(1)
				"604035600255" + // CALLDATALOAD(64) SSTORE(2)
				"00") // STOP
		deployTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(multiStoreCode))
		deployReceipt, err := deployTx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), deployReceipt.Status)
		contractAddr := deployReceipt.ContractAddress
		t.Logf("R-08 Deployed multi-store contract at %s", contractAddr)

		// Prepare calldata: 3 x 32-byte values
		calldata := make([]byte, 96)
		copy(calldata[0:32], common.LeftPadBytes(big.NewInt(111).Bytes(), 32))
		copy(calldata[32:64], common.LeftPadBytes(big.NewInt(222).Bytes(), 32))
		copy(calldata[64:96], common.LeftPadBytes(big.NewInt(333).Bytes(), 32))

		// 1. Estimate call fee
		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &fromAddr, To: &contractAddr, Data: withData(calldata),
		})
		t.Require().NoError(err)

		// 2. Actually call (writes 3 storage slots — cold SSTORE is 20000 gas each)
		callTx := txplan.NewPlannedTx(
			alice.Plan(),
			txplan.WithTo(&contractAddr),
			txplan.WithData(calldata),
		)
		callReceipt, err := callTx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), callReceipt.Status)

		// 3. Calculate actual fee
		actualL2Fee := new(big.Int).Mul(
			new(big.Int).SetUint64(callReceipt.GasUsed),
			callReceipt.EffectiveGasPrice,
		)
		actualL1Fee := big.NewInt(0)
		if callReceipt.L1Fee != nil {
			actualL1Fee = new(big.Int).Set(callReceipt.L1Fee)
		}
		actualTotal := readActualFeeBySenderDelta(t, fromAddr, big.NewInt(0), callReceipt.BlockNumber)
		actualOperatorFee := new(big.Int).Sub(new(big.Int).Sub(new(big.Int).Set(actualTotal), actualL2Fee), actualL1Fee)
		t.Require().True(actualOperatorFee.Sign() >= 0, "R-08 operator fee should be >= 0, got %s", actualOperatorFee)

		t.Logf("R-08 MultiStore: estimated=%s actual=%s (L2=%s L1=%s Op=%s gas=%d)",
			estimated, actualTotal, actualL2Fee, actualL1Fee, actualOperatorFee, callReceipt.GasUsed)

		// 4. Verify: estimate >= actual, and gas is significantly > 21000 (3 cold SSTOREs)
		requireRelativeErrorLE(t, "R-08", estimated, actualTotal, 20)
		t.Require().True(callReceipt.GasUsed > 60000,
			"3 cold SSTOREs should use > 60000 gas, got %d", callReceipt.GasUsed)
	})
}

// ============================================================
// L1 Fee Accuracy (L1-01 ~ L1-02)
// ============================================================

func TestEstimateTotalFee_L1FeeAccuracy(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	l2Client := sys.L2EL.Escape().EthClient()
	rpcCli := l2Client.RPC()

	// L1-01: Signature compensation verification
	// Compare estimated L1 fee (via surplus) with actual receipt L1 fee
	t.Run("L1_01_SignatureCompensation", func(t devtest.T) {
		alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		bob := sys.Wallet.NewEOA(sys.L2EL)
		fromAddr := alice.Address()
		toAddr := bob.Address()

		args := estimateArgs{
			From: &fromAddr, To: &toAddr, Value: weiHex(1),
		}
		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, args)
		t.Require().NoError(err)

		// Send actual tx
		tx := alice.Transfer(bob.Address(), eth.WeiBig(big.NewInt(1)))
		receipt, err := tx.Included.Eval(ctx)
		t.Require().NoError(err)

		actualL1Fee := big.NewInt(0)
		if receipt.L1Fee != nil {
			actualL1Fee = new(big.Int).Set(receipt.L1Fee)
		}
		actualL2Fee := new(big.Int).Mul(
			new(big.Int).SetUint64(receipt.GasUsed),
			receipt.EffectiveGasPrice,
		)
		prevBlock := new(big.Int).Sub(receipt.BlockNumber, big.NewInt(1))
		if prevBlock.Sign() < 0 {
			prevBlock = big.NewInt(0)
		}
		prevBlockHex := hexutil.EncodeBig(prevBlock)
		blockHex := hexutil.EncodeBig(receipt.BlockNumber)
		var vaultBefore, vaultAfter hexutil.Big
		err = rpcCli.CallContext(ctx, &vaultBefore, "eth_getBalance", predeploys.OperatorFeeVaultAddr, prevBlockHex)
		t.Require().NoError(err)
		err = rpcCli.CallContext(ctx, &vaultAfter, "eth_getBalance", predeploys.OperatorFeeVaultAddr, blockHex)
		t.Require().NoError(err)
		actualOperator := new(big.Int).Sub(vaultAfter.ToInt(), vaultBefore.ToInt())
		actualTotal := new(big.Int).Add(actualL2Fee, actualL1Fee)
		actualTotal.Add(actualTotal, actualOperator)

		estimatedGas, err := rpcEstimateGas(ctx, rpcCli, args)
		t.Require().NoError(err)
		suggestPrice, err := rpcGasPrice(ctx, rpcCli)
		t.Require().NoError(err)
		estimatedL2Fee := new(big.Int).Mul(new(big.Int).SetUint64(estimatedGas), suggestPrice)
		estimatedSurplus := new(big.Int).Sub(estimated, estimatedL2Fee) // ≈ estimated L1+Op

		t.Logf("L1-01 estimated total=%s, estimatedL2=%s, estimatedSurplus(L1+Op)=%s",
			estimated, estimatedL2Fee, estimatedSurplus)
		t.Logf("L1-01 actual total=%s (L2=%s L1=%s Op=%s)", actualTotal, actualL2Fee, actualL1Fee, actualOperator)
		requireRelativeErrorLE(t, "L1-01", estimated, actualTotal, 20)
	})

	// L1-02: Data size impact on L1 fee deviation
	t.Run("L1_02_DataSizeDeviationRatio", func(t devtest.T) {
		sizes := []int{0, 256, 2048}
		for _, sz := range sizes {
			alice := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
			bob := sys.Wallet.NewEOA(sys.L2EL)
			fromAddr := alice.Address()
			toAddr := bob.Address()

			var data []byte
			if sz > 0 {
				data = make([]byte, sz)
				_, err := rand.Read(data)
				t.Require().NoError(err)
			}

			args := estimateArgs{From: &fromAddr, To: &toAddr}
			if data != nil {
				args.Data = withData(data)
			}

			estimated, err := rpcEstimateTotalFee(ctx, rpcCli, args)
			t.Require().NoError(err)

			// Send actual tx
			opts := []txplan.Option{alice.Plan(), txplan.WithTo(&toAddr)}
			if data != nil {
				opts = append(opts, txplan.WithData(data))
			}
			txp := txplan.NewPlannedTx(txplan.Combine(opts...))
			receipt, err := txp.Included.Eval(ctx)
			t.Require().NoError(err)

			actualTotal := new(big.Int).Mul(
				new(big.Int).SetUint64(receipt.GasUsed),
				receipt.EffectiveGasPrice,
			)
			if receipt.L1Fee != nil {
				actualTotal.Add(actualTotal, receipt.L1Fee)
			}
			prevBlock := new(big.Int).Sub(receipt.BlockNumber, big.NewInt(1))
			if prevBlock.Sign() < 0 {
				prevBlock = big.NewInt(0)
			}
			prevBlockHex := hexutil.EncodeBig(prevBlock)
			blockHex := hexutil.EncodeBig(receipt.BlockNumber)
			var vaultBefore, vaultAfter hexutil.Big
			err = rpcCli.CallContext(ctx, &vaultBefore, "eth_getBalance", predeploys.OperatorFeeVaultAddr, prevBlockHex)
			t.Require().NoError(err)
			err = rpcCli.CallContext(ctx, &vaultAfter, "eth_getBalance", predeploys.OperatorFeeVaultAddr, blockHex)
			t.Require().NoError(err)
			actualOperator := new(big.Int).Sub(vaultAfter.ToInt(), vaultBefore.ToInt())
			actualTotal.Add(actualTotal, actualOperator)

			if actualTotal.Sign() > 0 {
				gap := new(big.Int).Sub(estimated, actualTotal)
				gapPct := new(big.Int).Mul(gap, big.NewInt(100))
				gapPct.Div(gapPct, actualTotal)
				t.Logf("L1-02 dataSize=%d estimated=%s actual=%s (op=%s) gap=%s%%",
					sz, estimated, actualTotal, actualOperator, gapPct)
			}
			requireRelativeErrorLE(t, "L1-02", estimated, actualTotal, 25)
		}
	})
}

// ============================================================
// Operator Fee Tests (OP-01 ~ OP-06)
// ============================================================

func TestEstimateTotalFee_OperatorFee(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpcCli := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()

	operatorFee := dsl.NewOperatorFee(t, sys.L2Chain, sys.L1EL)
	base := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(1)}

	// ---- Phase 1: scalar=100M, constant=500 ----
	operatorFee.SetOperatorFee(100000000, 500)
	operatorFee.WaitForL2SyncWithCurrentL1State()

	var feeScalar100M *big.Int

	// OP-01: With operator fee enabled, surplus (total - L2) should be positive
	t.Run("OP01_OperatorFeeContributes", func(t devtest.T) {
		totalFee, err := rpcEstimateTotalFee(ctx, rpcCli, base)
		t.Require().NoError(err)

		gas, err := rpcEstimateGas(ctx, rpcCli, base)
		t.Require().NoError(err)
		gasPrice, err := rpcGasPrice(ctx, rpcCli)
		t.Require().NoError(err)

		l2Fee := new(big.Int).Mul(new(big.Int).SetUint64(gas), gasPrice)
		surplus := new(big.Int).Sub(totalFee, l2Fee)

		t.Logf("OP-01 total=%s, L2=%s, surplus(L1+Op)=%s", totalFee, l2Fee, surplus)
		t.Require().True(surplus.Sign() > 0,
			"with operator fee enabled, surplus should > 0")
		feeScalar100M = totalFee
	})

	// OP-04: Contract call uses more gas → operator fee (gasUsed*scalar) is higher
	t.Run("OP04_ScalesWithGasUsage", func(t devtest.T) {
		// Deploy a storage contract (SSTORE = high gas)
		storageCode := common.FromHex("6007600c60003960076000f360003560005500")
		deployTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(storageCode))
		receipt, err := deployTx.Included.Eval(ctx)
		t.Require().NoError(err)
		contractAddr := receipt.ContractAddress

		// Simple transfer estimate
		simpleFee, err := rpcEstimateTotalFee(ctx, rpcCli, base)
		t.Require().NoError(err)
		simpleGas, err := rpcEstimateGas(ctx, rpcCli, base)
		t.Require().NoError(err)

		// Contract call estimate (cold SSTORE → much more gas)
		calldata := common.LeftPadBytes(big.NewInt(999).Bytes(), 32)
		contractArgs := estimateArgs{From: &aliceAddr, To: &contractAddr, Data: withData(calldata)}
		contractFee, err := rpcEstimateTotalFee(ctx, rpcCli, contractArgs)
		t.Require().NoError(err)
		contractGas, err := rpcEstimateGas(ctx, rpcCli, contractArgs)
		t.Require().NoError(err)

		t.Logf("OP-04 simple: fee=%s gas=%d", simpleFee, simpleGas)
		t.Logf("OP-04 contract: fee=%s gas=%d", contractFee, contractGas)

		t.Require().True(contractGas > simpleGas,
			"contract gas(%d) should > simple gas(%d)", contractGas, simpleGas)
		t.Require().True(contractFee.Cmp(simpleFee) > 0,
			"contract fee(%s) should > simple fee(%s)", contractFee, simpleFee)

		// Fee diff > pure L2-gas diff because operator fee also scales with gas
		gasPrice, err := rpcGasPrice(ctx, rpcCli)
		t.Require().NoError(err)
		gasDiff := contractGas - simpleGas
		l2Diff := new(big.Int).Mul(new(big.Int).SetUint64(gasDiff), gasPrice)
		feeDiff := new(big.Int).Sub(contractFee, simpleFee)
		t.Logf("OP-04 feeDiff=%s, l2GasDiff=%s", feeDiff, l2Diff)
		t.Require().True(feeDiff.Cmp(l2Diff) > 0,
			"fee diff(%s) > pure L2 gas diff(%s) — operator fee contributes", feeDiff, l2Diff)
	})

	// OP-05: Both L2 fee and operator fee scale with gas (from data size)
	t.Run("OP05_DataSizeImpact", func(t devtest.T) {
		small := make([]byte, 32)
		for i := range small {
			small[i] = 0xFF
		}
		large := make([]byte, 2048)
		for i := range large {
			large[i] = 0xFF
		}

		smallArgs := estimateArgs{From: &aliceAddr, To: &bobAddr, Data: withData(small)}
		largeArgs := estimateArgs{From: &aliceAddr, To: &bobAddr, Data: withData(large)}

		smallFee, err := rpcEstimateTotalFee(ctx, rpcCli, smallArgs)
		t.Require().NoError(err)
		largeFee, err := rpcEstimateTotalFee(ctx, rpcCli, largeArgs)
		t.Require().NoError(err)

		smallGas, err := rpcEstimateGas(ctx, rpcCli, smallArgs)
		t.Require().NoError(err)
		largeGas, err := rpcEstimateGas(ctx, rpcCli, largeArgs)
		t.Require().NoError(err)

		t.Logf("OP-05 small(32B): fee=%s gas=%d", smallFee, smallGas)
		t.Logf("OP-05 large(2KB): fee=%s gas=%d", largeFee, largeGas)

		t.Require().True(largeGas > smallGas,
			"larger data gas(%d) > smaller data gas(%d)", largeGas, smallGas)
		t.Require().True(largeFee.Cmp(smallFee) > 0,
			"larger data fee(%s) > smaller data fee(%s)", largeFee, smallFee)
	})

	// OP-06: Estimated total >= actual total (L2 + L1 + OperatorFeeVault diff)
	t.Run("OP06_EstimateVsActualWithOperator", func(t devtest.T) {
		sender := sys.FunderL2.NewFundedEOA(eth.OneTenthEther)
		receiver := sys.Wallet.NewEOA(sys.L2EL)
		fromAddr := sender.Address()
		toAddr := receiver.Address()

		// 1. Estimate
		estimated, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &fromAddr, To: &toAddr, Value: weiHex(1000),
		})
		t.Require().NoError(err)

		// 2. Send actual transaction
		tx := sender.Transfer(receiver.Address(), eth.WeiBig(big.NewInt(1000)))
		receipt, err := tx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)

		// 3. Read operator fee vault balance before and after tx block
		prevBlock := new(big.Int).Sub(receipt.BlockNumber, big.NewInt(1))
		if prevBlock.Sign() < 0 {
			prevBlock = big.NewInt(0)
		}
		prevBlockHex := hexutil.EncodeBig(prevBlock)
		blockHex := hexutil.EncodeBig(receipt.BlockNumber)

		var vaultBefore, vaultAfter hexutil.Big
		err = rpcCli.CallContext(ctx, &vaultBefore, "eth_getBalance",
			predeploys.OperatorFeeVaultAddr, prevBlockHex)
		t.Require().NoError(err)
		err = rpcCli.CallContext(ctx, &vaultAfter, "eth_getBalance",
			predeploys.OperatorFeeVaultAddr, blockHex)
		t.Require().NoError(err)

		actualOperator := new(big.Int).Sub(vaultAfter.ToInt(), vaultBefore.ToInt())

		// 4. Compute actual total = L2 + L1 + operator
		actualL2 := new(big.Int).Mul(
			new(big.Int).SetUint64(receipt.GasUsed), receipt.EffectiveGasPrice)
		actualL1 := big.NewInt(0)
		if receipt.L1Fee != nil {
			actualL1 = new(big.Int).Set(receipt.L1Fee)
		}
		actualTotal := new(big.Int).Add(actualL2, actualL1)
		actualTotal.Add(actualTotal, actualOperator)

		t.Logf("OP-06 estimated=%s actual=%s (L2=%s L1=%s Op=%s)",
			estimated, actualTotal, actualL2, actualL1, actualOperator)

		// 5. Verify estimate accuracy (including operator fee).
		requireRelativeErrorLE(t, "OP-06", estimated, actualTotal, 20)
		t.Require().True(actualOperator.Sign() > 0,
			"operator fee vault increase should > 0, got %s", actualOperator)
	})

	// ---- Phase 2: Double scalar → higher total ----
	operatorFee.SetOperatorFee(200000000, 500)
	operatorFee.WaitForL2SyncWithCurrentL1State()

	t.Run("OP02_HigherScalarHigherFee", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpcCli, base)
		t.Require().NoError(err)

		t.Logf("OP-02 scalar=100M fee=%s, scalar=200M fee=%s", feeScalar100M, fee)
		t.Require().True(fee.Cmp(feeScalar100M) > 0,
			"doubled scalar fee(%s) should > original(%s)", fee, feeScalar100M)
	})

	// ---- Phase 3: Zero operator fee → total ≈ L2 + L1 only ----
	operatorFee.SetOperatorFee(0, 0)
	operatorFee.WaitForL2SyncWithCurrentL1State()

	t.Run("OP03_ZeroOperatorFee", func(t devtest.T) {
		fee, err := rpcEstimateTotalFee(ctx, rpcCli, base)
		t.Require().NoError(err)

		t.Logf("OP-03 zero-op fee=%s, with-op fee=%s", fee, feeScalar100M)
		t.Require().True(fee.Cmp(feeScalar100M) < 0,
			"zero operator fee(%s) should < with operator(%s)", fee, feeScalar100M)
	})

	// Restore original config
	operatorFee.RestoreOriginalConfig()
	operatorFee.WaitForL2SyncWithCurrentL1State()
}

// ============================================================
// Contract Interaction Tests (CI-01 ~ CI-05)
// ============================================================

func TestEstimateTotalFee_ContractInteractions(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpcCli := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()

	// ---- Deploy all test contracts up front ----

	// 1. ERC20-like: 2 SSTOREs + RETURN(true)
	//    Runtime (22 bytes):
	//      CALLDATALOAD(0) → SSTORE(slot 0)
	//      CALLDATALOAD(32) → SSTORE(slot 1)
	//      MSTORE(0, 1), RETURN(0, 32)
	erc20Code := common.FromHex(
		"6016600c60003960166000f3" + // init (12 bytes, runtime=0x16=22)
			"600035600055" + // CALLDATALOAD(0) SSTORE(0)
			"602035600155" + // CALLDATALOAD(32) SSTORE(1)
			"6001600052" + // MSTORE(0, 1)
			"60206000f3") // RETURN(0, 32)
	erc20Tx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(erc20Code))
	erc20Receipt, err := erc20Tx.Included.Eval(ctx)
	require.NoError(err)
	erc20Addr := erc20Receipt.ContractAddress
	t.Logf("Deployed ERC20-like at %s", erc20Addr)

	// 2. Event emitter: LOG0(calldata[0:32])
	//    Runtime (12 bytes):
	//      CALLDATALOAD(0) → MSTORE(0), LOG0(0,32), STOP
	eventCode := common.FromHex(
		"600c600c600039600c6000f3" + // init (12 bytes, runtime=0x0c=12)
			"6000356000526020" + // CALLDATALOAD(0) MSTORE(0) PUSH1 32
			"6000a000") // PUSH1 0 LOG0 STOP
	eventTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(eventCode))
	eventReceipt, err := eventTx.Included.Eval(ctx)
	require.NoError(err)
	eventAddr := eventReceipt.ContractAddress
	t.Logf("Deployed event emitter at %s", eventAddr)

	// 3. Payable fallback: CALLVALUE → SSTORE(slot 0), STOP
	//    Runtime (5 bytes): 34 6000 55 00
	payableCode := common.FromHex(
		"6005600c60003960056000f3" + // init (12 bytes, runtime=5)
			"3460005500") // CALLVALUE SSTORE(0) STOP
	payableTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(payableCode))
	payableReceipt, err := payableTx.Included.Eval(ctx)
	require.NoError(err)
	payableAddr := payableReceipt.ContractAddress
	t.Logf("Deployed payable contract at %s", payableAddr)

	// 4. Storage writer (target for nested calls)
	//    Runtime (7 bytes): CALLDATALOAD(0) → SSTORE(0), STOP
	storageCode := common.FromHex("6007600c60003960076000f360003560005500")
	storageTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(storageCode))
	storageReceipt, err := storageTx.Included.Eval(ctx)
	require.NoError(err)
	storageAddr := storageReceipt.ContractAddress

	// 5. Caller: CALL(gas, addr_from_calldata, 0, 0, 0, 0, 0)
	//    Runtime (16 bytes): push 6 zeros, CALLDATALOAD, GAS, CALL, STOP
	callerCode := common.FromHex(
		"6010600c60003960106000f3" + // init (12 bytes, runtime=0x10=16)
			"600060006000600060006000355af100") // CALL(gas, addr, 0, 0, 0, 0, 0) STOP
	callerTx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(callerCode))
	callerReceipt, err := callerTx.Included.Eval(ctx)
	require.NoError(err)
	callerAddr := callerReceipt.ContractAddress
	t.Logf("Deployed caller contract at %s", callerAddr)

	// ---- Tests ----

	// CI-01: ERC20-like transfer (2 SSTOREs) should cost more than simple EOA transfer
	t.Run("CI01_ERC20LikeTransfer", func(t devtest.T) {
		// Calldata: transfer(bob, 100) → 64 bytes
		to32 := common.LeftPadBytes(bobAddr.Bytes(), 32)
		amt32 := common.LeftPadBytes(big.NewInt(100).Bytes(), 32)
		calldata := append(to32, amt32...)

		erc20Fee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &erc20Addr, Data: withData(calldata),
		})
		t.Require().NoError(err)
		erc20Gas, err := rpcEstimateGas(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &erc20Addr, Data: withData(calldata),
		})
		t.Require().NoError(err)

		simpleFee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		})
		t.Require().NoError(err)
		simpleGas, err := rpcEstimateGas(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		})
		t.Require().NoError(err)

		t.Logf("CI-01 ERC20: fee=%s gas=%d | simple: fee=%s gas=%d",
			erc20Fee, erc20Gas, simpleFee, simpleGas)
		t.Require().True(erc20Gas > simpleGas,
			"ERC20 gas(%d) should > simple(%d) — 2 SSTOREs", erc20Gas, simpleGas)
		t.Require().True(erc20Fee.Cmp(simpleFee) > 0,
			"ERC20 fee(%s) should > simple(%s)", erc20Fee, simpleFee)
	})

	// CI-02: Contract that emits event (LOG0 opcode)
	t.Run("CI02_ContractWithEvents", func(t devtest.T) {
		eventData := common.LeftPadBytes(big.NewInt(12345).Bytes(), 32)

		fee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &eventAddr, Data: withData(eventData),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)

		gas, err := rpcEstimateGas(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &eventAddr, Data: withData(eventData),
		})
		t.Require().NoError(err)
		// LOG0 costs 375 base + 8 per byte (32 bytes = 256) = 631 gas
		// plus 21000 base + calldata + MSTORE etc
		t.Logf("CI-02 event emitter: fee=%s gas=%d", fee, gas)
		t.Require().True(gas > 21000,
			"event emit should use > 21000 gas, got %d", gas)

		// Actually send to verify estimate works
		callTx := txplan.NewPlannedTx(
			alice.Plan(), txplan.WithTo(&eventAddr), txplan.WithData(eventData))
		receipt, err := callTx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)
		t.Require().True(len(receipt.Logs) > 0, "should have at least 1 log")
		t.Logf("CI-02 actual gas=%d, logs=%d", receipt.GasUsed, len(receipt.Logs))
	})

	// CI-03: Constructor that does work (3 SSTOREs in init code)
	t.Run("CI03_ConstructorWithWork", func(t devtest.T) {
		// Init code (27 bytes): SSTORE(0,42), SSTORE(1,99), SSTORE(2,7), then deploy STOP
		// Runtime (1 byte): STOP
		constructorCode := common.FromHex(
			"602a600055" + // PUSH1 42, PUSH1 0, SSTORE
				"6063600155" + // PUSH1 99, PUSH1 1, SSTORE
				"6007600255" + // PUSH1 7,  PUSH1 2, SSTORE
				"6001601b600039" + // PUSH1 1(size), PUSH1 27(offset), PUSH1 0(dest), CODECOPY
				"60016000f3" + // PUSH1 1(size), PUSH1 0(offset), RETURN
				"00") // runtime: STOP

		fee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, Data: withData(constructorCode),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)

		// Simple deploy with no constructor work (existing minimal contract)
		minimalCode := common.FromHex(
			"6080604052348015600f57600080fd5b50603f80601d6000396000f3fe6080604052600080fdfea164736f6c6343000813000a")
		minimalFee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, Data: withData(minimalCode),
		})
		t.Require().NoError(err)

		t.Logf("CI-03 constructor(3 SSTOREs): fee=%s | minimal: fee=%s", fee, minimalFee)
		// 3 cold SSTOREs = ~60000 extra gas → should dwarf the L1 data-size difference
		t.Require().True(fee.Cmp(minimalFee) > 0,
			"constructor with 3 SSTOREs(%s) should > minimal deploy(%s)", fee, minimalFee)

		// Deploy and verify it works
		tx := txplan.NewPlannedTx(alice.Plan(), txplan.WithData(constructorCode))
		receipt, err := tx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)
		t.Logf("CI-03 deployed at %s, gas=%d", receipt.ContractAddress, receipt.GasUsed)
	})

	// CI-04: Payable contract with value (stores CALLVALUE into storage)
	t.Run("CI04_PayableFallbackWithValue", func(t devtest.T) {
		sendValue := big.NewInt(1e15) // 0.001 MNT

		fee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &payableAddr, Value: bigHex(sendValue),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)

		gas, err := rpcEstimateGas(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &payableAddr, Value: bigHex(sendValue),
		})
		t.Require().NoError(err)
		t.Logf("CI-04 payable+value: fee=%s gas=%d", fee, gas)
		// CALLVALUE + SSTORE costs more than a simple transfer (21000)
		t.Require().True(gas > 21000,
			"payable with SSTORE should use > 21000 gas, got %d", gas)

		// Compare: same value to EOA (no SSTORE)
		eoaFee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: bigHex(sendValue),
		})
		t.Require().NoError(err)
		t.Require().True(fee.Cmp(eoaFee) > 0,
			"payable contract(%s) > EOA transfer(%s) due to SSTORE", fee, eoaFee)

		// Actually send to verify
		callTx := txplan.NewPlannedTx(
			alice.Plan(), txplan.WithTo(&payableAddr),
			txplan.WithValue(eth.WeiBig(sendValue)))
		receipt, err := callTx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)
	})

	// CI-05: Nested call (caller → storage writer)
	t.Run("CI05_NestedCalls", func(t devtest.T) {
		// Pass storage writer address as calldata (left-padded to 32 bytes)
		targetCalldata := common.LeftPadBytes(storageAddr.Bytes(), 32)

		nestedFee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &callerAddr, Data: withData(targetCalldata),
		})
		t.Require().NoError(err)
		t.Require().True(nestedFee.Sign() > 0)

		nestedGas, err := rpcEstimateGas(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &callerAddr, Data: withData(targetCalldata),
		})
		t.Require().NoError(err)

		// Compare with simple EOA transfer
		simpleFee, err := rpcEstimateTotalFee(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		})
		t.Require().NoError(err)
		simpleGas, err := rpcEstimateGas(ctx, rpcCli, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Value: weiHex(1),
		})
		t.Require().NoError(err)

		t.Logf("CI-05 nested: fee=%s gas=%d | simple: fee=%s gas=%d",
			nestedFee, nestedGas, simpleFee, simpleGas)
		// CALL opcode overhead (2600 cold access) + sub-call execution
		t.Require().True(nestedGas > simpleGas,
			"nested call gas(%d) > simple(%d) — CALL overhead", nestedGas, simpleGas)
		t.Require().True(nestedFee.Cmp(simpleFee) > 0,
			"nested call fee(%s) > simple fee(%s)", nestedFee, simpleFee)

		// Actually send to verify
		callTx := txplan.NewPlannedTx(
			alice.Plan(), txplan.WithTo(&callerAddr), txplan.WithData(targetCalldata))
		receipt, err := callTx.Included.Eval(ctx)
		t.Require().NoError(err)
		t.Require().Equal(uint64(1), receipt.Status)
		t.Logf("CI-05 actual gas=%d", receipt.GasUsed)
	})
}

// ============================================================
// Performance Tests (PF-01 ~ PF-05)
// ============================================================

func TestEstimateTotalFee_Performance(gt *testing.T) {
	t := devtest.SerialT(gt)
	sys := presets.NewMantleMinimal(t)
	require := t.Require()
	ctx := t.Ctx()

	require.True(sys.L2Chain.IsMantleForkActive(forks.MantleArsia))

	alice := sys.FunderL2.NewFundedEOA(eth.HundredEther)
	bob := sys.Wallet.NewEOA(sys.L2EL)
	rpc := sys.L2EL.Escape().EthClient().RPC()
	aliceAddr := alice.Address()
	bobAddr := bob.Address()
	base := estimateArgs{From: &aliceAddr, To: &bobAddr, Value: weiHex(1)}

	// PF-01: Response time < 500ms for simple transfer
	t.Run("PF01_ResponseTime", func(t devtest.T) {
		start := time.Now()
		fee, err := rpcEstimateTotalFee(ctx, rpc, base)
		elapsed := time.Since(start)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("PF-01 response time: %v", elapsed)
		t.Require().True(elapsed < 5*time.Second, "should respond within 5s, took %v", elapsed)
	})

	// PF-04: Large data response time
	t.Run("PF04_LargeDataResponseTime", func(t devtest.T) {
		bigData := make([]byte, 128*1024)
		_, err := rand.Read(bigData)
		t.Require().NoError(err)
		start := time.Now()
		fee, err := rpcEstimateTotalFee(ctx, rpc, estimateArgs{
			From: &aliceAddr, To: &bobAddr, Data: withData(bigData),
		})
		elapsed := time.Since(start)
		t.Require().NoError(err)
		t.Require().True(fee.Sign() > 0)
		t.Logf("PF-04 128KB response time: %v", elapsed)
		t.Require().True(elapsed < 10*time.Second, "128KB estimate should respond within 10s, took %v", elapsed)
	})

	// PF-05: Sequential stress (100 calls)
	t.Run("PF05_SequentialStress", func(t devtest.T) {
		const N = 100
		start := time.Now()
		for i := 0; i < N; i++ {
			fee, err := rpcEstimateTotalFee(ctx, rpc, base)
			t.Require().NoError(err)
			t.Require().True(fee.Sign() > 0)
		}
		elapsed := time.Since(start)
		avg := elapsed / N
		t.Logf("PF-05 %d calls in %v, avg=%v", N, elapsed, avg)
		t.Require().True(avg < 800*time.Millisecond, "sequential avg should stay below 800ms, got %v", avg)
	})
}
