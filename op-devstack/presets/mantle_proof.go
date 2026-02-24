package presets

import (
	gameTypes "github.com/ethereum-optimism/optimism/op-challenger/game/types"
	"github.com/ethereum-optimism/optimism/op-devstack/stack"
	"github.com/ethereum-optimism/optimism/op-devstack/sysgo"
)

// WithMantleProofs enables a Mantle minimal system with permissionless proofs enabled.
func WithMantleProofs() stack.CommonOption {
	return stack.MakeCommon(sysgo.MantleProofSystem(&sysgo.DefaultMinimalSystemIDs{}))
}

// WithMantleDisputeGameV2 applies V2 dispute games flags on top of Mantle proofs.
func WithMantleDisputeGameV2() stack.CommonOption {
	return stack.Combine(
		WithMantleProofs(),
		WithDisputeGameV2(),
	)
}

// WithMantleAddedGameType adds a game type for Mantle systems.
func WithMantleAddedGameType(gameType gameTypes.GameType) stack.CommonOption {
	return stack.Combine(
		stack.MakeCommon(sysgo.WithGameTypeAdded(gameType)),
		RequireGameTypePresent(gameType),
	)
}
