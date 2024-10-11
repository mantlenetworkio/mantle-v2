package derive

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-node/rollup/da"
	"github.com/ethereum-optimism/optimism/op-node/testlog"
	"github.com/ethereum-optimism/optimism/op-node/testutils"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
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

		out := DataFromEVMTransactions(cfg, batcherAddr, txs, testlog.Logger(t, log.LvlCrit))
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

func TestRetrieveBlob(t *testing.T) {
	// da := eigenda.NewEigenDAClient(
	cfg := eigenda.Config{
		DisperserUrl:        "disperser-holesky.eigenda.xyz:443",
		ProxyUrl:            "http://127.0.0.1:3100",
		DisperseBlobTimeout: 20 * time.Minute,
		RetrieveBlobTimeout: 20 * time.Minute,
	}
	// 	log.New(context.Background()),
	// 	nil,
	// )

	calldata, _ := hex.DecodeString("ed12f3010a20420bc17c0b13e62ad8204a13d88e0412dfe90d65b208d0df6ecbbf204363cc0d10cc0218c49e99012202000128fcb101a206bd01346430636338376539663261393432316539643165346138333531376130623430356434633336643730656333366535623330636461343666303864663736392d33313337333233383336333333373331333333383331333533323330333133313331333233373266333032663333333332663331326633333333326665336230633434323938666331633134396166626634633839393666623932343237616534316534363439623933346361343935393931623738353262383535")
	calldataFrame := &op_service.CalldataFrame{}
	err := proto.Unmarshal(calldata[1:], calldataFrame)
	if err != nil {
		return
	}
	frame := calldataFrame.Value.(*op_service.CalldataFrame_FrameRef)
	da := da.NewEigenDADataStore(context.Background(), log.New("t1"), &cfg, nil, nil)
	fmt.Printf("%x\n%x\n", frame.FrameRef.BatchHeaderHash, frame.FrameRef.Commitment)
	data, err := da.RetrieveBlob(frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex, frame.FrameRef.Commitment)
	if err != nil {
		t.Errorf("RetrieveBlob err:%v", err)
		return
	}
	fmt.Printf("RetrieveBlob %d\n", len(data))
}

func TestRetrieveBlobTx(t *testing.T) {
}

func TestRetrieveFromDaIndexer(t *testing.T) {
	eigenDA := eigenda.Config{
		ProxyUrl: "disperser-holesky.eigenda.xyz:443",
	}

	eigenDaSyncer := da.NewEigenDADataStore(context.Background(), log.New("t1"), &eigenDA, &da.MantleDataStoreConfig{
		MantleDaIndexerSocket: "127.0.0.1:32111",
		MantleDAIndexerEnable: true,
	}, nil)

	out := []eth.Data{}

	if eigenDaSyncer.IsDaIndexer() {
		data, err := eigenDaSyncer.RetrievalFramesFromDaIndexer("0x8494e3e2c70933fc69b82bc0a851f77716d385b52fa8f386df29b819c717be9b")
		if err != nil {
			fmt.Println("Retrieval frames from eigenDa indexer error", "err", err)
			return
		}
		outData := []eth.Data{}
		err = rlp.DecodeBytes(data, &outData)
		if err != nil {
			fmt.Println("Decode retrieval frames in error,skip wrong data", "err", err)
			return
		}
		out = append(out, outData...)

		fmt.Println(len(out))
		return
	}

}
