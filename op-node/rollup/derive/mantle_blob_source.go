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

	// Extract blob hashes from valid batcher transactions
	hashes := []eth.IndexedBlobHash{}
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
		// extract blob hashes
		for _, h := range tx.BlobHashes() {
			idh := eth.IndexedBlobHash{
				Index: uint64(blobIndex),
				Hash:  h,
			}
			hashes = append(hashes, idh)
			blobIndex++
		}
	}

	if len(hashes) == 0 {
		// there are no blobs to fetch
		return []eth.Data{}, nil
	}

	// download the actual blob bodies corresponding to the indexed blob hashes
	blobs, err := ds.blobsFetcher.GetBlobs(ctx, ds.ref, hashes)
	if errors.Is(err, ethereum.NotFound) {
		return nil, NewResetError(fmt.Errorf("failed to fetch blobs: %w", err))
	} else if err != nil {
		return nil, NewTemporaryError(fmt.Errorf("failed to fetch blobs: %w", err))
	}

	// Join all blobs together
	wholeBlobData := make([]byte, 0, len(blobs)*eth.MaxBlobDataSize)
	for _, blob := range blobs {
		if blob == nil {
			ds.log.Error("ignoring nil blob")
			continue
		}
		blobData, err := blob.ToData()
		if err != nil {
			ds.log.Error("ignoring blob due to parse failure", "err", err)
			continue
		}
		wholeBlobData = append(wholeBlobData, blobData...)
	}

	// Decode the joined blob data as an RLP-encoded frame array
	frameData := []eth.Data{}
	if len(wholeBlobData) > 0 {
		err = rlp.DecodeBytes(wholeBlobData, &frameData)
		if err != nil {
			ds.log.Error("DecodeBytes blob failure", "err", err)
			return nil, NewTemporaryError(fmt.Errorf("failed to decode blob data as RLP frame array: %w", err))
		}
	}

	return frameData, nil
}
