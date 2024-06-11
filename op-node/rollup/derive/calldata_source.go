package derive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/rlp"
	"google.golang.org/protobuf/proto"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup"
	"github.com/ethereum-optimism/optimism/op-service/eigenda"
	seth "github.com/ethereum-optimism/optimism/op-service/eth"
	"github.com/ethereum-optimism/optimism/op-service/proto/gen/op_service/v1"
)

const ConfirmDataStoreEventABI = "ConfirmDataStore(uint32,bytes32)"

var ConfirmDataStoreEventABIHash = crypto.Keccak256Hash([]byte(ConfirmDataStoreEventABI))

type blobOrCalldata struct {
	// union type. exactly one of calldata or blob should be non-nil
	blob     *seth.Blob
	calldata *seth.Data
}

type DataIter interface {
	Next(ctx context.Context) (eth.Data, error)
}

type L1TransactionFetcher interface {
	InfoAndTxsByHash(ctx context.Context, hash common.Hash) (eth.BlockInfo, types.Transactions, error)
	FetchReceipts(ctx context.Context, blockHash common.Hash) (eth.BlockInfo, types.Receipts, error)
}

type L1BlobsFetcher interface {
	// GetBlobs fetches blobs that were confirmed in the given L1 block with the given indexed hashes.
	GetBlobs(ctx context.Context, ref seth.L1BlockRef, hashes []seth.IndexedBlobHash) ([]*seth.Blob, error)
}

type MantleDaSyncer interface {
	RetrievalFramesFromDa(dataStoreId uint32) ([]byte, error)
	RetrievalFramesFromDaIndexer(dataStoreId uint32) ([]byte, error)
	IsDaIndexer() bool
}

// DataSourceFactory readers raw transactions from a given block & then filters for
// batch submitter transactions.
// This is not a stage in the pipeline, but a wrapper for another stage in the pipeline
type DataSourceFactory struct {
	log          log.Logger
	cfg          *rollup.Config
	fetcher      L1TransactionFetcher
	syncer       MantleDaSyncer
	metrics      Metrics
	daClient     eigenda.IEigenDA
	eng          EngineQueueStage
	blobsFetcher L1BlobsFetcher
}

func NewDataSourceFactory(log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, blobsFetcher L1BlobsFetcher, syncer MantleDaSyncer, metrics Metrics, daCfg *eigenda.Config) *DataSourceFactory {
	var daClient eigenda.IEigenDA
	if daCfg != nil {
		daClient = &eigenda.EigenDA{
			Log:    log,
			Config: *daCfg,
		}
	}
	return &DataSourceFactory{log: log, cfg: cfg, fetcher: fetcher, syncer: syncer, metrics: metrics, daClient: daClient, blobsFetcher: blobsFetcher}
}

func (ds *DataSourceFactory) RegisterEngineQueue(eng EngineQueueStage) {
	ds.eng = eng
}

// OpenData returns a DataIter. This struct implements the `Next` function.
func (ds *DataSourceFactory) OpenData(ctx context.Context, id eth.L1BlockRef, batcherAddr common.Address) DataIter {
	return NewDataSource(ctx, ds.log, ds.cfg, ds.fetcher, ds.syncer, ds.metrics, id, batcherAddr, ds.daClient, ds.eng.SafeL2Head(), ds.blobsFetcher)
}

// DataSource is a fault tolerant approach to fetching data.
// The constructor will never fail & it will instead re-attempt the fetcher
// at a later point.
type DataSource struct {
	// Internal state + data
	open bool
	data []eth.Data
	// Required to re-attempt fetching
	id           eth.L1BlockRef
	cfg          *rollup.Config // TODO: `DataFromEVMTransactions` should probably not take the full config
	fetcher      L1TransactionFetcher
	blobsFetcher L1BlobsFetcher
	syncer       MantleDaSyncer
	metrics      Metrics
	log          log.Logger
	daClient     eigenda.IEigenDA

	batcherAddr common.Address
	safeL2Ref   eth.L2BlockRef
}

