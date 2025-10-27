package derive

import (
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/require"
)

func TestArsiaSourcesMatchSpec(t *testing.T) {
	for _, test := range []struct {
		source       UpgradeDepositSource
		expectedHash string
	}{
		{
			source:       deployArsiaL1BlockSource,
			expectedHash: "0x343f879c393b73e52e8c1ecb51a38c7165faf06e25c482ff5f82333e5ae295e6",
		},
		{
			source:       deployArsiaGasPriceOracleSource,
			expectedHash: "0xfe44184ad58b4cb10db4f9b9aa8aeaec19b2e51d8028bb1ee771cbdd4c1cb5da",
		},
		{
			source:       updateArsiaL1BlockProxySource,
			expectedHash: "0xe353e514c6c2b30b8bba0a78a53aefc37b13a96ea3f21c29d0eed7acc3e17ad2",
		},
		{
			source:       updateArsiaGasPriceOracleProxySource,
			expectedHash: "0x1384418a52a30b6f0c234383ea19f93f3d18923c09dc1a685add20cde372b287",
		},
		{
			source:       enableArsiaSource,
			expectedHash: "0x04bcc18c47051eb12de7428483c4c5c385be859adaba4f6f73bd596971bbad84",
		},
	} {
		require.Equal(t, common.HexToHash(test.expectedHash), test.source.SourceHash(),
			"SourceHash mismatch for %s", test.source.Intent)
	}
}

func toDepositTxn(t *testing.T, data hexutil.Bytes) (common.Address, *types.Transaction) {
	txn := new(types.Transaction)
	err := txn.UnmarshalBinary(data)
	require.NoError(t, err)
	require.Truef(t, txn.IsDepositTx(), "expected deposit txn, got %v", txn.Type())
	require.False(t, txn.IsSystemTx())

	signer := types.NewLondonSigner(big.NewInt(420))
	from, err := signer.Sender(txn)
	require.NoError(t, err)

	return from, txn
}

