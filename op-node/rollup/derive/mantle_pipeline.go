package derive

import (
	"github.com/ethereum-optimism/optimism/op-core/forks"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

func (db *DerivationPipeline) transformMantleStages(oldOrigin, newOrigin eth.L1BlockRef) {
	fork := db.rollupCfg.IsMantleActivationBlock(oldOrigin.Time, newOrigin.Time)
	if fork == forks.MantleNone {
		return
	}

	optimismTransformer := func(fork forks.Name) {
		for _, stage := range db.stages {
			if tf, ok := stage.(ForkTransformer); ok {
				tf.Transform(fork)
			}
		}
	}

	switch fork {
	case forks.MantleArsia:
		db.log.Info("Transforming stages", "mantle fork", fork, "optimism fork", forks.Holocene)
		optimismTransformer(forks.Holocene)
	}
}
