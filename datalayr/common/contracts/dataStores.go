package contracts

import (
	"context"
	"math/big"

	sm "github.com/Layr-Labs/datalayr/common/contracts/bindings/DataLayrServiceManager"
	"github.com/Layr-Labs/datalayr/common/graphView"

	"github.com/ethereum/go-ethereum/common"
)

func (dlcc *DataLayrChainClient) InitDataStore(ctx context.Context, duration uint8, blockNumber uint32, totalOperatorIndex uint32, header []byte, fee *big.Int) (*common.Hash, error) {
	log := dlcc.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering InitDataStore function...")
	defer log.Trace().Msg("Exiting InitDataStore function...")

	//TODO: HANDLE CONFIRMER LESS STUPIDLY :)
	tx, err := dlcc.Bindings.DlSm.InitDataStore(dlcc.NoSendTransactOpts, dlcc.AccountAddress, dlcc.AccountAddress, duration, blockNumber, totalOperatorIndex, header)
	if err != nil {
		log.Error().Err(err).Msg("Error assembling InitDataStore tx")
		return nil, err
	}

	// estimate gas and send tx
	return dlcc.EstimateGasPriceAndLimitAndSendTx(context.Background(), tx, "InitDataStore")
}

func (dlcc *DataLayrChainClient) ConfirmDataStore(
	ctx context.Context,
	calldata []byte,
	idsEv *graphView.DataStoreInit,
) error {
	log := dlcc.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering ConfirmDataStore function...")
	defer log.Trace().Msg("Exiting ConfirmDataStore function...")

	//create the metadata from the event, set duration datastore id to 0 since we don't need it for the method
	//TODO: Handle confirmer less stupidly :)

	// IDataLayrServiceManager.DataStoreMetadata(
	// 	headerHash,
	// 	getNumDataStoresForDuration(duration),
	// 	dataStoresForDuration.dataStoreId,
	// 	blockNumber,
	// 	uint96(fee),
	// 	confirmer,
	// 	bytes32(0)
	// );
	dataStoreMetdata := sm.IDataLayrServiceManagerDataStoreMetadata{
		HeaderHash:           idsEv.DataCommitment,
		DurationDataStoreId:  idsEv.DurationDataStoreId,
		GlobalDataStoreId:    idsEv.StoreNumber,
		ReferenceBlockNumber: idsEv.ReferenceBlockNumber,
		BlockNumber:          uint32(idsEv.InitBlockNumber.Uint64()),
		Fee:                  idsEv.Fee,
		Confirmer:            dlcc.AccountAddress,
		SignatoryRecordHash:  emptyBytes,
	}
	searchData := sm.IDataLayrServiceManagerDataStoreSearchData{
		Duration:  idsEv.Duration,
		Timestamp: new(big.Int).SetUint64(uint64(idsEv.InitTime)),
		Index:     idsEv.Index,
		Metadata:  dataStoreMetdata,
	}

	tx, err := dlcc.Bindings.DlSm.ConfirmDataStore(dlcc.NoSendTransactOpts, calldata, searchData)
	if err != nil {
		log.Error().Err(err).Msg("Error assembling ConfirmDataStore tx")
		return err
	}

	// estimate gas and send tx
	_, err = dlcc.EstimateGasPriceAndLimitAndSendTx(ctx, tx, "ConfirmDataStore")
	return err
}