func TestArsiaNetworkTransactions(t *testing.T) {
	upgradeTxns, err := MantleArsiaNetworkUpgradeTransactions()
	require.NoError(t, err)
	require.Len(t, upgradeTxns, 5, "Arsia upgrade should have 5 transactions")

	// Test 1: Deploy L1Block implementation
	deployL1BlockSender, deployL1Block := toDepositTxn(t, upgradeTxns[0])
	require.Equal(t, deployL1BlockSender, L1BlockArsiaDeployerAddress,
		"L1Block deployer should be 0x4210...0000")
	require.Equal(t, common.HexToAddress("0x4210000000000000000000000000000000000000"), deployL1BlockSender,
		"L1Block deployer address verification")
	require.Equal(t, deployArsiaL1BlockSource.SourceHash(), deployL1Block.SourceHash(),
		"L1Block source hash should match")
	require.Nil(t, deployL1Block.To(), "L1Block deployment should have nil To (contract creation)")
	require.Equal(t, l1BlockArsiaDeploymentBytecode, deployL1Block.Data(),
		"L1Block deployment bytecode should match")

	// Test 2: Deploy GasPriceOracle implementation
	deployGasPriceOracleSender, deployGasPriceOracle := toDepositTxn(t, upgradeTxns[1])
	require.Equal(t, deployGasPriceOracleSender, GasPriceOracleArsiaDeployerAddress,
		"GasPriceOracle deployer should be 0x4210...0001")
	require.Equal(t, common.HexToAddress("0x4210000000000000000000000000000000000001"), deployGasPriceOracleSender,
		"GasPriceOracle deployer address verification")
	require.Equal(t, deployArsiaGasPriceOracleSource.SourceHash(), deployGasPriceOracle.SourceHash(),
		"GasPriceOracle source hash should match")
	require.Nil(t, deployGasPriceOracle.To(), "GasPriceOracle deployment should have nil To (contract creation)")
	require.Equal(t, gasPriceOracleArsiaDeploymentBytecode, deployGasPriceOracle.Data(),
		"GasPriceOracle deployment bytecode should match")

	// Test 3: Update L1Block proxy
	updateL1BlockProxySender, updateL1BlockProxy := toDepositTxn(t, upgradeTxns[2])
	require.Equal(t, updateL1BlockProxySender, common.Address{},
		"L1Block proxy update should be from zero address (proxy admin)")
	require.Equal(t, updateArsiaL1BlockProxySource.SourceHash(), updateL1BlockProxy.SourceHash(),
		"L1Block proxy update source hash should match")
	require.NotNil(t, updateL1BlockProxy.To(), "L1Block proxy update should have non-nil To")
	require.Equal(t, predeploys.L1BlockAddr, *updateL1BlockProxy.To(),
		"L1Block proxy update should target L1Block predeploy address")
	require.Equal(t, common.HexToAddress("0x4200000000000000000000000000000000000015"), *updateL1BlockProxy.To(),
		"L1Block predeploy address verification")

	// Verify upgradeTo calldata format: 0x3659cfe6 + padded address
	require.Equal(t, 4+32, len(updateL1BlockProxy.Data()),
		"L1Block proxy update calldata should be 36 bytes (4 byte selector + 32 byte address)")
	require.Equal(t, common.FromHex("0x3659cfe6"), updateL1BlockProxy.Data()[:4],
		"L1Block proxy update should call upgradeTo(address)")

	// Verify the new implementation address in calldata matches computed address
	expectedL1BlockAddr := arsiaL1BlockAddress
	calldataAddr := common.BytesToAddress(updateL1BlockProxy.Data()[4:])
	require.Equal(t, expectedL1BlockAddr, calldataAddr,
		"L1Block proxy update should point to new implementation address")

	// Test 4: Update GasPriceOracle proxy
	updateGasPriceOracleSender, updateGasPriceOracle := toDepositTxn(t, upgradeTxns[3])
	require.Equal(t, updateGasPriceOracleSender, common.Address{},
		"GasPriceOracle proxy update should be from zero address (proxy admin)")
	require.Equal(t, updateArsiaGasPriceOracleProxySource.SourceHash(), updateGasPriceOracle.SourceHash(),
		"GasPriceOracle proxy update source hash should match")
	require.NotNil(t, updateGasPriceOracle.To(), "GasPriceOracle proxy update should have non-nil To")
	require.Equal(t, predeploys.GasPriceOracleAddr, *updateGasPriceOracle.To(),
		"GasPriceOracle proxy update should target GasPriceOracle predeploy address")
	require.Equal(t, common.HexToAddress("0x420000000000000000000000000000000000000F"), *updateGasPriceOracle.To(),
		"GasPriceOracle predeploy address verification")

	// Verify upgradeTo calldata format
	require.Equal(t, 4+32, len(updateGasPriceOracle.Data()),
		"GasPriceOracle proxy update calldata should be 36 bytes")
	require.Equal(t, common.FromHex("0x3659cfe6"), updateGasPriceOracle.Data()[:4],
		"GasPriceOracle proxy update should call upgradeTo(address)")

	// Verify the new implementation address in calldata
	expectedGPOAddr := arsiaGasPriceOracleAddress
	calldataGPOAddr := common.BytesToAddress(updateGasPriceOracle.Data()[4:])
	require.Equal(t, expectedGPOAddr, calldataGPOAddr,
		"GasPriceOracle proxy update should point to new implementation address")

	// Test 5: Enable Arsia in GasPriceOracle
	enableArsiaSender, enableArsia := toDepositTxn(t, upgradeTxns[4])
	require.Equal(t, enableArsiaSender, L1InfoDepositerAddress,
		"Enable Arsia should be from L1InfoDepositer address")
	require.Equal(t, common.HexToAddress("0xDeaDDEaDDeAdDeAdDEAdDEaddeAddEAdDEAd0001"), enableArsiaSender,
		"L1InfoDepositer address verification")
	require.Equal(t, enableArsiaSource.SourceHash(), enableArsia.SourceHash(),
		"Enable Arsia source hash should match")
	require.NotNil(t, enableArsia.To(), "Enable Arsia should have non-nil To")
	require.Equal(t, predeploys.GasPriceOracleAddr, *enableArsia.To(),
		"Enable Arsia should target GasPriceOracle predeploy address")
	require.Equal(t, common.HexToAddress("0x420000000000000000000000000000000000000F"), *enableArsia.To(),
		"Enable Arsia target address verification")
	require.Equal(t, enableArsiaInput, enableArsia.Data(),
		"Enable Arsia should call setArsia()")
	require.Equal(t, 4, len(enableArsia.Data()),
		"Enable Arsia calldata should be 4 bytes (function selector only)")
}

