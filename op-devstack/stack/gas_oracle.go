package stack

import (
	"log/slog"

	"github.com/ethereum-optimism/optimism/op-service/eth"
)

// GasOracleID identifies a GasOracle service by name and chainID, is type-safe, and can be value-copied and used as map key.
type GasOracleID idWithChain

var _ IDWithChain = (*GasOracleID)(nil)

const GasOracleKind Kind = "GasOracle"

func NewGasOracleID(key string, chainID eth.ChainID) GasOracleID {
	return GasOracleID{
		key:     key,
		chainID: chainID,
	}
}

func (id GasOracleID) String() string {
	return idWithChain(id).string(GasOracleKind)
}

func (id GasOracleID) ChainID() eth.ChainID {
	return idWithChain(id).chainID
}

func (id GasOracleID) Kind() Kind {
	return GasOracleKind
}

func (id GasOracleID) Key() string {
	return id.key
}

func (id GasOracleID) LogValue() slog.Value {
	return slog.StringValue(id.String())
}

func (id GasOracleID) MarshalText() ([]byte, error) {
	return idWithChain(id).marshalText(GasOracleKind)
}

// GasOracle is the interface for gas oracle services in the devstack.
type GasOracle interface {
	Common
	ID() GasOracleID
}
