package crossdomain_test

import (
	"math/big"
	"testing"

	"github.com/ethereum-optimism/optimism/op-chain-ops/crossdomain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/stretchr/testify/require"
)

// TestEncode tests the encoding of a CrossDomainMessage. The assertion was
// created using solidity.
func TestEncode(t *testing.T) {
	t.Parallel()

	t.Run("V0", func(t *testing.T) {
		msg := crossdomain.NewCrossDomainMessage(
			crossdomain.EncodeVersionedNonce(common.Big0, common.Big0),
			common.Address{},
			common.Address{19: 0x01},
			big.NewInt(0),
			big.NewInt(0),
			big.NewInt(5),
			[]byte{},
		)

		require.Equal(t, uint64(0), msg.Version())

		encoded, err := msg.Encode()
		require.Nil(t, err)

		expect := hexutil.MustDecode("0xcbd4ece900000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000")
		require.Equal(t, expect, encoded)
	})

	t.Run("V1", func(t *testing.T) {
		msg := crossdomain.NewCrossDomainMessage(
			crossdomain.EncodeVersionedNonce(common.Big1, common.Big1),
			common.Address{19: 0x01},
			common.Address{19: 0x02},
			big.NewInt(100),
			big.NewInt(100),
			big.NewInt(555),
			[]byte{},
		)

		require.Equal(t, uint64(1), msg.Version())

		encoded, err := msg.Encode()
		require.Nil(t, err)

		expect := hexutil.MustDecode("0xff8daf1500010000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000640000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000022b00000000000000000000000000000000000000000000000000000000000000e00000000000000000000000000000000000000000000000000000000000000000")

		require.Equal(t, expect, encoded)
	})
}

// TestEncode tests the hash of a CrossDomainMessage. The assertion was
// created using solidity.
func TestHash(t *testing.T) {
	t.Parallel()

	t.Run("V0", func(t *testing.T) {
		msg := crossdomain.NewCrossDomainMessage(
			crossdomain.EncodeVersionedNonce(common.Big0, common.Big0),
			common.Address{},
			common.Address{19: 0x01},
			big.NewInt(10),
			big.NewInt(10),
			big.NewInt(5),
			[]byte{},
		)

		require.Equal(t, uint64(0), msg.Version())

		hash, err := msg.Hash()
		require.Nil(t, err)

		expect := common.HexToHash("0x5bb579a193681e7c4d43c8c2e4bc6c2c447d21ef9fa887ca23b2d3f9a0fac065")
		require.Equal(t, expect, hash)
	})

	t.Run("V1", func(t *testing.T) {
		msg := crossdomain.NewCrossDomainMessage(
			crossdomain.EncodeVersionedNonce(common.Big0, common.Big1),
			common.Address{},
			common.Address{19: 0x01},
			big.NewInt(0),
			big.NewInt(0),
			big.NewInt(5),
			[]byte{},
		)

		require.Equal(t, uint64(1), msg.Version())

		hash, err := msg.Hash()
		require.Nil(t, err)

		expect := common.HexToHash("0x61b18968aae1084c5652f8b277849f98e68786e3b15a427dba8d1bacab36acb8")
		require.Equal(t, expect, hash)
	})
}
