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

	calldata, _ := hex.DecodeString("ed12b1040a2081527385eba808ba6154daaafa62e746c2ec0b7856f8454091733eda4fa9d050105118a1879a012202000128b5bd01aa06fc03010000f901f6f852f842a01a967fd2735712173f3437ae56a48df44fab54c7cd2191cd95873c78f187370ba00803bd0ba7f6b70f8c659b68374ff566f9bc203c399657599b179cb79206f8d5820310cac480213701c401213720f9019f82b27551f873eba01962d84c8eb1d8f670697a3a44c7bae877c8a9b52f7f0857d485cdeee19f7c30820001824164832683a1a0ece199bfbd970b9e96a93535acdf57f0a1658414ffe9226661e00a423f807e8100832683fca081527385eba808ba6154daaafa62e746c2ec0b7856f8454091733eda4fa9d050b90120b152eb926fc18a88f275dea5e31cba2878a01db7847fea4c5e8c66d4a77d86cc944360c7cb6763f82899f7ae99aaa88ab27518544f0871ad948d18f173ce4b93b5bc2e33b3bf174a5f46e36ef5257bcbab3fc3e27e80911aaa4e4992cdde5e0fc04f2ffff84bc145d195eb06181e458e9411a333a8921756cafa87970380bca6ee16d27cd5b689f68816de5de3ee6b52dde63344277512cf843758e8134ce4cbe825440109b7714d92644ccb9494798d1336b914d49984a0e2e0ff8508d6bf9d92ebc351b1446b3c87b7eb97cb0d6b07c41dba813d48cc8bd723ac731963bab5e4f85c738d4026c47a13937e51aa432cb4afe75cff35e978f07149f11883c542213d41966f66eb8799ed6dda418ad4c72573d57f088fa70446572ca1b230ff01820001")
	calldataFrame := &op_service.CalldataFrame{}
	err := proto.Unmarshal(calldata[1:], calldataFrame)
	if err != nil {
		return
	}
	frame := calldataFrame.Value.(*op_service.CalldataFrame_FrameRef)
	da := da.NewEigenDADataStore(context.Background(), log.New("t1"), &cfg, nil, nil)
	fmt.Printf("%x\n%x\n", frame.FrameRef.BatchHeaderHash, frame.FrameRef.Commitment)
	data, err := da.RetrieveBlob(frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex, nil)
	if err != nil {
		t.Errorf("RetrieveBlob err:%v", err)
		return
	}
	fmt.Printf("RetrieveBlob size:%d blob length:%d\n", len(data), frame.FrameRef.BlobLength)

	data = data[:frame.FrameRef.BlobLength]
	outData := []eth.Data{}
	err = rlp.DecodeBytes(data, &outData)
	if err != nil {
		log.Error("Decode retrieval frames in error,skip wrong data", "err", err, "blobInfo", fmt.Sprintf("%x:%d", frame.FrameRef.BatchHeaderHash, frame.FrameRef.BlobIndex))
		return
	}

	fmt.Printf("RetrieveBlob %d\n", len(outData))
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
