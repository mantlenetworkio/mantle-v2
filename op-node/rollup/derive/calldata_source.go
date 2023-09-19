package derive

import (
	"context"
	"errors"
	"fmt"
	"github.com/ethereum-optimism/optimism/l2geth/rlp"
	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"io"
)

var (
	ConfirmDataStoreEventABI     = "ConfirmDataStore(uint32,bytes32)"
	ConfirmDataStoreEventABIHash = crypto.Keccak256Hash([]byte(ConfirmDataStoreEventABI))
)

type DataIter interface {
	Next(ctx context.Context) (eth.Data, error)
}

type L1TransactionFetcher interface {
	InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error)
	FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error)
}

type MantleDaSyncer interface {
	RetrievalFramesFromDa(dataStoreId uint32) ([]byte, error)
}

// DataSourceFactory readers raw transactions from a given block & then filters for
// batch submitter transactions.
// This is not a stage in the pipeline, but a wrapper for another stage in the pipeline
type DataSourceFactory struct {
	log     log.Logger
	cfg     *rollup.Config
	fetcher L1TransactionFetcher
	syncer  MantleDaSyncer
}

func NewDataSourceFactory(log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, syncer MantleDaSyncer) *DataSourceFactory {
	return &DataSourceFactory{log: log, cfg: cfg, fetcher: fetcher, syncer: syncer}
}

// OpenData returns a DataIter. This struct implements the `Next` function.
func (ds *DataSourceFactory) OpenData(ctx context.Context, id eth.BlockID, batcherAddr common.Address) DataIter {
	return NewDataSource(ctx, ds.log, ds.cfg, ds.fetcher, ds.syncer, id, batcherAddr)
}

// DataSource is a fault tolerant approach to fetching data.
// The constructor will never fail & it will instead re-attempt the fetcher
// at a later point.
type DataSource struct {
	// Internal state + data
	open bool
	data []eth.Data
	// Required to re-attempt fetching
	id      eth.BlockID
	cfg     *rollup.Config // TODO: `DataFromEVMTransactions` should probably not take the full config
	fetcher L1TransactionFetcher
	syncer  MantleDaSyncer
	log     log.Logger

	batcherAddr common.Address
}

// NewDataSource creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewDataSource(ctx context.Context, log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, syncer MantleDaSyncer, block eth.BlockID, batcherAddr common.Address) DataIter {
	if cfg.MantleDaSwitch {
		log.Info("Derived by mantle da", "MantleDaSwitch", cfg.MantleDaSwitch)
		_, receipts, err := fetcher.FetchReceipts(ctx, block.Hash)
		if err != nil {
			return &DataSource{
				open:        false,
				id:          block,
				cfg:         cfg,
				fetcher:     fetcher,
				syncer:      syncer,
				log:         log,
				batcherAddr: batcherAddr,
			}
		} else {
			return &DataSource{
				open: true,
				data: DataFromMantleDa(cfg, receipts, syncer, log.New("origin", block)),
			}
		}
	}
	_, txs, err := fetcher.InfoAndTxsByHash(ctx, block.Hash)
	if err != nil {
		return &DataSource{
			open:        false,
			id:          block,
			cfg:         cfg,
			fetcher:     fetcher,
			syncer:      syncer,
			log:         log,
			batcherAddr: batcherAddr,
		}
	} else {
		return &DataSource{
			open: true,
			data: DataFromEVMTransactions(cfg, batcherAddr, txs, log.New("origin", block)),
		}
	}
}

// Next returns the next piece of data if it has it. If the constructor failed, this
// will attempt to reinitialize itself. If it cannot find the block it returns a ResetError
// otherwise it returns a temporary error if fetching the block returns an error.
func (ds *DataSource) Next(ctx context.Context) (eth.Data, error) {
	if !ds.open {
		if ds.cfg.MantleDaSwitch {
			if _, receipts, err := ds.fetcher.FetchReceipts(ctx, ds.id.Hash); err == nil {
				ds.open = true
				ds.data = DataFromMantleDa(ds.cfg, receipts, ds.syncer, log.New("origin", ds.id))
			}
		} else if _, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.id.Hash); err == nil {
			ds.open = true
			ds.data = DataFromEVMTransactions(ds.cfg, ds.batcherAddr, txs, log.New("origin", ds.id))
		} else if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open calldata source: %w", err))
		} else {
			return nil, NewTemporaryError(fmt.Errorf("failed to open calldata source: %w", err))
		}
	}
	if len(ds.data) == 0 {
		return nil, io.EOF
	} else {
		data := ds.data[0]
		ds.data = ds.data[1:]
		return data, nil
	}
}

// DataFromEVMTransactions filters all of the transactions and returns the calldata from transactions
// that are sent to the batch inbox address from the batch sender address.
// This will return an empty array if no valid transactions are found.
func DataFromEVMTransactions(config *rollup.Config, batcherAddr common.Address, txs types.Transactions, log log.Logger) []eth.Data {
	var out []eth.Data
	l1Signer := config.L1Signer()
	for j, tx := range txs {
		if to := tx.To(); to != nil && *to == config.BatchInboxAddress {
			seqDataSubmitter, err := l1Signer.Sender(tx) // optimization: only derive sender if To is correct
			if err != nil {
				log.Warn("tx in inbox with invalid signature", "index", j, "err", err)
				continue // bad signature, ignore
			}
			// some random L1 user might have sent a transaction to our batch inbox, ignore them
			if seqDataSubmitter != batcherAddr {
				log.Warn("tx in inbox with unauthorized submitter", "index", j, "err", err)
				continue // not an authorized batch submitter, ignore
			}
			out = append(out, tx.Data())
		}
	}
	return out
}

func DataFromMantleDa(config *rollup.Config, receipts types.Receipts, syncer MantleDaSyncer, log log.Logger) []eth.Data {
	var out []eth.Data
	abiUint32, err := abi.NewType("uint32", "uint32", nil)
	if err != nil {
		log.Error("Abi new uint32 type error", "err", err)
		return nil
	}
	abiBytes32, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		log.Error("Abi new bytes32 type error", "err", err)
		return nil
	}
	confirmDataStoreArgs := abi.Arguments{
		{
			Name:    "dataStoreId",
			Type:    abiUint32,
			Indexed: false,
		}, {
			Name:    "headerHash",
			Type:    abiBytes32,
			Indexed: false,
		},
	}
	var dlsmData = make(map[string]interface{})
	for _, receipt := range receipts {
		for _, rlog := range receipt.Logs {
			if rlog.Topics[0] == ConfirmDataStoreEventABIHash {
				if len(rlog.Data) > 0 {
					err := confirmDataStoreArgs.UnpackIntoMap(dlsmData, rlog.Data)
					if err != nil {
						log.Error("unpack data into map fail", "err", err)
						continue
					}
				}
			}
		}
	}
	if len(dlsmData) > 0 {
		dataStoreId := dlsmData["dataStoreId"].(uint32)
		log.Info("Parse confirmed dataStoreId success", "dataStoreId", dlsmData["dataStoreId"].(uint32))
		// fetch frame by dataStoreId
		daFrames, err := syncer.RetrievalFramesFromDa(dataStoreId)
		if err != nil {
			log.Error("Retrieval frames from mantleDa error", "dataStoreId", dataStoreId, "err", err)
			return nil
		}
		err = rlp.DecodeBytes(daFrames, &out)
		if err != nil {
			log.Error("Decode retrieval frames in error", "err", err)
			return nil
		}
		log.Info("Decode bytes success", "out length", len(out))
	}
	return out
}
