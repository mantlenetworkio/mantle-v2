package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/ethereum-optimism/optimism/op-chain-ops/crossdomain"

	"github.com/ethereum-optimism/optimism/op-chain-ops/db"
	"github.com/mattn/go-isatty"

	"github.com/ethereum-optimism/optimism/op-node/eth"
	"github.com/ethereum-optimism/optimism/op-node/rollup/derive"
	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/log"

	"github.com/ethereum-optimism/optimism/op-chain-ops/genesis"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/urfave/cli"
)

func main() {
	log.Root().SetHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(isatty.IsTerminal(os.Stderr.Fd()))))

	app := &cli.App{
		Name:  "check-migration",
		Usage: "Run sanity checks on a migrated database",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "l1-rpc-url",
				Value:    "http://127.0.0.1:8545",
				Usage:    "RPC URL for an L1 Node",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "ovm-addresses",
				Usage:    "Path to ovm-addresses.json",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "ovm-allowances",
				Usage:    "Path to ovm-allowances.json",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "ovm-messages",
				Usage:    "Path to ovm-messages.json",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "witness-file",
				Usage:    "Path to witness file",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "db-path",
				Usage:    "Path to database",
				Required: true,
			},
			cli.StringFlag{
				Name:     "deploy-config",
				Usage:    "Path to hardhat deploy config file",
				Required: true,
			},
			cli.StringFlag{
				Name:     "l1-system-contracts",
				Usage:    "Path to l1-system-contracts json file",
				Required: true,
			},
			cli.IntFlag{
				Name:  "db-cache",
				Usage: "LevelDB cache size in mb",
				Value: 1024,
			},
			cli.IntFlag{
				Name:  "db-handles",
				Usage: "LevelDB number of handles",
				Value: 60,
			},
		},
		Action: func(ctx *cli.Context) error {
			deployConfig := ctx.String("deploy-config")
			config, err := genesis.NewDeployConfig(deployConfig)
			if err != nil {
				return err
			}

			ovmAddresses, err := crossdomain.NewAddresses(ctx.String("ovm-addresses"))
			if err != nil {
				return err
			}
			ovmAllowances, err := crossdomain.NewAllowances(ctx.String("ovm-allowances"))
			if err != nil {
				return err
			}
			ovmMessages, err := crossdomain.NewSentMessageFromJSON(ctx.String("ovm-messages"))
			if err != nil {
				return err
			}
			evmMessages, evmAddresses, err := crossdomain.ReadWitnessData(ctx.String("witness-file"))
			if err != nil {
				return err
			}

			log.Info(
				"Loaded witness data",
				"ovmAddresses", len(ovmAddresses),
				"evmAddresses", len(evmAddresses),
				"ovmAllowances", len(ovmAllowances),
				"ovmMessages", len(ovmMessages),
				"evmMessages", len(evmMessages),
			)

			migrationData := crossdomain.MigrationData{
				OvmAddresses:  ovmAddresses,
				EvmAddresses:  evmAddresses,
				OvmAllowances: ovmAllowances,
				OvmMessages:   ovmMessages,
				EvmMessages:   evmMessages,
			}

			l1SystemContracts, err := crossdomain.NewL1SystemContracts(ctx.String("l1-system-contracts"))
			if err != nil {
				return err
			}

			l1RpcURL := ctx.String("l1-rpc-url")
			l1Client, err := ethclient.Dial(l1RpcURL)
			if err != nil {
				return err
			}

			var block *types.Block
			tag := config.L1StartingBlockTag
			if tag.BlockNumber != nil {
				block, err = l1Client.BlockByNumber(context.Background(), big.NewInt(tag.BlockNumber.Int64()))
			} else if tag.BlockHash != nil {
				block, err = l1Client.BlockByHash(context.Background(), *tag.BlockHash)
			} else {
				return fmt.Errorf("invalid l1StartingBlockTag in deploy config: %v", tag)
			}
			if err != nil {
				return err
			}

			dbCache := ctx.Int("db-cache")
			dbHandles := ctx.Int("db-handles")

			// Read the required deployment addresses from disk if required
			if err := config.GetDeployedAddresses(nil, l1SystemContracts); err != nil {
				return err
			}

			if err := config.Check(); err != nil {
				return err
			}

			postLDB, err := db.Open(ctx.String("db-path"), dbCache, dbHandles)
			if err != nil {
				return err
			}

			if err := genesis.PostCheckMigratedDB(
				postLDB,
				migrationData,
				&config.L1CrossDomainMessengerProxy,
				config.L1ChainID,
				config.L2ChainID,
				config.FinalSystemOwner,
				config.ProxyAdminOwner,
				&derive.L1BlockInfo{
					Number:        block.NumberU64(),
					Time:          block.Time(),
					BaseFee:       block.BaseFee(),
					BlockHash:     block.Hash(),
					BatcherAddr:   config.BatchSenderAddress,
					L1FeeOverhead: eth.Bytes32(common.BigToHash(new(big.Int).SetUint64(config.GasPriceOracleOverhead))),
					L1FeeScalar:   eth.Bytes32(common.BigToHash(new(big.Int).SetUint64(config.GasPriceOracleScalar))),
				},
				&genesis.GasPriceOracleConfig{
					TokenRatio:          big.NewInt(int64(config.GasPriceOracleTokenRatio)),
					GasPriceOracleOwner: config.GasPriceOracleOwner,
				},
			); err != nil {
				return err
			}

			if err := postLDB.Close(); err != nil {
				return err
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Crit("error in migration", "err", err)
	}
}
