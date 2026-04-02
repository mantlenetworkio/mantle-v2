package mantleopreth

// Engine-API tests for op-reth with Mantle genesis.
//
// These tests exercise the OP Stack Engine API against an external op-reth
// subprocess initialised with Mantle-specific genesis and rollup config.
// They complement the mantleopgeth suite which tests the same behaviours
// via an in-process geth node.
//
// Scope note:
//   Mantle op-reth supports the OP Stack from the Skadi fork onwards.
//   Pre-Skadi fork tests (Regolith, PreCanyon, PreEcotone, PreFjord, …)
//   are not applicable to reth and are skipped here.  The corresponding
//   tests in mantleopgeth remain the authoritative coverage for those forks.
//   Post-Skadi tests (Canyon, Ecotone, Fjord, Isthmus, …) run against reth
//   with Skadi/Limb/Arsia activated at genesis alongside the tested fork.
//   reth v1.9.3+ activates all Mantle forks at @0 regardless of the genesis
//   JSON; setting the time offsets in the deploy config ensures that the
//   Mantle genesis construction in NewRethEngine produces a consistent chain
//   config for the Engine API client.
//
// Required env vars:
//
//	OP_E2E_L2_BIN=/path/to/op-reth   – path to the op-reth binary
//
// Optional:
//
//	OP_E2E_L2_READY_TIMEOUT=60s      – how long to wait for reth to be ready

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"

	op_e2e "github.com/ethereum-optimism/optimism/op-e2e"
	"github.com/ethereum-optimism/optimism/op-e2e/enginetest"
	"github.com/ethereum-optimism/optimism/op-e2e/system/e2esys"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

var (
	// rip7212Precompile is the secp256r1 signature-verification precompile (RIP-7212).
	// Activated by the Fjord fork (Mantle equivalent: Arsia).
	rip7212Precompile = common.HexToAddress("0x0000000000000000000000000000000000000100")

	// invalid7212Data is a minimal invalid input for the RIP-7212 precompile.
	invalid7212Data = []byte{0x00}

	// valid7212Data is a valid (hash, r, s, x, y) encoding for RIP-7212.
	// Taken from https://gist.github.com/ulerdogan/8f1714895e23a54147fc529ea30517eb
	valid7212Data = common.FromHex("4cee90eb86eaa050036147a12d49004b6b9c72bd725d39d4785011fe190f0b4da73bd4903f0ce3b639bbbf6e8e80d16931ff4bcf5993d58468e8fb19086e8cac36dbcd03009df8c59286b162af3bd7fcc0450c9aa81be5d10d312af6c66b1d604aebd3099c618202fcfe16ae7770b0c49ab5eadf74b754204a3bb6060e44eff37618b065f9832de4ca6ca971a7a1adc826d0f7c00181a5fb2ddf79ae00b4e10e")
)

// TestMissingGasLimit verifies that op-reth (with Mantle genesis) cannot build
// a block without a gas limit while OP Stack is active.
func TestMissingGasLimit(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opReth, err := NewRethEngine(t, ctx, &cfg)
	require.NoError(t, err)
	defer opReth.Close()

	attrs, err := opReth.CreatePayloadAttributes()
	require.NoError(t, err)
	attrs.GasLimit = nil

	res, err := opReth.StartBlockBuilding(ctx, attrs)
	require.Error(t, err)
	// reth returns standard JSON-RPC -32602 (Invalid params) for missing gasLimit;
	// geth returns the Engine API specific -38003 (InvalidPayloadAttributes).
	// Both are correct rejections — verify the call was rejected and returned no payload ID.
	var rpcErr rpc.Error
	require.ErrorAs(t, err, &rpcErr, "error should be an RPC error")
	code := rpcErr.ErrorCode()
	require.True(t,
		code == int(eth.InvalidPayloadAttributes) || code == -32602,
		"expected InvalidPayloadAttributes (-38003) or Invalid params (-32602), got: %d", code,
	)
	require.Nil(t, res)
}

