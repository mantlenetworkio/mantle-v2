package disperser

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	_ "net/http/pprof"

	"os"

	"time"

	"github.com/Layr-Labs/datalayr/common/contracts"
	"github.com/Layr-Labs/datalayr/common/crypto/bls"
	"github.com/Layr-Labs/datalayr/common/logging"

	"github.com/Layr-Labs/datalayr/common/graphView"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
	"github.com/consensys/gnark-crypto/ecc/bn254"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// Aggregator is seperate to encoder for conn reuse
// optimization is possible by mit balloon incentive
type Disperser struct {
	Config *Config

	KzgEncoderGroup *kzgRs.KzgEncoderGroup

	Logger         *logging.Logger
	Aggregator     *Aggregator
	Cache          *StoreCache
	CodedDataCache *CodedDataCache
	metrics        *Metrics

	GraphClient *graphView.GraphClient
	ChainClient *contracts.DataLayrChainClient
}

// Instantiates a new disperser.
func NewDisperser(config *Config, logger *logging.Logger) (*Disperser, error) {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Make sure config folder exists
	err := os.MkdirAll(config.DbPath, os.ModePerm)
	if err != nil {
		log.Println("Could not create db directory")
		return nil, err
	}

	// Chain Services
	chainLogger := logger.Sublogger("Chain")
	chainClient, err := contracts.NewDataLayrChainClient(
		config.ChainClientConfig,
		config.DlsmAddress,
		chainLogger,
	)
	if err != nil {
		logger.Error().Err(err).Msg("Could not create chain client")
		return nil, err
	}

	kzgEncoderGroup, err := kzgRs.NewKzgEncoderGroup(&config.KzgConfig)
	if err != nil {
		logger.Error().Err(err).Msg("Could not create encoder group")
		return nil, err
	}

	aggregatorLogger := logger.Sublogger("Aggregator")
	Aggregator := NewAggregator(chainClient, config.Timeout, aggregatorLogger)

	Cache := NewStoreCache()
	bigInt := new(big.Int).SetInt64(time.Now().UnixMilli())
	int64BigInt := new(big.Int).SetUint64(9223372036854775807)
	bigInt = bigInt.Mod(bigInt, int64BigInt)
	rand.Seed(int64(bigInt.Uint64()))
	logger.Debug().Msgf("seed %v", bigInt.Uint64())

	graphLogger := logger.Sublogger("Graph")
	gc := graphView.NewGraphClient(config.GraphProvider, graphLogger)

	metricLogger := logger.Sublogger("Metrics")
	metrics := NewMetrics(metricLogger)

	codedDataCacheLogger := logger.Sublogger("CodedDataCache")
	codedDataCache := NewCodedDataCache(
		config.CodedCacheSize,
		config.CodedCacheExpireDuration,
		config.CodedCacheCleanPeriod,
		metrics,
		codedDataCacheLogger,
	)

	disperser := &Disperser{
		Config:          config,
		Logger:          logger.Sublogger("Disperser"),
		KzgEncoderGroup: kzgEncoderGroup,
		Aggregator:      Aggregator,
		Cache:           Cache,
		CodedDataCache:  codedDataCache,
		GraphClient:     gc,
		ChainClient:     chainClient,
		metrics:         metrics,
	}

	return disperser, nil
}

func (d *Disperser) Start(ctx context.Context) error {
	if d.Config.EnableMetrics {
		httpSocket := fmt.Sprintf(":%s", d.Config.MetricsPort)
		d.metrics.Start(httpSocket)
	}

	//dlPmAddr, err := d.ChainClient.Bindings.DlSm.DataLayrPaymentManager(&bind.CallOpts{})
	//if err != nil {
	//	d.Logger.Error().Err(err).Msg("failed to retrieve dlpm addr from dlsm")
	//	return err
	//}
	//
	// wethBalance, err := d.ChainClient.ApprovePaymentToken(ctx, dlPmAddr)
	// if err != nil {
	//	d.Logger.Error().Err(err).Msg("Could not approve balance to pm")
	//	return err
	// }
	//
	// err = d.ChainClient.DepositFutureFees(ctx, wethBalance)
	//if err != nil {
	//	d.Logger.Error().Err(err).Msg("Could not deposit fees into pm")
	//	return err
	//}

	return nil
}

