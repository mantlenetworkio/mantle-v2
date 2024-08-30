package eigenda

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/ethereum-optimism/optimism/op-service/eigenda/encoding/kzg"
	"github.com/ethereum-optimism/optimism/op-service/eigenda/verify"
	"github.com/urfave/cli/v2"
)

type Config struct {
	// TODO(eigenlayer): Update quorum ID command-line parameters to support passing
	// and arbitrary number of quorum IDs.

	// DaRpc is the HTTP provider URL for the Data Availability node.
	RPC string

	// The total amount of time that the batcher will spend waiting for EigenDA to confirm a blob
	StatusQueryTimeout time.Duration

	// The amount of time to wait between status queries of a newly dispersed blob
	StatusQueryRetryInterval time.Duration

	// rpc timeout
	RPCTimeout time.Duration

	// EigenDA whether or not to use cloud hsm
	EnableHsm bool

	// The public-key of EigenDA account in hsm
	HsmCreden string

	// The public-key of EigenDA account in hsm
	HsmPubkey string

	// The API name of EigenDA account in hsm
	HsmAPIName string

	// The private key of EigenDA account if not using cloud hsm
	PrivateKey string

	// ETH vars
	EthRPC               string
	SvcManagerAddr       string
	EthConfirmationDepth uint64

	// KZG vars
	CacheDir string

	G1Path string
	G2Path string

	MaxBlobLength    uint64
	G2PowerOfTauPath string
}

const BytesPerSymbol = 31
const MaxCodingRatio = 8

var MaxSRSPoints = math.Pow(2, 28)

var MaxAllowedBlobSize = uint64(MaxSRSPoints * BytesPerSymbol / MaxCodingRatio)

func (c *Config) GetMaxBlobLength() (uint64, error) {
	if c.MaxBlobLength > MaxAllowedBlobSize {
		return 0, fmt.Errorf("excluding disperser constraints on max blob size, SRS points constrain the maxBlobLength configuration parameter to be less than than ~1 GB (%d bytes)", MaxAllowedBlobSize)
	}

	return c.MaxBlobLength, nil
}

func (c *Config) VerificationCfg() *verify.Config {
	numBytes, err := c.GetMaxBlobLength()
	if err != nil {
		panic(fmt.Errorf("Check() was not called on config object, err is not nil: %w", err))
	}
	kzgCfg := &kzg.KzgConfig{
		G1Path:          c.G1Path,
		G2PowerOf2Path:  c.G2PowerOfTauPath,
		CacheDir:        c.CacheDir,
		SRSOrder:        268435456,     // 2 ^ 32
		SRSNumberToLoad: numBytes / 32, // # of fp.Elements
		NumWorker:       uint64(runtime.GOMAXPROCS(0)),
	}

	if c.EthRPC == "" || c.SvcManagerAddr == "" {
		return &verify.Config{
			Verify:    false,
			KzgConfig: kzgCfg,
		}
	}

	return &verify.Config{
		Verify:               true,
		RPCURL:               c.EthRPC,
		SvcManagerAddr:       c.SvcManagerAddr,
		KzgConfig:            kzgCfg,
		EthConfirmationDepth: c.EthConfirmationDepth,
	}

}

// We add this because the urfave/cli library doesn't support uint32 specifically
func Uint32(ctx *cli.Context, flagName string) uint32 {
	daQuorumIDLong := ctx.Uint64(flagName)
	daQuorumID, success := SafeConvertUInt64ToUInt32(daQuorumIDLong)
	if !success {
		panic(fmt.Errorf("%s must be in the uint32 range", flagName))
	}
	return daQuorumID
}

func SafeConvertUInt64ToUInt32(val uint64) (uint32, bool) {
	if val <= math.MaxUint32 {
		return uint32(val), true
	}
	return 0, false
}
