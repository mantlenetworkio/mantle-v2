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

// MantleBlobDataSource fetches blobs, joins them together, and decodes them as an RLP-encoded frame array.
type MantleBlobDataSource struct {
	data         []eth.Data
	ref          eth.L1BlockRef
	batcherAddr  common.Address
	dsCfg        DataSourceConfig
	fetcher      L1TransactionFetcher
	blobsFetcher L1BlobsFetcher
	log          log.Logger
}

// NewMantleBlobDataSource creates a new Mantle blob data source.
func NewMantleBlobDataSource(ctx context.Context, log log.Logger, dsCfg DataSourceConfig, fetcher L1TransactionFetcher, blobsFetcher L1BlobsFetcher, ref eth.L1BlockRef, batcherAddr common.Address) DataIter {
	return &MantleBlobDataSource{
		ref:          ref,
		dsCfg:        dsCfg,
		fetcher:      fetcher,
		log:          log.New("origin", ref),
		batcherAddr:  batcherAddr,
		blobsFetcher: blobsFetcher,
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

// open fetches all blobs from valid batcher transactions, joins them together, and decodes them as an RLP-encoded frame array.
func (ds *MantleBlobDataSource) open(ctx context.Context) ([]eth.Data, error) {
	_, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.ref.Hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open blob data source: %w", err))
		}
		return nil, NewTemporaryError(fmt.Errorf("failed to open blob data source: %w", err))
	}

	// Collect blob hashes per valid batcher transaction
	type txBlobInfo struct {
		txHash common.Hash
		hashes []eth.IndexedBlobHash
	}
	var txBlobInfos []txBlobInfo
	blobIndex := 0
	for _, tx := range txs {
		// skip any non-batcher transactions
		if !isValidBatchTx(tx, ds.dsCfg.l1Signer, ds.dsCfg.batchInboxAddress, ds.batcherAddr, ds.log) {
			blobIndex += len(tx.BlobHashes())
			continue
		}
		// only process blob transactions
		if tx.Type() != types.BlobTxType {
			// skip non-blob batcher transactions for Mantle Everest
			continue
		}
		// extract blob hashes for this transaction
		txHashes := make([]eth.IndexedBlobHash, 0, len(tx.BlobHashes()))
		for _, h := range tx.BlobHashes() {
			idh := eth.IndexedBlobHash{
				Index: uint64(blobIndex),
				Hash:  h,
			}
			txHashes = append(txHashes, idh)
			blobIndex++
		}
		if len(txHashes) > 0 {
			txBlobInfos = append(txBlobInfos, txBlobInfo{
				txHash: tx.Hash(),
				hashes: txHashes,
			})
		}
	}

	if len(txBlobInfos) == 0 {
		// there are no blobs to fetch
		return []eth.Data{}, nil
	}

	// Collect all hashes for a single fetch
	allHashes := make([]eth.IndexedBlobHash, 0)
	for _, info := range txBlobInfos {
		allHashes = append(allHashes, info.hashes...)
	}

	// download the actual blob bodies corresponding to the indexed blob hashes
	blobs, err := ds.blobsFetcher.GetBlobs(ctx, ds.ref, allHashes)
	if errors.Is(err, ethereum.NotFound) {
		return nil, NewResetError(fmt.Errorf("failed to fetch blobs: %w", err))
	} else if err != nil {
		return nil, NewTemporaryError(fmt.Errorf("failed to fetch blobs: %w", err))
	}

	// Create a map from blob index to blob for easy lookup
	blobMap := make(map[uint64]*eth.Blob)
	for i, h := range allHashes {
		blobMap[h.Index] = blobs[i]
	}

	// Process each transaction's blobs separately
	allFrameData := []eth.Data{}
	for _, info := range txBlobInfos {
		// Join blobs for this transaction
		txBlobData := make([]byte, 0, len(info.hashes)*eth.MaxBlobDataSize)
		skipTx := false
		for _, h := range info.hashes {
			blob := blobMap[h.Index]
			if blob == nil {
				ds.log.Error("ignoring tx due to nil blob", "txHash", info.txHash, "blobIndex", h.Index)
				skipTx = true
				break
			}
			blobData, err := blob.ToData()
			if err != nil {
				ds.log.Error("ignoring tx due to blob parse failure", "txHash", info.txHash, "blobIndex", h.Index, "err", err)
				skipTx = true
				break
			}
			txBlobData = append(txBlobData, blobData...)
		}
		if skipTx {
			continue
		}

		// Decode the joined blob data for this transaction as an RLP-encoded frame array
		if len(txBlobData) > 0 {
			var frameData []eth.Data
			err = rlp.DecodeBytes(txBlobData, &frameData)
			if err != nil {
				ds.log.Error("ignoring tx due to RLP decode failure", "txHash", info.txHash, "err", err)
				continue
			}
			allFrameData = append(allFrameData, frameData...)
		}
	}

	return allFrameData, nil
}
