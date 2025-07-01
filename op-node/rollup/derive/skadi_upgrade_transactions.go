package derive

import (
	"math/big"

	"github.com/ethereum-optimism/optimism/op-service/predeploys"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

var (
	// EIP-4788
	beaconBlockRootDeployerSource     = UpgradeDepositSource{Intent: "Skadi: EIP-4788 Contract Deployment"}
	beaconBlockRootDeploymentBytecode = common.FromHex("0x60618060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604d57602036146024575f5ffd5b5f35801560495762001fff810690815414603c575f5ffd5b62001fff01545f5260205ff35b5f5ffd5b62001fff42064281555f359062001fff015500")

	// EIP-2935
	blockHashDeployerSource     = UpgradeDepositSource{Intent: "Skadi: EIP-2935 Contract Deployment"}
	blockHashDeploymentBytecode = common.FromHex("0x60538060095f395ff33373fffffffffffffffffffffffffffffffffffffffe14604657602036036042575f35600143038111604257611fff81430311604257611fff9006545f5260205ff35b5f5ffd5b5f35611fff60014303065500")
)

func MantleSkadiNetworkUpgradeTransactions() ([]hexutil.Bytes, error) {
	upgradeTxns := make([]hexutil.Bytes, 0, 2)

	deployBeaconBlockRootsContract, err := types.NewTx(&types.DepositTx{
		SourceHash:          beaconBlockRootDeployerSource.SourceHash(),
		From:                predeploys.EIP4788ContractDeployer,
		To:                  nil,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 250_000,
		IsSystemTransaction: false,
		Data:                beaconBlockRootDeploymentBytecode,
	}).MarshalBinary()
	if err != nil {
		return nil, err
	}
	upgradeTxns = append(upgradeTxns, deployBeaconBlockRootsContract)

	deployHistoricalBlockHashesContract, err := types.NewTx(&types.DepositTx{
		SourceHash:          blockHashDeployerSource.SourceHash(),
		From:                predeploys.EIP2935ContractDeployer,
		To:                  nil,
		Mint:                big.NewInt(0),
		Value:               big.NewInt(0),
		Gas:                 250_000,
		IsSystemTransaction: false,
		Data:                blockHashDeploymentBytecode,
	}).MarshalBinary()
	if err != nil {
		return nil, err
	}
	upgradeTxns = append(upgradeTxns, deployHistoricalBlockHashesContract)

	return upgradeTxns, nil
}
