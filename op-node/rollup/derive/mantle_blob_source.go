package derive

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// MantleBlobDataSource fetches blobs or calldata as appropriate and transforms them into usable rollup data.
// For blob transactions, it first tries to decode as Mantle format (joined blobs with RLP-encoded frame array).
// If that fails, it falls back to standard per-blob decoding.
type MantleBlobDataSource struct {
	data         []eth.Data
	ref          eth.L1BlockRef
	batcherAddr  common.Address
	dsCfg        DataSourceConfig
	fetcher      L1TransactionFetcher
	blobsFetcher L1BlobsFetcher
	log          log.Logger
	blobToggle   func()
}

// NewMantleBlobDataSource creates a new Mantle blob data source.
func NewMantleBlobDataSource(ctx context.Context, log log.Logger, dsCfg DataSourceConfig, fetcher L1TransactionFetcher, blobsFetcher L1BlobsFetcher, ref eth.L1BlockRef, batcherAddr common.Address, toggle func()) DataIter {
	return &MantleBlobDataSource{
		ref:          ref,
		dsCfg:        dsCfg,
		fetcher:      fetcher,
		log:          log.New("origin", ref),
		batcherAddr:  batcherAddr,
		blobsFetcher: blobsFetcher,
		blobToggle:   toggle,
	}
}

// Next returns the next piece of batcher data, or an io.EOF error if no data remains.
func (ds *MantleBlobDataSource) Next(ctx context.Context) (eth.Data, error) {
	if ds.data == nil {
		var err error
		if ds.data, err = ds.open(ctx); err != nil {
			return nil, err
		}
	}

	if len(ds.data) == 0 {
		return nil, io.EOF
	}

	next := ds.data[0]
	ds.data = ds.data[1:]
	return next, nil
}

// batcherTxData holds information for a single batcher transaction, preserving order
type batcherTxData struct {
	txHash   common.Hash
	calldata eth.Data              // non-nil for calldata transactions
	hashes   []eth.IndexedBlobHash // non-empty for blob transactions
}

// open fetches blobs and calldata from valid batcher transactions.
// For blob transactions, it first tries Mantle format (joined blobs, RLP decode),
// and falls back to standard per-blob decoding if that fails.
// Data is returned in the same order as transactions appear in the L1 block.
func (ds *MantleBlobDataSource) open(ctx context.Context) ([]eth.Data, error) {
	_, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.ref.Hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open blob data source: %w", err))
		}
		return nil, NewTemporaryError(fmt.Errorf("failed to open blob data source: %w", err))
	}

	// 1. Extract batcher transactions and blob hashes
	batcherTxs, allBlobHashes := ds.batcherTxsAndHashesFromTxs(txs)
	if len(batcherTxs) == 0 {
		return []eth.Data{}, nil
	}

	// 2. Fetch all blobs at once if there are any
	var blobMap map[uint64]*eth.Blob
	if len(allBlobHashes) > 0 {
		blobs, err := ds.blobsFetcher.GetBlobs(ctx, ds.ref, allBlobHashes)
		if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to fetch blobs: %w", err))
		} else if err != nil {
			return nil, NewTemporaryError(fmt.Errorf("failed to fetch blobs: %w", err))
		}
		blobMap = make(map[uint64]*eth.Blob)
		for i, h := range allBlobHashes {
			blobMap[h.Index] = blobs[i]
		}
	}

	// 3. Construct final result in transaction order
	var allData []eth.Data
	for _, txData := range batcherTxs {
		if txData.calldata != nil {
			allData = append(allData, txData.calldata)
		} else {
			txResult, err := ds.processTxBlobs(txData.txHash, txData.hashes, blobMap)
			if err != nil {
				return nil, NewResetError(fmt.Errorf("failed to process blobs: %w", err))
			}
			allData = append(allData, txResult...)
		}
	}
	return allData, nil
}

// batcherTxsAndHashesFromTxs extracts batcher transaction data and blob hashes from transactions.
// It returns batcher transactions in order and all blob hashes for batch fetching.
func (ds *MantleBlobDataSource) batcherTxsAndHashesFromTxs(txs types.Transactions) ([]batcherTxData, []eth.IndexedBlobHash) {
	var batcherTxs []batcherTxData
	var allBlobHashes []eth.IndexedBlobHash
	blobIndex := 0

	for _, tx := range txs {
		// skip any non-batcher transactions
		if !isValidBatchTx(tx, ds.dsCfg.l1Signer, ds.dsCfg.batchInboxAddress, ds.batcherAddr, ds.log) {
			blobIndex += len(tx.BlobHashes())
			continue
		}
		// handle non-blob batcher transactions by extracting their calldata
		if tx.Type() != types.BlobTxType {
			calldata := eth.Data(tx.Data())
			batcherTxs = append(batcherTxs, batcherTxData{
				txHash:   tx.Hash(),
				calldata: calldata,
			})
			continue
		}
		// handle blob batcher transactions by extracting their blob hashes, ignoring any calldata
		if len(tx.Data()) > 0 {
			ds.log.Warn("blob tx has calldata, which will be ignored", "txhash", tx.Hash())
		}
		// extract blob hashes for this transaction
		txHashes := make([]eth.IndexedBlobHash, 0, len(tx.BlobHashes()))
		for _, h := range tx.BlobHashes() {
			idh := eth.IndexedBlobHash{
				Index: uint64(blobIndex),
				Hash:  h,
			}
			txHashes = append(txHashes, idh)
			allBlobHashes = append(allBlobHashes, idh)
			blobIndex++
		}
		if len(txHashes) > 0 {
			batcherTxs = append(batcherTxs, batcherTxData{
				txHash: tx.Hash(),
				hashes: txHashes,
			})
		}
	}

	return batcherTxs, allBlobHashes
}

// processTxBlobs processes blobs for a single transaction.
// It first tries Mantle format (join all blobs, RLP decode as frame array).
// If that fails, it falls back to standard per-blob decoding.
// Returns (result, error) where error indicates a fatal condition (nil blob).
func (ds *MantleBlobDataSource) processTxBlobs(txHash common.Hash, hashes []eth.IndexedBlobHash, blobMap map[uint64]*eth.Blob) ([]eth.Data, error) {
	txBlobData := make([]byte, 0, len(hashes)*eth.MaxBlobDataSize)
	fallbackResult := make([]eth.Data, 0, len(hashes))
	allBlobsValid := true

	for _, h := range hashes {
		blob := blobMap[h.Index]
		if blob == nil {
			// nil blob is a fatal error, matching BlobDataSource behavior
			return nil, fmt.Errorf("nil blob for tx %s at index %d", txHash, h.Index)
		}
		blobData, err := blob.ToData()
		if err != nil {
			ds.log.Error("ignoring blob due to parse failure", "txHash", txHash, "blobIndex", h.Index, "err", err)
			allBlobsValid = false
			continue
		}
		fallbackResult = append(fallbackResult, blobData)
		txBlobData = append(txBlobData, blobData...)
	}

	// Try Mantle format if all blobs are valid
	if allBlobsValid && len(txBlobData) > 0 {
		var frameData []eth.Data
		if err := rlp.DecodeBytes(txBlobData, &frameData); err == nil {
			ds.log.Debug("decoded tx blobs using Mantle format", "txHash", txHash, "frames", len(frameData))
			return frameData, nil
		} else {
			ds.blobToggle()
			ds.log.Debug("Mantle format decode failed, falling back to standard blob format", "txHash", txHash, "err", err)
		}
	}

	// Fallback: return each valid blob's data individually (standard format)
	return fallbackResult, nil
}
