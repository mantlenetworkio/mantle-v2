package flags

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/ethereum-optimism/optimism/op-node/chaincfg"
	nodeflags "github.com/ethereum-optimism/optimism/op-node/flags"
	"github.com/ethereum-optimism/optimism/op-node/sources"
	service "github.com/ethereum-optimism/optimism/op-service"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
)

const EnvVarPrefix = "OP_PROGRAM"

var (
	RollupConfig = &cli.StringFlag{
		Name:    "rollup.config",
		Usage:   "Rollup chain parameters",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "ROLLUP_CONFIG"),
	}
	Network = &cli.StringFlag{
		Name:    "network",
		Usage:   fmt.Sprintf("Predefined network selection. Available networks: %s", strings.Join(chaincfg.AvailableNetworks(), ", ")),
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "NETWORK"),
	}
	DataDir = &cli.StringFlag{
		Name:    "datadir",
		Usage:   "Directory to use for preimage data storage. Default uses in-memory storage",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "DATADIR"),
	}
	L2NodeAddr = &cli.StringFlag{
		Name:    "l2",
		Usage:   "Address of L2 JSON-RPC endpoint to use (eth and debug namespace required)",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L2_RPC"),
	}
	L1Head = &cli.StringFlag{
		Name:    "l1.head",
		Usage:   "Hash of the L1 head block. Derivation stops after this block is processed.",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L1_HEAD"),
	}
	L2Head = &cli.StringFlag{
		Name:    "l2.head",
		Usage:   "Hash of the agreed L2 block to start derivation from",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L2_HEAD"),
	}
	L2Claim = &cli.StringFlag{
		Name:    "l2.claim",
		Usage:   "Claimed L2 output root to validate",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L2_CLAIM"),
	}
	L2BlockNumber = &cli.Uint64Flag{
		Name:    "l2.blocknumber",
		Usage:   "Number of the L2 block that the claim is from",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L2_BLOCK_NUM"),
	}
	L2GenesisPath = &cli.StringFlag{
		Name:    "l2.genesis",
		Usage:   "Path to the op-geth genesis file",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L2_GENESIS"),
	}
	L1NodeAddr = &cli.StringFlag{
		Name:    "l1",
		Usage:   "Address of L1 JSON-RPC endpoint to use (eth namespace required)",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L1_RPC"),
	}
	L1TrustRPC = &cli.BoolFlag{
		Name:    "l1.trustrpc",
		Usage:   "Trust the L1 RPC, sync faster at risk of malicious/buggy RPC providing bad or inconsistent L1 data",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L1_TRUST_RPC"),
	}
	L1RPCProviderKind = &cli.GenericFlag{
		Name: "l1.rpckind",
		Usage: "The kind of RPC provider, used to inform optimal transactions receipts fetching, and thus reduce costs. Valid options: " +
			nodeflags.EnumString[sources.RPCProviderKind](sources.RPCProviderKinds),
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "L1_RPC_KIND"),
		Value: func() *sources.RPCProviderKind {
			out := sources.RPCKindBasic
			return &out
		}(),
	}
	Exec = &cli.StringFlag{
		Name:    "exec",
		Usage:   "Run the specified client program as a separate process detached from the host. Default is to run the client program in the host process.",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "EXEC"),
	}
	Server = &cli.BoolFlag{
		Name:    "server",
		Usage:   "Run in pre-image server mode without executing any client program.",
		EnvVars: service.PrefixEnvVar(EnvVarPrefix, "SERVER"),
	}
)

// Flags contains the list of configuration options available to the binary.
var Flags []cli.Flag

var requiredFlags = []cli.Flag{
	L1Head,
	L2Head,
	L2Claim,
	L2BlockNumber,
}
var programFlags = []cli.Flag{
	RollupConfig,
	Network,
	DataDir,
	L2NodeAddr,
	L2GenesisPath,
	L1NodeAddr,
	L1TrustRPC,
	L1RPCProviderKind,
	Exec,
	Server,
}

func init() {
	Flags = append(Flags, oplog.CLIFlags(EnvVarPrefix)...)
	Flags = append(Flags, requiredFlags...)
	Flags = append(Flags, programFlags...)
}

func CheckRequired(ctx *cli.Context) error {
	rollupConfig := ctx.String(RollupConfig.Name)
	network := ctx.String(Network.Name)
	if rollupConfig == "" && network == "" {
		return fmt.Errorf("flag %s or %s is required", RollupConfig.Name, Network.Name)
	}
	if rollupConfig != "" && network != "" {
		return fmt.Errorf("cannot specify both %s and %s", RollupConfig.Name, Network.Name)
	}
	if network == "" && ctx.String(L2GenesisPath.Name) == "" {
		return fmt.Errorf("flag %s is required for custom networks", L2GenesisPath.Name)
	}
	for _, flag := range requiredFlags {
		if !ctx.IsSet(flag.Names()[0]) {
			return fmt.Errorf("flag %s is required", flag.Names()[0])
		}
	}
	return nil
}