// NewDataSource creates a new calldata source. It suppresses errors in fetching the L1 block if they occur.
// If there is an error, it will attempt to fetch the result on the next call to `Next`.
func NewDataSource(ctx context.Context, log log.Logger, cfg *rollup.Config, fetcher L1TransactionFetcher, syncer MantleDaSyncer, metrics Metrics, block eth.L1BlockRef, batcherAddr common.Address, daClient eigenda.IEigenDA, safeL2Ref eth.L2BlockRef, blobsFetcher L1BlobsFetcher) DataIter {
	if cfg.MantleDaSwitch {
		daUpgradeCfg := eigenda.GetDaUpgradeConfigForMantle(cfg.L2ChainID)
		if daUpgradeCfg != nil && (safeL2Ref.Number+1 >= daUpgradeCfg.EigenDaUpgradeHeight.Uint64()) {
			log.Info("Derived by Eigenda da", "EigenDaUpgradeHeight", daUpgradeCfg.EigenDaUpgradeHeight, "safeL2Ref", safeL2Ref, "l1InBoxBlock", block)
			_, txs, err := fetcher.InfoAndTxsByHash(ctx, block.Hash)
			if err != nil {
				log.Error("Fetch txs by hash fail", "err", err)
				// Here is the original return method keeping op-stack
				return &DataSource{
					open:         false,
					id:           block,
					cfg:          cfg,
					fetcher:      fetcher,
					syncer:       syncer,
					metrics:      metrics,
					log:          log,
					batcherAddr:  batcherAddr,
					daClient:     daClient,
					safeL2Ref:    safeL2Ref,
					blobsFetcher: blobsFetcher,
				}
			} else {
				data, blobHashes, err := dataFromEigenDa(cfg, txs, daClient, metrics, log.New("origin", block), batcherAddr)
				if err != nil {
					return &DataSource{
						open:         false,
						id:           block,
						cfg:          cfg,
						fetcher:      fetcher,
						syncer:       syncer,
						metrics:      metrics,
						log:          log,
						batcherAddr:  batcherAddr,
						daClient:     daClient,
						safeL2Ref:    safeL2Ref,
						blobsFetcher: blobsFetcher,
					}
				} else {
					log.Info("get data from eigenda", "size", len(data), "blobHashes", blobHashes)
					if blobsFetcher == nil && len(blobHashes) > 0 {
						log.Error("find blob transaction, but blobsFetcher is nil")
						return &DataSource{
							open:         false,
							id:           block,
							cfg:          cfg,
							fetcher:      fetcher,
							syncer:       syncer,
							metrics:      metrics,
							log:          log,
							batcherAddr:  batcherAddr,
							daClient:     daClient,
							safeL2Ref:    safeL2Ref,
							blobsFetcher: blobsFetcher,
						}
					}
					if len(blobHashes) > 0 {
						// download the actual blob bodies corresponding to the indexed blob hashes
						log.Info("get data from blob", "client", blobsFetcher, "blobHashes", blobHashes)
						blobs, err := blobsFetcher.GetBlobs(ctx, seth.L1BlockRef(block), blobHashes)
						if err != nil {
							return &DataSource{
								open:         false,
								id:           block,
								cfg:          cfg,
								fetcher:      fetcher,
								syncer:       syncer,
								metrics:      metrics,
								log:          log,
								batcherAddr:  batcherAddr,
								daClient:     daClient,
								safeL2Ref:    safeL2Ref,
								blobsFetcher: blobsFetcher,
							}
						}
						wholeBlobData := make([]byte, 0, len(blobs)*seth.MaxBlobDataSize)
						for _, blob := range blobs {
							blobData, err := blob.ToData()
							if err != nil {
								log.Error("ignoring blob due to parse failure", "err", err)
								continue
							}
							wholeBlobData = append(wholeBlobData, blobData...)
						}
						frameData := []eth.Data{}
						err = rlp.DecodeBytes(wholeBlobData, &frameData)
						if err != nil {
							log.Error("DecodeBytes blob failure", "err", err)
						}
						data = append(data, frameData...)
						log.Info("get data from blob tx", "size", len(data), "blobHashes", blobHashes)
					}
					return &DataSource{
						open: true,
						data: data,
					}

				}
			}
		} else {
			log.Info("Derived by mantle da", "MantleDaSwitch", cfg.MantleDaSwitch)
			_, receipts, err := fetcher.FetchReceipts(ctx, block.Hash)
			if err != nil {
				log.Error("Fetch txs by hash fail", "err", err)
				// Here is the original return method keeping op-stack
				return &DataSource{
					open:         false,
					id:           block,
					cfg:          cfg,
					fetcher:      fetcher,
					syncer:       syncer,
					metrics:      metrics,
					log:          log,
					batcherAddr:  batcherAddr,
					daClient:     daClient,
					blobsFetcher: blobsFetcher,
				}
			} else {
				data, err := dataFromMantleDa(cfg, receipts, syncer, metrics, log.New("origin", block))
				if err != nil {
					return &DataSource{
						open:         false,
						id:           block,
						cfg:          cfg,
						fetcher:      fetcher,
						syncer:       syncer,
						metrics:      metrics,
						log:          log,
						batcherAddr:  batcherAddr,
						daClient:     daClient,
						blobsFetcher: blobsFetcher,
					}
				} else {
					return &DataSource{
						open: true,
						data: data,
					}

				}
			}
		}

	}
	_, txs, err := fetcher.InfoAndTxsByHash(ctx, block.Hash)
	if err != nil {
		log.Error("Fetch txs by hash fail", "err", err)
		// Here is the original return method keeping op-stack
		return &DataSource{
			open:         false,
			id:           block,
			cfg:          cfg,
			fetcher:      fetcher,
			syncer:       syncer,
			metrics:      metrics,
			log:          log,
			batcherAddr:  batcherAddr,
			blobsFetcher: blobsFetcher,
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
			daUpgradeCfg := eigenda.GetDaUpgradeConfigForMantle(ds.cfg.L2ChainID)
			if daUpgradeCfg.EigenDaUpgradeHeight != nil && (ds.safeL2Ref.Number+1 >= daUpgradeCfg.EigenDaUpgradeHeight.Uint64()) {
				if _, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.id.Hash); err == nil {
					data, blobHashes, err := dataFromEigenDa(ds.cfg, txs, ds.daClient, ds.metrics, log.New("origin", ds.id), ds.batcherAddr)
					if err != nil {
						return nil, NewTemporaryError(fmt.Errorf("failed to open mantle da calldata source: %w", err))
					}
					log.Info("get data from eigenda", "size", len(data), "blobHashes", blobHashes)
					if ds.blobsFetcher == nil && len(blobHashes) > 0 {
						log.Error("find blob transaction, but blobsFetcher is nil")
						return nil, NewResetError(fmt.Errorf("failed to fetch blobs"))
					}
					if len(blobHashes) > 0 {
						// download the actual blob bodies corresponding to the indexed blob hashes
						log.Info("get data from blob", "client", ds.blobsFetcher, "blobHashes", blobHashes)
						blobs, err := ds.blobsFetcher.GetBlobs(ctx, seth.L1BlockRef(ds.id), blobHashes)
						if errors.Is(err, ethereum.NotFound) {
							// If the L1 block was available, then the blobs should be available too. The only
							// exception is if the blob retention window has expired, which we will ultimately handle
							// by failing over to a blob archival service.
							return nil, NewResetError(fmt.Errorf("failed to fetch blobs: %w", err))
						} else if err != nil {
							return nil, NewTemporaryError(fmt.Errorf("failed to fetch blobs: %w", err))
						}
						wholeBlobData := make([]byte, 0, len(blobs)*seth.MaxBlobDataSize)
						for _, blob := range blobs {
							blobData, err := blob.ToData()
							if err != nil {
								ds.log.Error("ignoring blob due to parse failure", "err", err)
								continue
							}
							wholeBlobData = append(wholeBlobData, blobData...)
						}
						frameData := []eth.Data{}
						err = rlp.DecodeBytes(wholeBlobData, &frameData)
						if err != nil {
							log.Error("DecodeBytes blob failure", "err", err)
						}
						data = append(data, frameData...)
						log.Info("get data from blob tx", "size", len(data), "blobHashes", blobHashes)
					}
					ds.open = true
					ds.data = data
				} else if errors.Is(err, ethereum.NotFound) {
					return nil, NewResetError(fmt.Errorf("failed to open eigen da calldata source: %w", err))
				} else {
					return nil, NewTemporaryError(fmt.Errorf("failed to open eigen da calldata source: %w", err))
				}
			} else {
				if _, receipts, err := ds.fetcher.FetchReceipts(ctx, ds.id.Hash); err == nil {
					data, err := dataFromMantleDa(ds.cfg, receipts, ds.syncer, ds.metrics, log.New("origin", ds.id))
					if err != nil {
						return nil, NewTemporaryError(fmt.Errorf("failed to open mantle da calldata source: %w", err))
					}
					ds.open = true
					ds.data = data
				} else if errors.Is(err, ethereum.NotFound) {
					return nil, NewResetError(fmt.Errorf("failed to open mantle da calldata source: %w", err))
				} else {
					return nil, NewTemporaryError(fmt.Errorf("failed to open mantle da calldata source: %w", err))
				}
			}

		} else {
			_, txs, err := ds.fetcher.InfoAndTxsByHash(ctx, ds.id.Hash)
			if err == nil { // fetch data from EOA
				ds.open = true
				ds.data = DataFromEVMTransactions(ds.cfg, ds.batcherAddr, txs, log.New("origin", ds.id))
			} else if errors.Is(err, ethereum.NotFound) {
				return nil, NewResetError(fmt.Errorf("failed to open eoa calldata source: %w", err))
			} else {
				return nil, NewTemporaryError(fmt.Errorf("failed to open eoa calldata source: %w", err))
			}
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

func dataFromMantleDa(config *rollup.Config, receipts types.Receipts, syncer MantleDaSyncer, metrics Metrics, log log.Logger) ([]eth.Data, error) {
	var out []eth.Data
	abiUint32, err := abi.NewType("uint32", "uint32", nil)
	if err != nil {
		log.Error("Abi new uint32 type error", "err", err)
		return out, err
	}
	abiBytes32, err := abi.NewType("bytes32", "bytes32", nil)
	if err != nil {
		log.Error("Abi new bytes32 type error", "err", err)
		return out, err
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
					nextDataStoreId := dataStoreData["dataStoreId"].(uint32)
					dataStoreId := nextDataStoreId - 1
					log.Info("Parse confirmed dataStoreId success", "dataStoreId", dataStoreId, "address", rLog.Address.String())
					var daFrames []byte
					if syncer.IsDaIndexer() {
						daFrames, err = syncer.RetrievalFramesFromDaIndexer(dataStoreId)
					} else {
						daFrames, err = syncer.RetrievalFramesFromDa(dataStoreId)
					}
					if err != nil {
						log.Error("Retrieval frames from mantleDa error", "dataStoreId", dataStoreId, "err", err)
						return out, err
					}
					//skip this dataStore id
					if daFrames == nil {
						return out, nil
					}
					log.Info("Retrieval frames from mantle da success", "daFrames length", len(daFrames), "dataStoreId", dataStoreId)
					err = rlp.DecodeBytes(daFrames, &out)
					if err != nil {
						log.Error("Decode retrieval frames in error,skip wrong data", "err", err, "skip datastore id", dataStoreId)
						continue
					}
					metrics.RecordParseDataStoreId(dataStoreId)
					log.Info("Decode bytes success", "out length", len(out), "dataStoreId", dataStoreId)
				}
				return out, nil
			}
		}
	}
	return out, nil
}

