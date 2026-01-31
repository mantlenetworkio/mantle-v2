package derive

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/solabi"
)

func parseSystemConfigUpdateBaseFee(data []byte) (*big.Int, error) {
	reader := bytes.NewReader(data)
	if pointer, err := solabi.ReadUint64(reader); err != nil || pointer != 32 {
		return nil, NewCriticalError(errors.New("invalid pointer field"))
	}
	if length, err := solabi.ReadUint64(reader); err != nil || length != 32 {
		return nil, NewCriticalError(errors.New("invalid length field"))
	}
	baseFee, err := solabi.ReadUint256(reader)
	if err != nil {
		return nil, NewCriticalError(errors.New("could not read base fee"))
	}
	if !solabi.EmptyReader(reader) {
		return nil, NewCriticalError(errors.New("too many bytes"))
	}
	return baseFee, nil
}
