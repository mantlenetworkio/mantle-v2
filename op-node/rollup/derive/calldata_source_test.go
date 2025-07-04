package derive

import (
	"crypto/ecdsa"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
)

type testTx struct {
	to      *common.Address
	dataLen int
	author  *ecdsa.PrivateKey
	good    bool
	value   int
}

func (tx *testTx) Create(t *testing.T, signer types.Signer, rng *rand.Rand) *types.Transaction {
	t.Helper()
	out, err := types.SignNewTx(tx.author, signer, &types.DynamicFeeTx{
		ChainID:   signer.ChainID(),
		Nonce:     0,
		GasTipCap: big.NewInt(2 * params.GWei),
		GasFeeCap: big.NewInt(30 * params.GWei),
		Gas:       100_000,
		To:        tx.to,
		Value:     big.NewInt(int64(tx.value)),
		Data:      testutils.RandomData(rng, tx.dataLen),
	})
	require.NoError(t, err)
	return out
}

type calldataTest struct {
	name string
	txs  []testTx
}

// TestDataFromEVMTransactions creates some transactions from a specified template and asserts
// that DataFromEVMTransactions properly filters and returns the data from the authorized transactions
// inside the transaction set.
func TestDataFromEVMTransactions(t *testing.T) {
	inboxPriv := testutils.RandomKey()
	batcherPriv := testutils.RandomKey()
	cfg := &rollup.Config{
		L1ChainID:         big.NewInt(100),
		BatchInboxAddress: crypto.PubkeyToAddress(inboxPriv.PublicKey),
	}
	batcherAddr := crypto.PubkeyToAddress(batcherPriv.PublicKey)

	altInbox := testutils.RandomAddress(rand.New(rand.NewSource(1234)))
	altAuthor := testutils.RandomKey()

	testCases := []calldataTest{
		{
			name: "simple",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: batcherPriv, good: true}},
		},
		{
			name: "other inbox",
			txs:  []testTx{{to: &altInbox, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "other author",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: altAuthor, good: false}}},
		{
			name: "inbox is author",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, author: inboxPriv, good: false}}},
		{
			name: "author is inbox",
			txs:  []testTx{{to: &batcherAddr, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "unrelated",
			txs:  []testTx{{to: &altInbox, dataLen: 1234, author: altAuthor, good: false}}},
		{
			name: "contract creation",
			txs:  []testTx{{to: nil, dataLen: 1234, author: batcherPriv, good: false}}},
		{
			name: "empty tx",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 0, author: batcherPriv, good: true}}},
		{
			name: "value tx",
			txs:  []testTx{{to: &cfg.BatchInboxAddress, dataLen: 1234, value: 42, author: batcherPriv, good: true}}},
		{
			name: "empty block", txs: []testTx{},
		},
		{
			name: "mixed txs",
			txs: []testTx{
				{to: &cfg.BatchInboxAddress, dataLen: 1234, value: 42, author: batcherPriv, good: true},
				{to: &cfg.BatchInboxAddress, dataLen: 3333, value: 32, author: altAuthor, good: false},
				{to: &cfg.BatchInboxAddress, dataLen: 2000, value: 22, author: batcherPriv, good: true},
				{to: &altInbox, dataLen: 2020, value: 12, author: batcherPriv, good: false},
			},
		},
		// TODO: test with different batcher key, i.e. when it's changed from initial config value by L1 contract
	}

	for i, tc := range testCases {
		rng := rand.New(rand.NewSource(int64(i)))
		signer := cfg.L1Signer()

		var expectedData []eth.Data
		var txs []*types.Transaction
		for i, tx := range tc.txs {
			txs = append(txs, tx.Create(t, signer, rng))
			if tx.good {
				expectedData = append(expectedData, txs[i].Data())
			}
		}

		out := DataFromEVMTransactions(cfg, batcherAddr, txs, testlog.Logger(t, log.LevelCrit))
		require.ElementsMatch(t, expectedData, out)
	}

}

