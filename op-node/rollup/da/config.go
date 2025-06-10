package da

import "github.com/ethereum-optimism/optimism/op-service/eigenda"

type Config struct {
	eigenda.Config
	// The socket for the MantleDA indexer
	MantleDaIndexerSocket string
	// Whether the MantleDA indexer is enabled
	MantleDAIndexerEnable bool
}