// TestInvalidDepositInFCU verifies that an invalid deposit is still included in
// the block (deposits must never prevent block production).
func TestInvalidDepositInFCU(t *testing.T) {
	op_e2e.InitParallel(t)
	cfg := e2esys.DefaultSystemConfig(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	opReth, err := NewRethEngine(t, ctx, &cfg)
	require.NoError(t, err)
	defer opReth.Close()

	fromKey, err := crypto.GenerateKey()
	require.NoError(t, err)
	fromAddr := crypto.PubkeyToAddress(fromKey.PublicKey)
	balance, err := opReth.L2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)
	require.Equal(t, 0, balance.Cmp(common.Big0))

	badDepositTx := types.NewTx(&types.DepositTx{
		From:                fromAddr,
		To:                  &fromAddr,
		Value:               big.NewInt(params.Ether),
		Gas:                 25000,
		IsSystemTransaction: false,
	})

	_, err = opReth.AddL2Block(ctx, badDepositTx)
	require.NoError(t, err)

	balance, err = opReth.L2Client.BalanceAt(ctx, fromAddr, nil)
	require.Nil(t, err)
	require.Equal(t, 0, balance.Cmp(common.Big0))
}

// TestPreregolith is skipped for op-reth.
//
// Mantle op-reth supports the OP Stack from the Skadi fork onwards and does
// not implement pre-Regolith gas-accounting semantics.  The authoritative
// coverage for pre-Regolith behaviour lives in mantleopgeth.
func TestPreregolith(t *testing.T) {
	t.Skip("pre-Skadi: Mantle op-reth only supports forks from Skadi onwards; " +
		"pre-Regolith behaviour is covered by mantleopgeth")
}

// TestRegolith is skipped for op-reth.
//
// Mantle op-reth supports the OP Stack from the Skadi fork onwards and does
// not implement Regolith-specific gas-accounting semantics in isolation.
// The authoritative coverage for Regolith behaviour lives in mantleopgeth.
func TestRegolith(t *testing.T) {
	t.Skip("pre-Skadi: Mantle op-reth only supports forks from Skadi onwards; " +
		"Regolith behaviour is covered by mantleopgeth")
}

// TestGethOnlyPendingBlockIsLatest is skipped for op-reth.
//
// This test exercises geth-specific pending-block semantics (pending == latest
// when no pending transactions exist).  op-reth exposes the same behaviour at
// the Engine API level but its public RPC does not support eth_getBlockByNumber
// with "pending" in the same way.  Covered by mantleopgeth.
func TestGethOnlyPendingBlockIsLatest(t *testing.T) {
	t.Skip("geth-specific: pending block semantics differ in reth; covered by mantleopgeth")
}

// TestPreCanyon is skipped for op-reth.
//
// The test configures a genesis with Skadi scheduled in the future so that the
// node starts in a pre-Canyon state.  Mantle op-reth always activates Skadi at
// genesis, making the pre-Canyon configuration meaningless for reth.
// Covered by mantleopgeth.
func TestPreCanyon(t *testing.T) {
	t.Skip("pre-Skadi: Mantle op-reth only supports forks from Skadi onwards; " +
		"pre-Canyon behaviour is covered by mantleopgeth")
}