func TestRLPEncodeDecodeEthData(t *testing.T) {
	var dataS = make([]eth.Data, 0)
	dataS = append(dataS,
		eth.Data(common.Hex2Bytes("test1")),
		eth.Data(common.Hex2Bytes("test2")),
		eth.Data(common.Hex2Bytes("test3")),
	)
	// encode
	bz, err := rlp.EncodeToBytes(dataS)
	require.NoError(t, err)

	// decode
	var dataL = make([]eth.Data, 0)
	err = rlp.DecodeBytes(bz, &dataL)
	require.NoError(t, err)
}

func TestRetrieveBlobTx(t *testing.T) {
}

func TestIsValidBatchTx(t *testing.T) {
	// test setup
	rng := rand.New(rand.NewSource(12345))
	privateKey := testutils.InsecureRandomKey(rng)
	privateKey2 := testutils.InsecureRandomKey(rng)
	publicKey, _ := privateKey.Public().(*ecdsa.PublicKey)
	batcherAddr := crypto.PubkeyToAddress(*publicKey)
	batchInboxAddr := testutils.RandomAddress(rng)
	//logger := testlog.Logger(t, log.LvlInfo)

	chainId := new(big.Int).SetUint64(rng.Uint64())
	signer := types.NewPragueSigner(chainId)

	// valid legacy tx
	txData := &types.LegacyTx{
		Nonce:    rng.Uint64(),
		GasPrice: new(big.Int).SetUint64(rng.Uint64()),
		Gas:      2_000_000,
		To:       &batchInboxAddr,
		Value:    big.NewInt(10),
		Data:     testutils.RandomData(rng, rng.Intn(1000)),
	}
	legacyTx, _ := types.SignNewTx(privateKey, signer, txData)
	res := isValidBatchTx(legacyTx, signer, batchInboxAddr, batcherAddr)
	require.Equal(t, true, res)

	// valid dynamic fee tx
	dynamicFeeTxData := &types.DynamicFeeTx{
		Nonce:     rng.Uint64(),
		GasTipCap: new(big.Int).SetUint64(rng.Uint64()),
		GasFeeCap: new(big.Int).SetUint64(rng.Uint64()),
		Gas:       2_000_000,
		To:        &batchInboxAddr,
		Value:     big.NewInt(10),
		Data:      testutils.RandomData(rng, rng.Intn(1000)),
	}
	dynamicFeeTx, _ := types.SignNewTx(privateKey, signer, dynamicFeeTxData)
	res = isValidBatchTx(dynamicFeeTx, signer, batchInboxAddr, batcherAddr)
	require.Equal(t, true, res)

	// invalid batcher addr
	dynamicFeeTx2, _ := types.SignNewTx(privateKey2, signer, dynamicFeeTxData)
	res = isValidBatchTx(dynamicFeeTx2, signer, batchInboxAddr, batcherAddr)
	require.Equal(t, false, res)

	// valid blob tx
	blobHash := testutils.RandomHash(rng)
	blobTxData := &types.BlobTx{
		Nonce:      rng.Uint64(),
		Gas:        2_000_000,
		To:         batchInboxAddr,
		Data:       testutils.RandomData(rng, rng.Intn(1000)),
		BlobHashes: []common.Hash{blobHash},
	}
	blobTx, _ := types.SignNewTx(privateKey, signer, blobTxData)
	res = isValidBatchTx(blobTx, signer, batchInboxAddr, batcherAddr)
	require.Equal(t, true, res)

	// make sure SetCode transactions are ignored.
	setCodeTxData := &types.SetCodeTx{
		Nonce: rng.Uint64(),
		Gas:   2_000_000,
		To:    batchInboxAddr,
		Data:  testutils.RandomData(rng, rng.Intn(1000)),
	}
	setCodeTx, err := types.SignNewTx(privateKey, types.NewPragueSigner(chainId), setCodeTxData)
	require.NoError(t, err)
	res = isValidBatchTx(setCodeTx, signer, batchInboxAddr, batcherAddr)
	require.Equal(t, false, res)
}
