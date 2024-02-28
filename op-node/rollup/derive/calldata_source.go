package derive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
)

const ConfirmDataStoreEventABI = "ConfirmDataStore(uint32,bytes32)"

var ConfirmDataStoreEventABIHash = crypto.Keccak256Hash([]byte(ConfirmDataStoreEventABI))

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
	metrics Metrics
}

func NewDataSourceFactory(log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, syncer MantleDaSyncer, metrics Metrics) *DataSourceFactory {
	return &DataSourceFactory{log: log, cfg: cfg, fetcher: fetcher, syncer: syncer, metrics: metrics}
}

// OpenData returns a DataIter. This struct implements the `Next` function.
func (ds *DataSourceFactory) OpenData(ctx context.Context, id eth.BlockID, batcherAddr common.Address) DataIter {
	return NewDataSource(ctx, ds.log, ds.cfg, ds.fetcher, ds.syncer, ds.metrics, id, batcherAddr)
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
	metrics Metrics
	log     log.Logger

	batcherAddr common.Address
}

// NewDataSource creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewDataSource(ctx context.Context, log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, syncer MantleDaSyncer, metrics Metrics, block eth.BlockID, batcherAddr common.Address) DataIter {
	if cfg.MantleDaSwitch {
		log.Info("Derived by mantle da", "MantleDaSwitch", cfg.MantleDaSwitch)
		_, receipts, err := fetcher.FetchReceipts(ctx, block.Hash)
		if err != nil {
			log.Error("Fetch txs by hash fail", "err", err)
			// Here is the original return method keeping op-stack
			return &DataSource{
				open:        false,
				id:          block,
				cfg:         cfg,
				fetcher:     fetcher,
				syncer:      syncer,
				metrics:     metrics,
				log:         log,
				batcherAddr: batcherAddr,
			}
		} else {
			return &DataSource{
				open: true,
				data: dataFromMantleDa(cfg, receipts, syncer, metrics, log.New("origin", block)),
			}
		}
	}
	_, txs, err := fetcher.InfoAndTxsByHash(ctx, block.Hash)
	if err != nil {
		log.Error("Fetch txs by hash fail", "err", err)
		// Here is the original return method keeping op-stack
		return &DataSource{
			open:        false,
			id:          block,
			cfg:         cfg,
			fetcher:     fetcher,
			syncer:      syncer,
			metrics:     metrics,
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
		if ds.cfg.MantleDaSwitch { // fetch data from mantleDA
			if _, receipts, err := ds.fetcher.FetchReceipts(ctx, ds.id.Hash); err == nil {
				ds.open = true
				ds.data = dataFromMantleDa(ds.cfg, receipts, ds.syncer, ds.metrics, log.New("origin", ds.id))
			} else if errors.Is(err, ethereum.NotFound) {
				return nil, NewResetError(fmt.Errorf("failed to open mantle da calldata source: %w", err))
			} else {
				return nil, NewTemporaryError(fmt.Errorf("failed to open mantle da calldata source: %w", err))
			}
		} else if _, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.id.Hash); err == nil { // fetch data from EOA
			ds.open = true
			ds.data = DataFromEVMTransactions(ds.cfg, ds.batcherAddr, txs, log.New("origin", ds.id))
		} else if errors.Is(err, ethereum.NotFound) {
			return nil, NewResetError(fmt.Errorf("failed to open eoa calldata source: %w", err))
		} else {
			return nil, NewTemporaryError(fmt.Errorf("failed to open eoa calldata source: %w", err))
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

func dataFromMantleDa(config *rollup.Config, receipts types.Receipts, syncer MantleDaSyncer, metrics Metrics, log log.Logger) []eth.Data {
	defer func() {
		log.Info("----------- end to loop receipts ", time.Now())
	}()
	var out []eth.Data
	abiUint32, err := abi.NewType("uint32", "uint32", nil)
	if err != nil {
		log.Error("Abi new uint32 type error", "err", err)
		return out
	}
	abiBytes32, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		log.Error("Abi new bytes32 type error", "err", err)
		return out
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
	var dataStoreData = make(map[string]interface{})
	log.Info("------- start to loop receipts ", "time", time.Now())
	for _, receipt := range receipts {
		for _, rLog := range receipt.Logs {
			if strings.ToLower(rLog.Address.String()) != strings.ToLower(config.DataLayrServiceManagerAddr) {
				continue
			}
			if rLog.Topics[0] != ConfirmDataStoreEventABIHash {
				continue
			}
			if len(rLog.Data) > 0 {
				err := confirmDataStoreArgs.UnpackIntoMap(dataStoreData, rLog.Data)
				if err != nil {
					log.Error("Unpack data into map fail", "err", err)
					continue
				}
				if dataStoreData != nil {
					dataStoreId := dataStoreData["dataStoreId"].(uint32)
					log.Info("Parse confirmed dataStoreId success", "dataStoreId", dataStoreId, "address", rLog.Address.String())
					daFrames, err := syncer.RetrievalFramesFromDa(dataStoreId - 1)
					if err != nil {
						log.Error("Retrieval frames from mantleDa error", "dataStoreId", dataStoreId, "err", err)
						continue
					}
					log.Info("Retrieval frames from mantle da success", "daFrames length", len(daFrames), "dataStoreId", dataStoreId)
					err = rlp.DecodeBytes(daFrames, &out)
					if err != nil {
						log.Error("Decode retrieval frames in error", "err", err)
						continue
					}
					metrics.RecordParseDataStoreId(dataStoreId)
					log.Info("Decode bytes success", "out length", len(out), "dataStoreId", dataStoreId)
				}
				return out
			}
		}
	}
	return out
}