func dataFromEigenDa(config *rollup.Config, txs types.Transactions, daClient eigenda.IEigenDA, metrics Metrics, log log.Logger, batcherAddr common.Address) ([]eth.Data, []seth.IndexedBlobHash, error) {
	out := []eth.Data{}
	var hashes []seth.IndexedBlobHash
	blobIndex := 0 // index of each blob in the block's blob sidecar
	for j, tx := range txs {
		if !isValidBatchTx(tx, config.L1Signer(), config.BatchInboxAddress, batcherAddr) {
			blobIndex += len(tx.BlobHashes())
			continue
		}
		data := tx.Data()
		log.Info("Prefix derivation enabled, checking derivation version")
		switch {
		case len(data) == 0:
			if tx.Type() == types.BlobTxType {
				for _, h := range tx.BlobHashes() {
					idh := seth.IndexedBlobHash{
						Index: uint64(blobIndex),
						Hash:  h,
					}
					hashes = append(hashes, idh)
					blobIndex += 1
				}
			}
			continue
		case data[0] == eigenda.DerivationVersionEigenda:
			log.Info("EigenDA derivation version detected")
			// skip the first byte and unwrap the data with protobuf
			data = data[1:]
		}

		calldataFrame := &op_service.CalldataFrame{}
		err := proto.Unmarshal(data, calldataFrame)
		if err != nil {
			log.Warn("unable to decode calldata frame", "index", j, "err", err)
			return nil, nil, err
		}

		switch calldataFrame.Value.(type) {
		case *op_service.CalldataFrame_FrameRef:
			frameRef := calldataFrame.GetFrameRef()
			if len(frameRef.QuorumIds) == 0 {
				log.Warn("decoded frame ref contains no quorum IDs", "index", j, "err", err)
				return nil, nil, err
			}

			log.Info("requesting data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber,
				"batchHeaderHash", base64.StdEncoding.EncodeToString(frameRef.BatchHeaderHash), "blobIndex", frameRef.BlobIndex, "blobLength", frameRef.BlobLength)
			data, err := daClient.RetrieveBlob(context.Background(), frameRef.BatchHeaderHash, frameRef.BlobIndex)
			if err != nil {
				retrieveReqJSON, _ := json.Marshal(struct {
					BatchHeaderHash string
					BlobIndex       uint32
				}{
					BatchHeaderHash: base64.StdEncoding.EncodeToString(frameRef.BatchHeaderHash),
					BlobIndex:       frameRef.BlobIndex,
				})
				log.Warn("could not retrieve data from EigenDA", "request", string(retrieveReqJSON), "err", err)
				return nil, nil, err
			}
			log.Info("Successfully retrieved data from EigenDA", "quorum id", frameRef.QuorumIds[0], "confirmation block number", frameRef.ReferenceBlockNumber, "blob length", frameRef.BlobLength)
			data = data[:frameRef.BlobLength]
			outData := []eth.Data{}
			err = rlp.DecodeBytes(data, &outData)
			if err != nil {
				log.Error("Decode retrieval frames in error,skip wrong data", "err", err, "blobInfo", fmt.Sprintf("%x:%d", frameRef.BatchHeaderHash, frameRef.BlobIndex))
				continue
			}
			out = append(out, outData...)
		case *op_service.CalldataFrame_Frame:
			log.Info("Successfully read data from calldata (not EigenDA)")
			frame := calldataFrame.GetFrame()
			out = append(out, frame)
		}

	}
	return out, hashes, nil
}

// isValidBatchTx returns true if:
//  1. the transaction has a To() address that matches the batch inbox address, and
//  2. the transaction has a valid signature from the batcher address
func isValidBatchTx(tx *types.Transaction, l1Signer types.Signer, batchInboxAddr, batcherAddr common.Address) bool {
	to := tx.To()
	if to == nil || *to != batchInboxAddr {
		return false
	}
	seqDataSubmitter, err := l1Signer.Sender(tx) // optimization: only derive sender if To is correct
	if err != nil {
		log.Warn("tx in inbox with invalid signature", "hash", tx.Hash(), "err", err)
		return false
	}
	// some random L1 user might have sent a transaction to our batch inbox, ignore them
	if seqDataSubmitter != batcherAddr {
		log.Warn("tx in inbox with unauthorized submitter", "addr", seqDataSubmitter, "hash", tx.Hash(), "err", err)
		return false
	}
	return true
}
