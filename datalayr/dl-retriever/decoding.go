package retriever

import (
	"context"

	"github.com/Layr-Labs/datalayr/common/header"
	kzgRs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
)

func (r *Retriever) RecoverFrames(ctx context.Context, frames []kzgRs.Frame, indices []uint64, dsHeader header.DataStoreHeader) ([]byte, []kzgRs.Frame, error) {
	log := r.Logger.SubloggerId(ctx)

	encoder, err := r.KzgGroup.GetKzgEncoder(uint64(dsHeader.NumSys), uint64(dsHeader.NumPar), uint64(dsHeader.OrigDataSize))
	if err != nil {
		return nil, nil, err
	}

	data, err := encoder.DecodeSafe(frames, indices, uint64(dsHeader.OrigDataSize))
	if err != nil {
		return nil, nil, err
	}

	numSysReceived := uint32(0)
	for _, j := range indices {
		if j < uint64(dsHeader.NumSys) {
			numSysReceived += 1
		}
	}

	recodeFrames := make([]kzgRs.Frame, dsHeader.NumSys)
	if numSysReceived != dsHeader.NumSys {
		// encode to get all sys frames
		log.Debug().Msgf("recode. numSysReceived %v, numsys %v", numSysReceived, dsHeader.NumSys)
		_, _, recodeFrames, _, err = encoder.EncodeBytes(ctx, data)
		if err != nil {
			return nil, nil, err
		}
	} else {
		for i, j := range indices {
			if j < uint64(dsHeader.NumSys) {
				recodeFrames[j] = frames[i]
			}
		}
	}

	return data, recodeFrames, nil
}
