package chain

import (
	"github.com/Layr-Labs/datalayr/common"
	"github.com/urfave/cli"
)

var (
	RpcUrlFlagName     = "chain.rpc"
	PrivateKeyFlagName = "chain.private-key"
	ChainIdFlagName    = "chain.chainId"
)

type ClientConfig struct {
	RpcUrl           string
	PrivateKeyString string
	ChainId          uint64
}

func CLIFlags(envPrefix string) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:     RpcUrlFlagName,
			Usage:    "Chain rpc",
			Required: true,
			EnvVar:   common.PrefixEnvVar(envPrefix, "CHAIN_RPC"),
		},
		cli.StringFlag{
			Name:     PrivateKeyFlagName,
			Usage:    "Ethereum private key for disperser",
			Required: true,
			EnvVar:   common.PrefixEnvVar(envPrefix, "PRIVATE_KEY"),
		},
		cli.StringFlag{
			Name:     ChainIdFlagName,
			Usage:    "Id of the chain",
			Required: true,
			EnvVar:   common.PrefixEnvVar(envPrefix, "CHAIN_ID"),
		},
	}
}

func ReadCLIConfig(ctx *cli.Context) ClientConfig {
	cfg := ClientConfig{}
	cfg.RpcUrl = ctx.GlobalString(RpcUrlFlagName)
	cfg.PrivateKeyString = ctx.GlobalString(PrivateKeyFlagName)
	cfg.ChainId = ctx.GlobalUint64(ChainIdFlagName)
	return cfg
}
