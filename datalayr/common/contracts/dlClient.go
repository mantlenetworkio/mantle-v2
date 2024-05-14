package contracts

import (
	"github.com/Layr-Labs/datalayr/common/chain"
	"github.com/Layr-Labs/datalayr/common/logging"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"

	rg "github.com/Layr-Labs/datalayr/common/contracts/bindings/BLSRegistry"
	sm "github.com/Layr-Labs/datalayr/common/contracts/bindings/DataLayrServiceManager"
)

type DataLayrChainClient struct {
	*chain.Client
	Logger   *logging.Logger
	Bindings *ContractBindings
}

type ContractBindings struct {
	DlSm     sm.ContractDataLayrServiceManager
	DlSmAddr common.Address
	DlReg    rg.ContractBLSRegistry
}

func NewDataLayrChainClient(config chain.ClientConfig, dlSmAddress string, logger *logging.Logger) (*DataLayrChainClient, error) {
	c, err := chain.NewClient(config, logger)
	if err != nil {
		logger.Error().Err(err).Msg("Cannot create chain.Client")
		return nil, err
	}

	dlSmAddressCast := common.HexToAddress(dlSmAddress)

	dlcc := &DataLayrChainClient{
		Client: c,
		Logger: logger,
	}

	err = dlcc.updateContractBindings(dlSmAddressCast)

	return dlcc, err
}

func (dlcc *DataLayrChainClient) updateContractBindings(dlSmAddr common.Address) error {
	dlSm, err := sm.NewContractDataLayrServiceManager(dlSmAddr, dlcc.ChainClient)
	if err != nil {
		dlcc.Logger.Error().Err(err).Msg("Failed to retrieve dlSm instance")
		return err
	}

	dlRegAddr, err := dlSm.Registry(&bind.CallOpts{})
	if err != nil {
		dlcc.Logger.Error().Err(err).Msg("Failed to retrieve DlReg address from DlSm")
		return err
	}
	dlReg, err := rg.NewContractBLSRegistry(dlRegAddr, dlcc.ChainClient)
	if err != nil {
		dlcc.Logger.Error().Err(err).Msg("Failed to retrieve DlReg instance")
		return err
	}

	Bindings := &ContractBindings{
		DlSm:     *dlSm,
		DlSmAddr: dlSmAddr,
		DlReg:    *dlReg,
	}

	dlcc.Bindings = Bindings
	return nil
}