func (d *Disperser) Disperse(ctx context.Context, req StoreRequest) (*Store, *AggregateSigs, error) {
	log := d.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering Disperse function...")
	defer log.Trace().Msg("Exiting Disperse function...")

	store, stateView, err := d.CreateStore(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	// METRICS
	d.metrics.NumRegistered.Set(float64(len(stateView.Registrants)))

	headerHashHex := hexutil.Encode(store.HeaderHash[:])
	log.Info().Msgf("DataStore Encoded with %v. TotalOperatorsIndex %v. BlockNumber %v", headerHashHex, store.TotalOperatorsIndex, store.ReferenceBlockNumber)

	// Initialize Data Store
	txHash, err := d.ChainClient.InitDataStore(
		ctx,
		store.Duration,
		store.ReferenceBlockNumber,
		store.TotalOperatorsIndex,
		store.HeaderBytes,
		store.Fee,
	)
	if err != nil {
		return store, nil, err
	}

	//wait for precommit
	event, ok := d.GraphClient.PollingInitDataStore(
		ctx,
		txHash.Bytes(),
		d.Config.Timeout,
	)
	log.Info().Msgf("InitDataStore pooling completes")

	if !ok {
		return store, nil, ErrPrecommitTimeout
	}

	log.Trace().Msg("InitDataStore event received")
	log.Trace().Msgf("store.HeaderHash %v", hexutil.Encode(store.HeaderHash[:]))
	log.Trace().Msgf("event.MsgHash[:] %v", hexutil.Encode(event.MsgHash[:]))

	// Update store metadata with storeNumber and msgHash
	store.StoreId = event.StoreNumber
	store.MsgHash = event.MsgHash

	log.Debug().
		Uint32("NumSys", store.Header.NumSys).
		Uint32("NumPar", store.Header.NumPar).
		Str("Quorum", store.MantleFirstQuorumThreshold.String()).
		Uint32("storeNumber", store.StoreId).
		Msg("Passing DataStore to aggregator")

	// Disperse to dlns
	aggResult, err := d.Aggregator.Aggregate(ctx, store, stateView)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to aggregate signatures")
		return store, nil, err
	}

	calldata := d.PrepareConfirmationCallData(
		ctx,
		store,
		event.MsgHash[:],
		aggResult.StoredAggPubkeyG1,
		aggResult.UsedAggPubkeyG2,
		aggResult.NonSignerPubkeys,
		aggResult.AggSig,
		stateView.TotalOperator.Index,
		uint64(stateView.TotalStake.Index),
		store.ReferenceBlockNumber,
	)

	log.Trace().Msgf("StoredAggPubKey %v", aggResult.StoredAggPubkeyG1)
	log.Trace().Msgf("UsedAggPubKey %v", aggResult.UsedAggPubkeyG2)
	log.Trace().Msgf("NonSignerPubKeys %v", aggResult.NonSignerPubkeys)
	log.Trace().Msgf("AggSig %v", aggResult.AggSig)
	log.Trace().Msgf("op    Index %v", stateView.TotalOperator.Index)
	log.Trace().Msgf("stake Index %v", stateView.TotalStake.Index)
	log.Trace().Msgf("ReferenceBlockNumber %v", store.ReferenceBlockNumber)

	// msgHashCheck := GetMessageHash(contractEvent)

	err = d.ChainClient.ConfirmDataStore(
		ctx,
		calldata,
		&event,
	)
	if err != nil {
		return store, nil, err
	}

	log.Trace().Msg("Exiting function Disperse")
	return store, aggResult, nil
}

