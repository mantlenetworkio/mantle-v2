package addresses

import "github.com/ethereum/go-ethereum/common"

type MantleImplContracts struct {
	AllSetUp bool

	// Mantle uses legacy implementations which could not be shared between chains.
	OptimismPortalImpl               common.Address
	SystemConfigImpl                 common.Address
	L1CrossDomainMessengerImpl       common.Address // relies on OptimismPortalProxy
	L1Erc721BridgeImpl               common.Address
	L1StandardBridgeImpl             common.Address
	OptimismMintableErc20FactoryImpl common.Address
	L2OutputOracleImpl               common.Address
}