// TestCanyon verifies Canyon-era behaviour on op-reth with Mantle genesis.
//
// Canyon introduced:
//   - non-nil (empty) withdrawals field in execution payloads
//   - PUSH0 opcode (EIP-3855)
//
// In Mantle, Canyon features are activated by Arsia.  The test sets
// Skadi/Limb/Arsia time offsets alongside the Canyon time so that the Mantle
// genesis is well-formed for NewRethEngine.
//
// reth v1.9.3+ activates all forks at genesis @0 regardless of genesis JSON
// time offsets, so only the ActivateAtGenesis configuration is exercised.
func TestCanyon(t *testing.T) {
	canyonTime := hexutil.Uint64(0)

	run := func(name string, body func(t *testing.T, opReth *enginetest.OpEngine, cfg e2esys.SystemConfig)) {
		t.Run(name, func(t *testing.T) {
			op_e2e.InitParallel(t)
			cfg := e2esys.CanyonSystemConfig(t, &canyonTime)
			// Align Mantle forks with the Canyon activation time so that
			// NewRethEngine builds a valid Mantle genesis.
			cfg.DeployConfig.L2GenesisMantleSkadiTimeOffset = &canyonTime
			cfg.DeployConfig.L2GenesisMantleLimbTimeOffset = &canyonTime
			cfg.DeployConfig.L2GenesisMantleArsiaTimeOffset = &canyonTime

			// outer ctx: scopes the reth node lifetime (startup + shutdown).
			// Each body creates its own ctx for individual RPC calls so that
			// per-call timeouts are independent of node startup time.
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			opReth, err := NewRethEngine(t, ctx, &cfg)
			require.NoError(t, err)
			defer opReth.Close()

			body(t, opReth, cfg)
		})
	}

	run("ReturnsEmptyWithdrawals", func(t *testing.T, opReth *enginetest.OpEngine, _ e2esys.SystemConfig) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		b, err := opReth.AddL2Block(ctx)
		require.NoError(t, err)
		// Canyon: withdrawals field is present and empty (not nil).
		assert.Equal(t, *b.ExecutionPayload.Withdrawals, types.Withdrawals{})

		l2Block, err := opReth.L2Client.BlockByNumber(ctx, nil)
		require.Nil(t, err)
		assert.Equal(t, l2Block.Withdrawals(), types.Withdrawals{})
	})

	run("AcceptsPushZeroTxn", func(t *testing.T, opReth *enginetest.OpEngine, cfg e2esys.SystemConfig) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		pushZeroContractCreateTxn := types.NewTx(&types.DepositTx{
			From:                cfg.Secrets.Addresses().Alice,
			Value:               big.NewInt(params.Ether),
			Gas:                 1000001,
			Data:                []byte{byte(vm.PUSH0)},
			IsSystemTransaction: false,
		})

		_, err := opReth.AddL2Block(ctx, pushZeroContractCreateTxn)
		require.NoError(t, err)

		receipt, err := opReth.L2Client.TransactionReceipt(ctx, pushZeroContractCreateTxn.Hash())
		require.NoError(t, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	})
}

// TestPreEcotone is skipped for op-reth.
//
// The test configures a genesis with Skadi in the future (pre-Skadi state).
// Mantle op-reth always activates Skadi at genesis.  Covered by mantleopgeth.
func TestPreEcotone(t *testing.T) {
	t.Skip("pre-Skadi: Mantle op-reth only supports forks from Skadi onwards; " +
		"pre-Ecotone behaviour is covered by mantleopgeth")
}