func TestArsiaDeployerAddresses(t *testing.T) {
	// Verify deployer addresses are correctly set
	require.Equal(t, common.HexToAddress("0x4210000000000000000000000000000000000000"),
		L1BlockArsiaDeployerAddress,
		"L1Block Arsia deployer address should be 0x4210...0000")

	require.Equal(t, common.HexToAddress("0x4210000000000000000000000000000000000001"),
		GasPriceOracleArsiaDeployerAddress,
		"GasPriceOracle Arsia deployer address should be 0x4210...0001")

	// Verify deployer addresses are sequential
	l1BlockLastByte := L1BlockArsiaDeployerAddress.Bytes()[19]
	gpoLastByte := GasPriceOracleArsiaDeployerAddress.Bytes()[19]
	require.Equal(t, l1BlockLastByte+1, gpoLastByte,
		"GasPriceOracle deployer should be one after L1Block deployer")
}

func TestArsiaSetArsiaFunctionSelector(t *testing.T) {
	// Verify setArsia() function selector is correctly computed
	require.Equal(t, 4, len(enableArsiaInput),
		"Function selector should be 4 bytes")

	// The actual selector is computed at init time by crypto.Keccak256("setArsia()")[:4]
	// Expected: keccak256("setArsia()")[0:4]
	expectedSelector := crypto.Keccak256([]byte("setArsia()"))[:4]
	require.Equal(t, expectedSelector, []byte(enableArsiaInput),
		"setArsia() selector should match keccak256 hash")

	t.Logf("setArsia() selector: 0x%x", enableArsiaInput)
}

func TestArsiaUpgradeTransactionOrder(t *testing.T) {
	// Verify the upgrade transaction order is correct
	// This is critical for the upgrade to work properly

	upgradeTxns, err := MantleArsiaNetworkUpgradeTransactions()
	require.NoError(t, err)

	// Expected order:
	// 1. Deploy L1Block implementation
	// 2. Deploy GasPriceOracle implementation
	// 3. Upgrade L1Block proxy
	// 4. Upgrade GasPriceOracle proxy
	// 5. Enable Arsia mode

	// Check deployment transactions come before proxy upgrades
	deployL1Block := upgradeTxns[0]
	deployGPO := upgradeTxns[1]
	upgradeL1BlockProxy := upgradeTxns[2]
	upgradeGPOProxy := upgradeTxns[3]
	enableArsiaTx := upgradeTxns[4]

	// Deployments should have nil To (contract creation)
	var deployL1BlockTx, deployGPOTx types.Transaction
	require.NoError(t, deployL1BlockTx.UnmarshalBinary(deployL1Block))
	require.NoError(t, deployGPOTx.UnmarshalBinary(deployGPO))
	require.Nil(t, deployL1BlockTx.To(), "First tx should be L1Block deployment")
	require.Nil(t, deployGPOTx.To(), "Second tx should be GasPriceOracle deployment")

	// Proxy upgrades should have non-nil To
	var upgradeL1BlockProxyTx, upgradeGPOProxyTx, enableArsiaTxParsed types.Transaction
	require.NoError(t, upgradeL1BlockProxyTx.UnmarshalBinary(upgradeL1BlockProxy))
	require.NoError(t, upgradeGPOProxyTx.UnmarshalBinary(upgradeGPOProxy))
	require.NoError(t, enableArsiaTxParsed.UnmarshalBinary(enableArsiaTx))
	require.NotNil(t, upgradeL1BlockProxyTx.To(), "Third tx should be L1Block proxy upgrade")
	require.NotNil(t, upgradeGPOProxyTx.To(), "Fourth tx should be GasPriceOracle proxy upgrade")
	require.NotNil(t, enableArsiaTxParsed.To(), "Fifth tx should be enable Arsia")

	t.Log("✅ Arsia upgrade transaction order is correct")
}

func TestArsiaComputedAddresses(t *testing.T) {
	// Verify that the computed addresses match what CREATE will generate
	// This is critical - if these don't match, the upgrade will fail

	// L1Block implementation address
	computedL1BlockAddr := crypto.CreateAddress(L1BlockArsiaDeployerAddress, 0)
	require.Equal(t, arsiaL1BlockAddress, computedL1BlockAddr,
		"Computed L1Block address should match arsiaL1BlockAddress")
	t.Logf("L1Block implementation will be deployed at: %s", computedL1BlockAddr.Hex())

	// GasPriceOracle implementation address
	computedGPOAddr := crypto.CreateAddress(GasPriceOracleArsiaDeployerAddress, 0)
	require.Equal(t, arsiaGasPriceOracleAddress, computedGPOAddr,
		"Computed GasPriceOracle address should match arsiaGasPriceOracleAddress")
	t.Logf("GasPriceOracle implementation will be deployed at: %s", computedGPOAddr.Hex())
}
