package graphView

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

const POLLING_INTERVAL = 500 * time.Millisecond

func (g *GraphClient) PollingInitDataStore(
	ctx context.Context,
	txHash []byte,
	pollingDuration time.Duration,
) (DataStoreInit, bool) {
	log := g.Logger.SubloggerId(ctx)

	log.Debug().Msgf("PollingInitDataStore from txHash: %v", hexutil.Encode(txHash[:]))

	exit := time.NewTimer(pollingDuration)
	defer exit.Stop()
	ticker := time.NewTicker(POLLING_INTERVAL)
	defer ticker.Stop()

	for {
		ds, err := g.QueryInitDataStoreByTxHash(txHash[:])
		if err == nil {
			log.Debug().Msg("Got initDataStore")
			return *ds, true
		} else {
			log.Error().Err(err).Msg("Did not get initDataStore")
		}

		select {
		case <-exit.C:
			log.Debug().Msgf("Polling duration exceeded")
			return DataStoreInit{}, false
		case <-ticker.C:
		}
	}
}

// Can optimize code to reduce redundency
func (g *GraphClient) PollingInitDataStoreByMsgHash(
	ctx context.Context,
	msgHash []byte,
	pollingDuration time.Duration,
) (DataStoreInit, bool) {
	log := g.Logger.SubloggerId(ctx)

	log.Trace().Msgf("PollingInitDataStore from msgHash %v", hexutil.Encode(msgHash[:]))

	exit := time.NewTimer(pollingDuration)
	defer exit.Stop()

	ticker := time.NewTicker(POLLING_INTERVAL)
	defer ticker.Stop()

	for {
		ds, err := g.QueryInitDataStoreByMsgHash(msgHash[:])
		if err == nil {
			log.Info().Msg("Got initDataStore")
			return *ds, true
		} else {
			log.Debug().Msg("Did not get initDataStore")
		}

		select {
		case <-exit.C:
			log.Debug().Msgf("Polling duration exceeded")
			return DataStoreInit{}, false
		case <-ticker.C:
		}
	}
}

func (g *GraphClient) PollingConfirmDataStore(
	ctx context.Context,
	headerHash []byte,
	retry uint32,
) (*DataStoreConfirm, bool) {
	log := g.Logger.SubloggerId(ctx)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	count := uint32(0)

	log.Trace().Msgf("Confirm data store. Retry", retry)

	for {
		select {
		case <-ticker.C:
			count += 1
			if count == retry {
				log.Debug().Msgf("Retry Confirmation Over. But failed")
				return nil, false
			} else {
				log.Debug().Msgf("Retry %v times", count)
			}
		}
	}
}