// TestEcotone verifies Ecotone-era behaviour on op-reth with Mantle genesis.
//
// Ecotone introduced:
//   - non-nil ParentBeaconBlockRoot in execution payloads (EIP-4788)
//   - TSTORE / TLOAD opcodes (EIP-1153)
//
// In Mantle, Ecotone features are activated by Arsia.  The test sets
// Skadi/Limb/Arsia time offsets alongside the Ecotone time.
//
// reth v1.9.3+ activates all forks at genesis @0 regardless of genesis JSON
// time offsets, so only the ActivateAtGenesis configuration is exercised.
func TestEcotone(t *testing.T) {
	ecotoneTime := hexutil.Uint64(0)

	run := func(name string, body func(t *testing.T, opReth *enginetest.OpEngine, cfg e2esys.SystemConfig)) {
		t.Run(name, func(t *testing.T) {
			op_e2e.InitParallel(t)
			cfg := e2esys.EcotoneSystemConfig(t, &ecotoneTime)
			// Align Mantle forks with the Ecotone activation time so that
			// NewRethEngine builds a valid Mantle genesis.
			cfg.DeployConfig.L2GenesisMantleSkadiTimeOffset = &ecotoneTime
			cfg.DeployConfig.L2GenesisMantleLimbTimeOffset = &ecotoneTime
			cfg.DeployConfig.L2GenesisMantleArsiaTimeOffset = &ecotoneTime

			// outer ctx: scopes the reth node lifetime (startup + shutdown).
			// Each body creates its own ctx for individual RPC calls so that
			// per-call timeouts are independent of node startup time.
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			opReth, err := NewRethEngine(t, ctx, &cfg)
			require.NoError(t, err)
			defer opReth.Close()

			body(t, opReth, cfg)
		})
	}

	run("HashParentBeaconBlockRoot", func(t *testing.T, opReth *enginetest.OpEngine, _ e2esys.SystemConfig) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		b, err := opReth.AddL2Block(ctx)
		require.NoError(t, err)
		// Ecotone: ParentBeaconBlockRoot must be present and equal the L1 head's.
		require.NotNil(t, b.ParentBeaconBlockRoot)
		assert.Equal(t, b.ParentBeaconBlockRoot, opReth.L1Head.ParentBeaconRoot())

		l2Block, err := opReth.L2Client.BlockByNumber(ctx, nil)
		require.NoError(t, err)
		assert.NotNil(t, l2Block.Header().ParentBeaconRoot)
		assert.Equal(t, l2Block.Header().ParentBeaconRoot, opReth.L1Head.ParentBeaconRoot())
	})

	run("TstoreTxn", func(t *testing.T, opReth *enginetest.OpEngine, cfg e2esys.SystemConfig) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		tstoreTxn := types.NewTx(&types.DepositTx{
			From:  cfg.Secrets.Addresses().Alice,
			Value: big.NewInt(params.Ether),
			Gas:   1000001,
			Data: []byte{
				byte(vm.PUSH1), 0x01,
				byte(vm.PUSH1), 0x01,
				byte(vm.TSTORE),
				byte(vm.PUSH0),
			},
			IsSystemTransaction: false,
		})

		_, err := opReth.AddL2Block(ctx, tstoreTxn)
		require.NoError(t, err)

		_, err = opReth.AddL2Block(ctx, tstoreTxn)
		require.NoError(t, err)

		receipt, err := opReth.L2Client.TransactionReceipt(ctx, tstoreTxn.Hash())
		require.NoError(t, err)
		assert.Equal(t, types.ReceiptStatusSuccessful, receipt.Status)
	})
}

// TestPreFjord is skipped for op-reth.
//
// The FjordNotScheduled sub-test sets SkadiTimeOffset=nil and
// FjordNotYetActive sets it to a future time — both produce a pre-Skadi
// genesis that Mantle op-reth cannot replicate.  Covered by mantleopgeth.
func TestPreFjord(t *testing.T) {
	t.Skip("pre-Skadi: Mantle op-reth only supports forks from Skadi onwards; " +
		"pre-Fjord behaviour is covered by mantleopgeth")
}

// TestFjord verifies the RIP-7212 secp256r1 precompile on op-reth with Mantle
// genesis.  RIP-7212 is introduced by the Fjord fork; in Mantle it is
// activated by Arsia.  The test aligns Skadi/Limb/Arsia with the Fjord time.
func TestFjord(t *testing.T) {
	// reth v1.9.3 activates ALL forks at genesis @0; a future-fork genesis
	// (fjordTime > 0) causes a genesis-hash mismatch and forkChoiceUpdated
	// returns SYNCING.  Only the ActivateAtGenesis case is meaningful here.
	fjordTime := hexutil.Uint64(0)
	t.Run("RIP7212_ActivateAtGenesis", func(t *testing.T) {
		op_e2e.InitParallel(t)
		cfg := e2esys.FjordSystemConfig(t, &fjordTime)
		// reth requires Mantle forks aligned with Fjord activation:
		// Arsia is the Mantle equivalent of Fjord; Skadi and Limb must precede it.
		cfg.DeployConfig.L2GenesisMantleSkadiTimeOffset = &fjordTime
		cfg.DeployConfig.L2GenesisMantleLimbTimeOffset = &fjordTime
		cfg.DeployConfig.L2GenesisMantleArsiaTimeOffset = &fjordTime

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		opReth, err := NewRethEngine(t, ctx, &cfg)
		require.NoError(t, err)
		defer opReth.Close()

		// Valid secp256r1 signature → precompile returns 1 (32-byte left-padded).
		response, err := opReth.L2Client.CallContract(ctx, ethereum.CallMsg{
			To:   &rip7212Precompile,
			Data: valid7212Data,
		}, nil)
		require.NoError(t, err)
		require.Equal(t, common.LeftPadBytes([]byte{1}, 32), response,
			"should return 1 for valid secp256r1 signature after Fjord/Arsia activation")

		// Invalid input → precompile returns empty response.
		response, err = opReth.L2Client.CallContract(ctx, ethereum.CallMsg{
			To:   &rip7212Precompile,
			Data: invalid7212Data,
		}, nil)
		require.NoError(t, err)
		require.Equal(t, []byte{}, response,
			"should return empty response for invalid secp256r1 signature")
	})
}

