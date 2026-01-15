package derive

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"io"
	"math/big"
	"math/rand"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/holiman/uint256"
	"github.com/stretchr/testify/require"

	"github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/testlog"
	"github.com/ethereum-optimism/optimism/op-service/testutils"
	"github.com/ethereum/go-ethereum/log"
)

func TestMantleBlobDataSource_Next(t *testing.T) {
	// test setup
	rng := rand.New(rand.NewSource(12345))
	privateKey := testutils.InsecureRandomKey(rng)
	publicKey, _ := privateKey.Public().(*ecdsa.PublicKey)
	batcherAddr := crypto.PubkeyToAddress(*publicKey)
	batchInboxAddr := testutils.RandomAddress(rng)
	logger := testlog.Logger(t, log.LvlInfo)

	chainId := new(big.Int).SetUint64(rng.Uint64())
	signer := types.NewPragueSigner(chainId)
	config := DataSourceConfig{
		l1Signer:          signer,
		batchInboxAddress: batchInboxAddr,
	}

	ref := eth.L1BlockRef{
		Hash:       testutils.RandomHash(rng),
		Number:     rng.Uint64(),
		Time:       rng.Uint64(),
		ParentHash: testutils.RandomHash(rng),
	}

	ctx := context.Background()

	t.Run("successful blob fetch and decode", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		// Create blob transaction with blob hash
		blobHash := testutils.RandomHash(rng)
		blobTxData := &types.BlobTx{
			ChainID:    uint256.NewInt(chainId.Uint64()),
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       []byte{},
			BlobHashes: []common.Hash{blobHash},
		}
		blobTx, _ := types.SignNewTx(privateKey, signer, blobTxData)
		txs := types.Transactions{blobTx}

		// Create frame data and encode as RLP
		frameData := []eth.Data{
			eth.Data{0x01, 0x02, 0x03},
			eth.Data{0x04, 0x05, 0x06},
		}
		encodedFrameData, err := rlp.EncodeToBytes(frameData)
		require.NoError(t, err)

		// Create blob from encoded frame data
		var ethBlob eth.Blob
		err = ethBlob.FromData(encodedFrameData)
		require.NoError(t, err)

		// Setup mocks
		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, txs, nil)

		indexedBlobHash := eth.IndexedBlobHash{
			Index: 0,
			Hash:  blobHash,
		}
		mockBlobsFetcher.ExpectOnGetBlobs(ctx, ref, []eth.IndexedBlobHash{indexedBlobHash}, []*eth.Blob{&ethBlob}, nil)

		// Create data source
		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Test Next() - should return first frame
		data, err := ds.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, frameData[0], data)

		// Test Next() - should return second frame
		data, err = ds.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, frameData[1], data)

		// Test Next() - should return EOF
		data, err = ds.Next(ctx)
		require.Equal(t, io.EOF, err)
		require.Nil(t, data)

		mockFetcher.AssertExpectations(t)
		mockBlobsFetcher.AssertExpectations(t)
	})

	t.Run("no blobs found", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		// Create non-blob batcher transaction
		txData := &types.LegacyTx{
			Nonce:    rng.Uint64(),
			GasPrice: new(big.Int).SetUint64(rng.Uint64()),
			Gas:      2_000_000,
			To:       &batchInboxAddr,
			Value:    big.NewInt(10),
			Data:     testutils.RandomData(rng, rng.Intn(1000)),
		}
		calldataTx, _ := types.SignNewTx(privateKey, signer, txData)
		txs := types.Transactions{calldataTx}

		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, txs, nil)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return EOF immediately since no blob transactions
		data, err := ds.Next(ctx)
		require.Equal(t, io.EOF, err)
		require.Nil(t, data)

		mockFetcher.AssertExpectations(t)
	})

	t.Run("non-batcher blob transaction ignored", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		// Create blob transaction signed by wrong address
		blobHash := testutils.RandomHash(rng)
		blobTxData := &types.BlobTx{
			ChainID:    uint256.NewInt(chainId.Uint64()),
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       []byte{},
			BlobHashes: []common.Hash{blobHash},
		}
		wrongKey := testutils.RandomKey()
		blobTx, _ := types.SignNewTx(wrongKey, signer, blobTxData)
		txs := types.Transactions{blobTx}

		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, txs, nil)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return EOF since transaction is not from batcher
		data, err := ds.Next(ctx)
		require.Equal(t, io.EOF, err)
		require.Nil(t, data)

		mockFetcher.AssertExpectations(t)
	})

	t.Run("block not found error", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, nil, ethereum.NotFound)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return ResetError
		data, err := ds.Next(ctx)
		require.Error(t, err)
		require.Nil(t, data)
		require.ErrorIs(t, err, ErrReset)

		mockFetcher.AssertExpectations(t)
	})

	t.Run("temporary error fetching block", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		expectedErr := errors.New("temporary error")
		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, nil, expectedErr)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return TemporaryError
		data, err := ds.Next(ctx)
		require.Error(t, err)
		require.Nil(t, data)
		require.ErrorIs(t, err, ErrTemporary)

		mockFetcher.AssertExpectations(t)
	})

	t.Run("blob not found error", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		blobHash := testutils.RandomHash(rng)
		blobTxData := &types.BlobTx{
			ChainID:    uint256.NewInt(chainId.Uint64()),
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       []byte{},
			BlobHashes: []common.Hash{blobHash},
		}
		blobTx, _ := types.SignNewTx(privateKey, signer, blobTxData)
		txs := types.Transactions{blobTx}

		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, txs, nil)

		indexedBlobHash := eth.IndexedBlobHash{
			Index: 0,
			Hash:  blobHash,
		}
		mockBlobsFetcher.ExpectOnGetBlobs(ctx, ref, []eth.IndexedBlobHash{indexedBlobHash}, nil, ethereum.NotFound)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return ResetError
		data, err := ds.Next(ctx)
		require.Error(t, err)
		require.Nil(t, data)
		require.ErrorIs(t, err, ErrReset)

		mockFetcher.AssertExpectations(t)
		mockBlobsFetcher.AssertExpectations(t)
	})

	t.Run("temporary error fetching blobs", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		blobHash := testutils.RandomHash(rng)
		blobTxData := &types.BlobTx{
			ChainID:    uint256.NewInt(chainId.Uint64()),
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       []byte{},
			BlobHashes: []common.Hash{blobHash},
		}
		blobTx, _ := types.SignNewTx(privateKey, signer, blobTxData)
		txs := types.Transactions{blobTx}

		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, txs, nil)

		expectedErr := errors.New("temporary blob error")
		indexedBlobHash := eth.IndexedBlobHash{
			Index: 0,
			Hash:  blobHash,
		}
		mockBlobsFetcher.ExpectOnGetBlobs(ctx, ref, []eth.IndexedBlobHash{indexedBlobHash}, nil, expectedErr)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return TemporaryError
		data, err := ds.Next(ctx)
		require.Error(t, err)
		require.Nil(t, data)
		require.ErrorIs(t, err, ErrTemporary)

		mockFetcher.AssertExpectations(t)
		mockBlobsFetcher.AssertExpectations(t)
	})

	t.Run("nil blob ignored", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		blobHash := testutils.RandomHash(rng)
		blobTxData := &types.BlobTx{
			ChainID:    uint256.NewInt(chainId.Uint64()),
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       []byte{},
			BlobHashes: []common.Hash{blobHash},
		}
		blobTx, _ := types.SignNewTx(privateKey, signer, blobTxData)
		txs := types.Transactions{blobTx}

		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, txs, nil)

		indexedBlobHash := eth.IndexedBlobHash{
			Index: 0,
			Hash:  blobHash,
		}
		// Return nil blob
		mockBlobsFetcher.ExpectOnGetBlobs(ctx, ref, []eth.IndexedBlobHash{indexedBlobHash}, []*eth.Blob{nil}, nil)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return EOF since nil blob is ignored
		data, err := ds.Next(ctx)
		require.Equal(t, io.EOF, err)
		require.Nil(t, data)

		mockFetcher.AssertExpectations(t)
		mockBlobsFetcher.AssertExpectations(t)
	})

	t.Run("multiple blob transactions", func(t *testing.T) {
		mockFetcher := &testutils.MockL1Source{}
		mockBlobsFetcher := &testutils.MockBlobsFetcher{}

		// Create two blob transactions, each with two blobs containing its own RLP-encoded frame data
		blobHash1 := testutils.RandomHash(rng)
		blobHash2 := testutils.RandomHash(rng)
		blobHash3 := testutils.RandomHash(rng)
		blobHash4 := testutils.RandomHash(rng)
		blobTxData1 := &types.BlobTx{
			ChainID:    uint256.NewInt(chainId.Uint64()),
			Nonce:      rng.Uint64(),
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       []byte{},
			BlobHashes: []common.Hash{blobHash1, blobHash2},
		}
		blobTxData2 := &types.BlobTx{
			ChainID:    uint256.NewInt(chainId.Uint64()),
			Nonce:      rng.Uint64() + 1,
			Gas:        2_000_000,
			To:         batchInboxAddr,
			Data:       []byte{},
			BlobHashes: []common.Hash{blobHash3, blobHash4},
		}
		blobTx1, _ := types.SignNewTx(privateKey, signer, blobTxData1)
		blobTx2, _ := types.SignNewTx(privateKey, signer, blobTxData2)
		txs := types.Transactions{blobTx1, blobTx2}

		blockInfo := testutils.RandomBlockInfo(rng)
		mockFetcher.ExpectInfoAndTxsByHash(ref.Hash, blockInfo, txs, nil)

		// Create frame data for each transaction (TX-scoped: each tx has its own complete RLP-encoded frames)
		frameData1 := []eth.Data{
			eth.Data{0x01, 0x02},
			eth.Data{0x03, 0x04},
		}
		frameData2 := []eth.Data{
			eth.Data{0x05, 0x06},
			eth.Data{0x07, 0x08},
		}
		encodedFrameData1, err := rlp.EncodeToBytes(frameData1)
		require.NoError(t, err)
		encodedFrameData2, err := rlp.EncodeToBytes(frameData2)
		require.NoError(t, err)

		// Split each transaction's RLP-encoded data across 2 blobs
		// TX1: split encodedFrameData1 across blob1 and blob2
		midPoint1 := len(encodedFrameData1) / 2
		if midPoint1 == 0 {
			midPoint1 = 1
		}
		var ethBlob1 eth.Blob
		err = ethBlob1.FromData(encodedFrameData1[:midPoint1])
		require.NoError(t, err)
		var ethBlob2 eth.Blob
		err = ethBlob2.FromData(encodedFrameData1[midPoint1:])
		require.NoError(t, err)

		// TX2: split encodedFrameData2 across blob3 and blob4
		midPoint2 := len(encodedFrameData2) / 2
		if midPoint2 == 0 {
			midPoint2 = 1
		}
		var ethBlob3 eth.Blob
		err = ethBlob3.FromData(encodedFrameData2[:midPoint2])
		require.NoError(t, err)
		var ethBlob4 eth.Blob
		err = ethBlob4.FromData(encodedFrameData2[midPoint2:])
		require.NoError(t, err)

		indexedBlobHash1 := eth.IndexedBlobHash{Index: 0, Hash: blobHash1}
		indexedBlobHash2 := eth.IndexedBlobHash{Index: 1, Hash: blobHash2}
		indexedBlobHash3 := eth.IndexedBlobHash{Index: 2, Hash: blobHash3}
		indexedBlobHash4 := eth.IndexedBlobHash{Index: 3, Hash: blobHash4}

		mockBlobsFetcher.ExpectOnGetBlobs(ctx, ref,
			[]eth.IndexedBlobHash{indexedBlobHash1, indexedBlobHash2, indexedBlobHash3, indexedBlobHash4},
			[]*eth.Blob{&ethBlob1, &ethBlob2, &ethBlob3, &ethBlob4}, nil)

		ds := NewMantleBlobDataSource(ctx, logger, config, mockFetcher, mockBlobsFetcher, ref, batcherAddr)

		// Should return frames from each transaction separately (TX-scoped decoding)
		// TX1 blobs are joined and decoded to get frameData1
		data, err := ds.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, frameData1[0], data)

		data, err = ds.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, frameData1[1], data)

		// TX2 blobs are joined and decoded to get frameData2
		data, err = ds.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, frameData2[0], data)

		data, err = ds.Next(ctx)
		require.NoError(t, err)
		require.Equal(t, frameData2[1], data)

		// Test Next() - should return EOF
		data, err = ds.Next(ctx)
		require.Equal(t, io.EOF, err)
		require.Nil(t, data)

		mockFetcher.AssertExpectations(t)
		mockBlobsFetcher.AssertExpectations(t)
	})
}
