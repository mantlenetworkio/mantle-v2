package contracts

import (
	"context"
	"math/big"

	contractBLSPublicKeyCompendium "github.com/Layr-Labs/datalayr/common/contracts/bindings/BLSPublicKeyCompendium"
	contractBLSRegistry "github.com/Layr-Labs/datalayr/common/contracts/bindings/BLSRegistry"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	"github.com/Layr-Labs/datalayr/common/crypto/bls"
)

var (
	//use bytes32(0) for ephemeral key because bomb functionality not implemented yet
	emptyBytes = [32]byte{}
)

// *        (1) is registrantType that specifies whether the operator is an ETH DataLayr node,
// *            or Eigen DataLayr node or both,
// *        (2) is mantleFirstStakeLength which specifies length of (3),
// *        (3) is the list of the form [<(operator address, operator's ETH deposit)>, total ETH deposit],
// *            where <(operator address, operator's ETH deposit)> is the array of tuple
// *            (operator address, operator's ETH deposit) for operators who are DataLayr nodes,
// *        (4) is mantleSencodStakeLength which specifies length of (5),
// *        (5) is the list of the form [<(operator address, operator's Eigen deposit)>, total Eigen deposit],
// *            where <(operator address, operator's Eigen deposit)> is the array of tuple
// *            (operator address, operator's Eigen deposit) for operators who are DataLayr nodes,
// *        (6) is socketLength which specifies length of (7),
// *        (7) is the socket
// TODO: change this to bytes!
func (dlcc *DataLayrChainClient) Register(ctx context.Context, keypair *bls.BlsKeyPair, operatorType uint8, socket string, fee *big.Int) error {
	log := dlcc.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering Register function...")
	defer log.Trace().Msg("Exiting Register function...")

	pubkeyG1 := bls.ToG1Point(keypair.PublicKey)
	pubkeyG1CompendiumParam := contractBLSPublicKeyCompendium.BN254G1Point{X: pubkeyG1.X, Y: pubkeyG1.Y}
	pubkeyG1RegistryParam := contractBLSRegistry.BN254G1Point{X: pubkeyG1.X, Y: pubkeyG1.Y}

	err := dlcc.RegisterBLSPublicKey(ctx, keypair, pubkeyG1CompendiumParam)
	if err != nil {
		log.Error().Err(err).Msg("Failed to register BLS pubkey")
		return err
	}

	// assemble tx
	tx, err := dlcc.Bindings.DlReg.RegisterOperator(dlcc.NoSendTransactOpts, operatorType, pubkeyG1RegistryParam, socket)
	if err != nil {
		log.Error().Err(err).Msg("Error assembling RegisterOperator tx")
		return err
	}

	// estimate gas and send tx
	_, err = dlcc.EstimateGasPriceAndLimitAndSendTx(context.Background(), tx, "RegisterOperator")
	return err
}

func (dlcc *DataLayrChainClient) UpdateSocket(ctx context.Context, socket string) error {
	log := dlcc.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering OptIntoSlashing function...")
	defer log.Trace().Msg("Exiting OptIntoSlashing function...")

	// assemble tx
	tx, err := dlcc.Bindings.DlReg.UpdateSocket(dlcc.NoSendTransactOpts, socket)
	if err != nil {
		log.Error().Err(err).Msg("Error assembling UpdateSocket tx")
		return err
	}
	// estimate gas and send tx
	_, err = dlcc.EstimateGasPriceAndLimitAndSendTx(context.Background(), tx, "UpdateSocket")
	return err
}

func (dlcc *DataLayrChainClient) RegisterBLSPublicKey(ctx context.Context, keypair *bls.BlsKeyPair, pubkeyG1Param contractBLSPublicKeyCompendium.BN254G1Point) error {
	log := dlcc.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering RegisterBLSPublicKey function...")
	defer log.Trace().Msg("Exiting RegisterBLSPublicKey function...")

	//first register the public key with the compendium
	pkCompendiumAddr, err := dlcc.Bindings.DlReg.PubkeyCompendium(&bind.CallOpts{})
	if err != nil {
		log.Error().Err(err).Msg("error getting pkCompendium")
		return err
	}

	pkCompendium, err := contractBLSPublicKeyCompendium.NewContractBLSPublicKeyCompendium(pkCompendiumAddr, dlcc.ChainClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve pkComp instance")
		return err
	}

	pkh, err := pkCompendium.OperatorToPubkeyHash(&bind.CallOpts{}, dlcc.AccountAddress)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve bls pubkey registered status from chain")
		return err
	}

	s, R, pubkeyG2 := keypair.MakeRegistrationData(dlcc.AccountAddress)

	RParam := contractBLSPublicKeyCompendium.BN254G1Point{X: R.X, Y: R.Y}
	pubkeyG2Param := contractBLSPublicKeyCompendium.BN254G2Point{X: pubkeyG2.X, Y: pubkeyG2.Y}

	//if no pubkey registered already, then register
	if pkh == emptyBytes {
		// assemble tx
		tx, err := pkCompendium.RegisterBLSPublicKey(dlcc.NoSendTransactOpts, s, RParam, pubkeyG1Param, pubkeyG2Param)
		if err != nil {
			log.Error().Err(err).Msg("Error assembling RegisterBLSPublicKey tx")
			return err
		}
		// estimate gas and send tx
		_, err = dlcc.EstimateGasPriceAndLimitAndSendTx(ctx, tx, "RegisterBLSPubkey")

		if err != nil {
			log.Error().Err(err).Msg("Failed to register pubkey")
			return err
		}
	}
	return nil
}

// UpdateStakes allows anyone to update the stakes for a given operator
func (dlcc *DataLayrChainClient) UpdateStakes(ctx context.Context, operators []common.Address) error {
	log := dlcc.Logger.SubloggerId(ctx)
	log.Trace().Msg("Entering UpdateStakes function...")
	defer log.Trace().Msg("Exiting UpdateStakes function...")

	insertAfters := make([]*big.Int, len(operators))

	// assemble tx
	tx, err := dlcc.Bindings.DlReg.UpdateStakes(dlcc.NoSendTransactOpts, operators, insertAfters)
	if err != nil {
		log.Error().Err(err).Msg("Error assembling UpdateStakes tx")
		return err
	}
	// estimate gas and send tx

	_, err = dlcc.EstimateGasPriceAndLimitAndSendTx(ctx, tx, "UpdateStakes")
	return err
}