func GetMessageHash(event graphView.DataStoreInit) []byte {
	msg := make([]byte, 0)
	msg = append(msg, uint32ToByteSlice(event.StoreNumber)...)
	msg = append(msg, event.DataCommitment[:]...)
	msg = append(msg, byte(event.Duration))
	msg = append(msg, packTo(uint32ToByteSlice(event.InitTime), 32)...)
	msg = append(msg, uint32ToByteSlice(event.Index)...)
	msgHash := crypto.Keccak256(msg)

	return msgHash
}

func (d *Disperser) CheckDataLength(dataLen int, numRegistrant int) error {
	if dataLen < 31*numRegistrant {
		d.Logger.Info().Msgf("Get a request of length %v", dataLen)
		return ErrInvalidInputLength
	}
	return nil
}

func (d *Disperser) PrepareConfirmationCallData(ctx context.Context, store *Store, msgHash []byte, storedAggPubkeyG1 *bn254.G1Affine, usedAggPubkeyG2 *bn254.G2Affine,
	nonPubkeys []bn254.G1Affine, aggSig *bn254.G1Affine, apkIndex uint32, stakeHistoryLength uint64,
	referenceBlockNumber uint32,
) []byte {
	log := d.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering PrepareConfirmationCallData function...")
	defer log.Trace().Msg("Exiting PrepareConfirmationCallData function...")

	aggSigBytes := bls.SerializeG1(aggSig)
	storedAggPubKeyBytesG1 := bls.SerializeG1(storedAggPubkeyG1)
	usedAggPubKeyBytesG2 := bls.SerializeG2(usedAggPubkeyG2)

	// Get sorted nonSignerPubKeys
	nonPubKeysBytes := make([][]byte, 0)
	nonPubKeyHashes := make([][]byte, 0)
	for _, nonPubKey := range nonPubkeys {
		hashNonPub := bls.SerializeG1(&nonPubKey)
		// TODO  this is a hack
		zeros := bigIntToBytes(new(big.Int).SetUint64(uint64(0)), 4)
		h := append(hashNonPub, zeros...)
		nonPubKeysBytes = append(nonPubKeysBytes, h)
		nonPubKeyHashes = append(nonPubKeyHashes, crypto.Keccak256(hashNonPub))
	}

	quickSort(nonPubKeyHashes, nonPubKeysBytes, 0, len(nonPubKeyHashes)-1)

	flattenedNonPubKeysBytes := make([]byte, 0)
	for i := 0; i < len(nonPubKeysBytes); i++ {
		d.Logger.Printf("%v                %v\n", i, nonPubKeysBytes[i])
		flattenedNonPubKeysBytes = append(flattenedNonPubKeysBytes, nonPubKeysBytes[i]...)
	}

	storeNumberBytes := bigIntToBytes(new(big.Int).SetUint64(uint64(store.StoreId)), 4)
	referenceBlockNumberBytes := bigIntToBytes(new(big.Int).SetUint64(uint64(referenceBlockNumber)), 4)

	totalStakeHistoryIndexBytes := bigIntToBytes(
		new(big.Int).SetUint64(stakeHistoryLength),
		6,
	)

	apkIndexBytes := bigIntToBytes(new(big.Int).SetUint64(uint64(apkIndex)), 4)
	numNonPubKeysBytes := bigIntToBytes(new(big.Int).SetUint64(uint64(len(nonPubkeys))), 4)

	// format calldata
	var calldata []byte
	calldata = append(calldata, msgHash...)
	calldata = append(calldata, totalStakeHistoryIndexBytes...)
	calldata = append(calldata, referenceBlockNumberBytes...)
	calldata = append(calldata, storeNumberBytes...)
	calldata = append(calldata, numNonPubKeysBytes...)
	calldata = append(calldata, flattenedNonPubKeysBytes...)
	calldata = append(calldata, apkIndexBytes...)
	calldata = append(calldata, storedAggPubKeyBytesG1...)
	calldata = append(calldata, usedAggPubKeyBytesG2...)
	calldata = append(calldata, aggSigBytes...)
	return calldata

}