// TestIsthmus verifies the EIP-2537 BLS12-381 precompiles on op-reth with
// Mantle genesis.  These precompiles are introduced by Isthmus.
//
// The test aligns Skadi/Limb/Arsia with the Isthmus time so that the Mantle
// genesis built by NewRethEngine is consistent with the IsthmusSystemConfig.
//
// Note: the mantleopgeth TestIsthmus includes a "BeforeActivation" case that
// tests precompile behaviour before Isthmus is active.  This case is omitted
// here because reth v1.9.3+ always activates all forks (including Isthmus)
// at genesis @0 and cannot reproduce the pre-Isthmus state.
func TestIsthmus(t *testing.T) {
	// reth v1.9.3 activates ALL forks at genesis @0; a future-fork genesis
	// (isthmusTime > 0) causes genesis-hash mismatch and SYNCING errors.
	// Only ActivateAtGenesis is tested here.
	isthmusTime := hexutil.Uint64(0)

	// EIP-2537 test vectors from https://eips.ethereum.org/assets/eip-2537/test-vectors
	precompilesToTest := []struct {
		precompileName        string
		precompileAddr        common.Address
		failInput             []byte
		expectedErrorContains string
		successInput          []byte
		expectedResult        []byte
	}{
		{
			precompileName:        "G1Add",
			precompileAddr:        common.BytesToAddress([]byte{0x0b}),
			failInput:             common.FromHex("0x0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb00000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a2100000000000000000000000000000000112b98340eee2777cc3c14163dea3ec97977ac3dc5c70da32e6e87578f44912e902ccef9efe28d4a78b8999dfbca942600000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a21"),
			successInput:          common.FromHex("0x0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e100000000000000000000000000000000112b98340eee2777cc3c14163dea3ec97977ac3dc5c70da32e6e87578f44912e902ccef9efe28d4a78b8999dfbca942600000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a21"),
			expectedResult:        common.FromHex("0x000000000000000000000000000000000a40300ce2dec9888b60690e9a41d3004fda4886854573974fab73b046d3147ba5b7a5bde85279ffede1b45b3918d82d0000000000000000000000000000000006d3d887e9f53b9ec4eb6cedf5607226754b07c01ace7834f57f3e7315faefb739e59018e22c492006190fba4a870025"),
			expectedErrorContains: "PrecompileError",
		},
		{
			precompileName:        "G2Add",
			precompileAddr:        common.BytesToAddress([]byte{0x0d}),
			failInput:             common.FromHex("0x000000000000000000000000000000001c4bb49d2a0ef12b7123acdd7110bd292b5bc659edc54dc21b81de057194c79b2a5803255959bbef8e7f56c8c12168630000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be00000000000000000000000000000000103121a2ceaae586d240843a398967325f8eb5a93e8fea99b62b9f88d8556c80dd726a4b30e84a36eeabaf3592937f2700000000000000000000000000000000086b990f3da2aeac0a36143b7d7c824428215140db1bb859338764cb58458f081d92664f9053b50b3fbd2e4723121b68000000000000000000000000000000000f9e7ba9a86a8f7624aa2b42dcc8772e1af4ae115685e60abc2c9b90242167acef3d0be4050bf935eed7c3b6fc7ba77e000000000000000000000000000000000d22c3652d0dc6f0fc9316e14268477c2049ef772e852108d269d9c38dba1d4802e8dae479818184c08f9a569d878451"),
			successInput:          common.FromHex("0x00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be00000000000000000000000000000000103121a2ceaae586d240843a398967325f8eb5a93e8fea99b62b9f88d8556c80dd726a4b30e84a36eeabaf3592937f2700000000000000000000000000000000086b990f3da2aeac0a36143b7d7c824428215140db1bb859338764cb58458f081d92664f9053b50b3fbd2e4723121b68000000000000000000000000000000000f9e7ba9a86a8f7624aa2b42dcc8772e1af4ae115685e60abc2c9b90242167acef3d0be4050bf935eed7c3b6fc7ba77e000000000000000000000000000000000d22c3652d0dc6f0fc9316e14268477c2049ef772e852108d269d9c38dba1d4802e8dae479818184c08f9a569d878451"),
			expectedResult:        common.FromHex("0x000000000000000000000000000000000b54a8a7b08bd6827ed9a797de216b8c9057b3a9ca93e2f88e7f04f19accc42da90d883632b9ca4dc38d013f71ede4db00000000000000000000000000000000077eba4eecf0bd764dce8ed5f45040dd8f3b3427cb35230509482c14651713282946306247866dfe39a8e33016fcbe520000000000000000000000000000000014e60a76a29ef85cbd69f251b9f29147b67cfe3ed2823d3f9776b3a0efd2731941d47436dc6d2b58d9e65f8438bad073000000000000000000000000000000001586c3c910d95754fef7a732df78e279c3d37431c6a2b77e67a00c7c130a8fcd4d19f159cbeb997a178108fffffcbd20"),
			expectedErrorContains: "PrecompileError",
		},
		{
			precompileName:        "G1MSM",
			precompileAddr:        common.BytesToAddress([]byte{0x0c}),
			failInput:             common.FromHex("0x0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb00000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a21000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000112b98340eee2777cc3c14163dea3ec97977ac3dc5c70da32e6e87578f44912e902ccef9efe28d4a78b8999dfbca942600000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a210000000000000000000000000000000000000000000000000000000000000002"),
			successInput:          common.FromHex("0x0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e1000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000112b98340eee2777cc3c14163dea3ec97977ac3dc5c70da32e6e87578f44912e902ccef9efe28d4a78b8999dfbca942600000000000000000000000000000000186b28d92356c4dfec4b5201ad099dbdede3781f8998ddf929b4cd7756192185ca7b8f4ef7088f813270ac3d48868a210000000000000000000000000000000000000000000000000000000000000002"),
			expectedResult:        common.FromHex("0x00000000000000000000000000000000148f92dced907361b4782ab542a75281d4b6f71f65c8abf94a5a9082388c64662d30fd6a01ced724feef3e284752038c0000000000000000000000000000000015c3634c3b67bc18e19150e12bfd8a1769306ed010f59be645a0823acb5b38f39e8e0d86e59b6353fdafc59ca971b769"),
			expectedErrorContains: "PrecompileError",
		},
		{
			precompileName:        "G2MSM",
			precompileAddr:        common.BytesToAddress([]byte{0x0e}),
			failInput:             common.FromHex("0x000000000000000000000000000000001c4bb49d2a0ef12b7123acdd7110bd292b5bc659edc54dc21b81de057194c79b2a5803255959bbef8e7f56c8c12168630000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000103121a2ceaae586d240843a398967325f8eb5a93e8fea99b62b9f88d8556c80dd726a4b30e84a36eeabaf3592937f2700000000000000000000000000000000086b990f3da2aeac0a36143b7d7c824428215140db1bb859338764cb58458f081d92664f9053b50b3fbd2e4723121b68000000000000000000000000000000000f9e7ba9a86a8f7624aa2b42dcc8772e1af4ae115685e60abc2c9b90242167acef3d0be4050bf935eed7c3b6fc7ba77e000000000000000000000000000000000d22c3652d0dc6f0fc9316e14268477c2049ef772e852108d269d9c38dba1d4802e8dae479818184c08f9a569d8784510000000000000000000000000000000000000000000000000000000000000002"),
			successInput:          common.FromHex("0x00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000103121a2ceaae586d240843a398967325f8eb5a93e8fea99b62b9f88d8556c80dd726a4b30e84a36eeabaf3592937f2700000000000000000000000000000000086b990f3da2aeac0a36143b7d7c824428215140db1bb859338764cb58458f081d92664f9053b50b3fbd2e4723121b68000000000000000000000000000000000f9e7ba9a86a8f7624aa2b42dcc8772e1af4ae115685e60abc2c9b90242167acef3d0be4050bf935eed7c3b6fc7ba77e000000000000000000000000000000000d22c3652d0dc6f0fc9316e14268477c2049ef772e852108d269d9c38dba1d4802e8dae479818184c08f9a569d8784510000000000000000000000000000000000000000000000000000000000000002"),
			expectedResult:        common.FromHex("0x00000000000000000000000000000000009cc9ed6635623ba19b340cbc1b0eb05c3a58770623986bb7e041645175b0a38d663d929afb9a949f7524656043bccc000000000000000000000000000000000c0fb19d3f083fd5641d22a861a11979da258003f888c59c33005cb4a2df4df9e5a2868832063ac289dfa3e997f21f8a00000000000000000000000000000000168bf7d87cef37cf1707849e0a6708cb856846f5392d205ae7418dd94d94ef6c8aa5b424af2e99d957567654b9dae1d90000000000000000000000000000000017e0fa3c3b2665d52c26c7d4cea9f35443f4f9007840384163d3aa3c7d4d18b21b65ff4380cf3f3b48e94b5eecb221dd"),
			expectedErrorContains: "PrecompileError",
		},
		{
			precompileName:        "MapFpToG1",
			precompileAddr:        common.BytesToAddress([]byte{0x10}),
			failInput:             common.FromHex("0x000000000000000000000000000000002f6d9c5465982c0421b61e74579709b3b5b91e57bdd4f6015742b4ff301abb7ef895b9cce00c33c7d48f8e5fa4ac09ae"),
			successInput:          common.FromHex("0x00000000000000000000000000000000147e1ed29f06e4c5079b9d14fc89d2820d32419b990c1c7bb7dbea2a36a045124b31ffbde7c99329c05c559af1c6cc82"),
			expectedResult:        common.FromHex("0x00000000000000000000000000000000009769f3ab59bfd551d53a5f846b9984c59b97d6842b20a2c565baa167945e3d026a3755b6345df8ec7e6acb6868ae6d000000000000000000000000000000001532c00cf61aa3d0ce3e5aa20c3b531a2abd2c770a790a2613818303c6b830ffc0ecf6c357af3317b9575c567f11cd2c"),
			expectedErrorContains: "PrecompileError",
		},
		{
			precompileName:        "MapFp2ToG2",
			precompileAddr:        common.BytesToAddress([]byte{0x11}),
			failInput:             common.FromHex("0x0000000000000000000000000000000021366f100476ce8d3be6cfc90d59fe13349e388ed12b6dd6dc31ccd267ff000e2c993a063ca66beced06f804d4b8e5af0000000000000000000000000000000002829ce3c021339ccb5caf3e187f6370e1e2a311dec9b75363117063ab2015603ff52c3d3b98f19c2f65575e99e8b78c"),
			successInput:          common.FromHex("0x0000000000000000000000000000000007355d25caf6e7f2f0cb2812ca0e513bd026ed09dda65b177500fa31714e09ea0ded3a078b526bed3307f804d4b93b040000000000000000000000000000000002829ce3c021339ccb5caf3e187f6370e1e2a311dec9b75363117063ab2015603ff52c3d3b98f19c2f65575e99e8b78c"),
			expectedResult:        common.FromHex("0x0000000000000000000000000000000000e7f4568a82b4b7dc1f14c6aaa055edf51502319c723c4dc2688c7fe5944c213f510328082396515734b6612c4e7bb700000000000000000000000000000000126b855e9e69b1f691f816e48ac6977664d24d99f8724868a184186469ddfd4617367e94527d4b74fc86413483afb35b000000000000000000000000000000000caead0fd7b6176c01436833c79d305c78be307da5f6af6c133c47311def6ff1e0babf57a0fb5539fce7ee12407b0a42000000000000000000000000000000001498aadcf7ae2b345243e281ae076df6de84455d766ab6fcdaad71fab60abb2e8b980a440043cd305db09d283c895e3d"),
			expectedErrorContains: "PrecompileError",
		},
		{
			precompileName:        "PairingCheck",
			precompileAddr:        common.BytesToAddress([]byte{0x0f}),
			failInput:             common.FromHex("0x000000000000000000000000000000000123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef00000000000000000000000000000000193fb7cedb32b2c3adc06ec11a96bc0d661869316f5e4a577a9f7c179593987beb4fb2ee424dbb2f5dd891e228b46c4a00000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e100000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000d1b3cc2c7027888be51d9ef691d77bcb679afda66c73f17f9ee3837a55024f78c71363275a75d75d86bab79f74782aa0000000000000000000000000000000013fa4d4a0ad8b1ce186ed5061789213d993923066dddaf1040bc3ff59f825c78df74f2d75467e25e0f55f8a00fa030ed"),
			successInput:          common.FromHex("0x0000000000000000000000000000000017f1d3a73197d7942695638c4fa9ac0fc3688c4f9774b905a14e3a3f171bac586c55e83ff97a1aeffb3af00adb22c6bb0000000000000000000000000000000008b3f481e3aaa0f1a09e30ed741d8ae4fcf5e095d5d00af600db18cb2c04b3edd03cc744a2888ae40caa232946c5e7e100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000024aa2b2f08f0a91260805272dc51051c6e47ad4fa403b02b4510b647ae3d1770bac0326a805bbefd48056c8c121bdb80000000000000000000000000000000013e02b6052719f607dacd3a088274f65596bd0d09920b61ab5da61bbdc7f5049334cf11213945d57e5ac7d055d042b7e000000000000000000000000000000000ce5d527727d6e118cc9cdc6da2e351aadfd9baa8cbdd3a76d429a695160d12c923ac9cc3baca289e193548608b82801000000000000000000000000000000000606c4a02ea734cc32acd2b02bc28b99cb3e287e85a763af267492ab572e99ab3f370d275cec1da1aaa9075ff05f79be"),
			expectedResult:        common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000001"),
			expectedErrorContains: "PrecompileError",
		},
	}

	for _, precompileToTest := range precompilesToTest {
		t.Run(fmt.Sprintf("EIP2537_%s", precompileToTest.precompileName), func(t *testing.T) {
			op_e2e.InitParallel(t)
			cfg := e2esys.IsthmusSystemConfig(t, &isthmusTime)
			// Align Mantle forks with the Isthmus time for well-formed Mantle genesis.
			cfg.DeployConfig.L2GenesisMantleSkadiTimeOffset = &isthmusTime
			cfg.DeployConfig.L2GenesisMantleLimbTimeOffset = &isthmusTime
			cfg.DeployConfig.L2GenesisMantleArsiaTimeOffset = &isthmusTime

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			opReth, err := NewRethEngine(t, ctx, &cfg)
			require.NoError(t, err)
			defer opReth.Close()

			response, err := opReth.L2Client.CallContract(ctx, ethereum.CallMsg{
				To:   &precompileToTest.precompileAddr,
				Data: precompileToTest.successInput,
			}, nil)
			require.NoError(t, err)
			require.Equal(t, precompileToTest.expectedResult, response,
				"EIP-2537 %s: unexpected result for valid input", precompileToTest.precompileName)

			_, err = opReth.L2Client.CallContract(ctx, ethereum.CallMsg{
				To:   &precompileToTest.precompileAddr,
				Data: precompileToTest.failInput,
			}, nil)
			require.Error(t, err)
			require.ErrorContains(t, err, precompileToTest.expectedErrorContains,
				"EIP-2537 %s: unexpected error for invalid input", precompileToTest.precompileName)
		})
	}
}
