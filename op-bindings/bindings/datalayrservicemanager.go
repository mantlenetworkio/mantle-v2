// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package bindings

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// BLSSignatureCheckerSignatoryTotals is an auto generated low-level Go binding around an user-defined struct.
type BLSSignatureCheckerSignatoryTotals struct {
	SignedStakeFirstQuorum  *big.Int
	SignedStakeSecondQuorum *big.Int
	TotalStakeFirstQuorum   *big.Int
	TotalStakeSecondQuorum  *big.Int
}

// IDataLayrServiceManagerDataStoreMetadata is an auto generated low-level Go binding around an user-defined struct.
type IDataLayrServiceManagerDataStoreMetadata struct {
	HeaderHash           [32]byte
	DurationDataStoreId  uint32
	GlobalDataStoreId    uint32
	ReferenceBlockNumber uint32
	BlockNumber          uint32
	Fee                  *big.Int
	Confirmer            common.Address
	SignatoryRecordHash  [32]byte
}

// IDataLayrServiceManagerDataStoreSearchData is an auto generated low-level Go binding around an user-defined struct.
type IDataLayrServiceManagerDataStoreSearchData struct {
	Metadata  IDataLayrServiceManagerDataStoreMetadata
	Duration  uint8
	Timestamp *big.Int
	Index     uint32
}

// ContractDataLayrServiceManagerMetaData contains all meta data concerning the ContractDataLayrServiceManager contract.
var ContractDataLayrServiceManagerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"contractIQuorumRegistry\",\"name\":\"_registry\",\"type\":\"address\"},{\"internalType\":\"contractIInvestmentManager\",\"name\":\"_investmentManager\",\"type\":\"address\"},{\"internalType\":\"contractIEigenLayrDelegation\",\"name\":\"_eigenLayrDelegation\",\"type\":\"address\"},{\"internalType\":\"contractIERC20\",\"name\":\"_collateralToken\",\"type\":\"address\"},{\"internalType\":\"contractDataLayrChallenge\",\"name\":\"_dataLayrChallenge\",\"type\":\"address\"},{\"internalType\":\"contractDataLayrBombVerifier\",\"name\":\"_dataLayrBombVerifier\",\"type\":\"address\"},{\"internalType\":\"contractIRegistryPermission\",\"name\":\"_dataPermissionManager\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"adversaryThresholdBasisPoints\",\"type\":\"uint16\"}],\"name\":\"AdversaryThresholdBasisPointsUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"BombVerifierSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"dataStoreId\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"headerHash\",\"type\":\"bytes32\"}],\"name\":\"ConfirmDataStore\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"previousValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newValue\",\"type\":\"uint256\"}],\"name\":\"FeePerBytePerTimeSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"FeeSetterChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"feePayer\",\"type\":\"address\"},{\"components\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"headerHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"durationDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"globalDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint96\",\"name\":\"fee\",\"type\":\"uint96\"},{\"internalType\":\"address\",\"name\":\"confirmer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"signatoryRecordHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIDataLayrServiceManager.DataStoreMetadata\",\"name\":\"metadata\",\"type\":\"tuple\"},{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"indexed\":false,\"internalType\":\"structIDataLayrServiceManager.DataStoreSearchData\",\"name\":\"searchData\",\"type\":\"tuple\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"header\",\"type\":\"bytes\"}],\"name\":\"InitDataStore\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"Paused\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address\",\"name\":\"previousAddress\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"newAddress\",\"type\":\"address\"}],\"name\":\"PaymentManagerSet\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint16\",\"name\":\"quorumTHresholdBasisPoints\",\"type\":\"uint16\"}],\"name\":\"QuorumThresholdBasisPointsUpdated\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"},{\"indexed\":false,\"internalType\":\"uint32\",\"name\":\"taskNumber\",\"type\":\"uint32\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"signedStakeFirstQuorum\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"signedStakeSecondQuorum\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"bytes32[]\",\"name\":\"pubkeyHashes\",\"type\":\"bytes32[]\"}],\"name\":\"SignatoryRecord\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"Unpaused\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BIP_MULTIPLIER\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BLOCK_STALE_MEASURE\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DURATION_SCALE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_DATASTORE_DURATION\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_DATASTORE_DURATION\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NUM_DS_PER_BLOCK_PER_DURATION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"adversaryThresholdBasisPoints\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"totalBytes\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_feePerBytePerTime\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"storePeriodLength\",\"type\":\"uint32\"}],\"name\":\"calculateFee\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"}],\"name\":\"checkSignatures\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"taskNumberToConfirm\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"},{\"components\":[{\"internalType\":\"uint256\",\"name\":\"signedStakeFirstQuorum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"signedStakeSecondQuorum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalStakeFirstQuorum\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"totalStakeSecondQuorum\",\"type\":\"uint256\"}],\"internalType\":\"structBLSSignatureChecker.SignatoryTotals\",\"name\":\"signedTotals\",\"type\":\"tuple\"},{\"internalType\":\"bytes32\",\"name\":\"compressedSignatoryRecord\",\"type\":\"bytes32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"collateralToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"headerHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"durationDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"globalDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint96\",\"name\":\"fee\",\"type\":\"uint96\"},{\"internalType\":\"address\",\"name\":\"confirmer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"signatoryRecordHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIDataLayrServiceManager.DataStoreMetadata\",\"name\":\"metadata\",\"type\":\"tuple\"},{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"internalType\":\"structIDataLayrServiceManager.DataStoreSearchData\",\"name\":\"searchData\",\"type\":\"tuple\"}],\"name\":\"confirmDataStore\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"confirmDataStoreTimeout\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dataLayrBombVerifier\",\"outputs\":[{\"internalType\":\"contractDataLayrBombVerifier\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dataLayrChallenge\",\"outputs\":[{\"internalType\":\"contractDataLayrChallenge\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dataPermissionManager\",\"outputs\":[{\"internalType\":\"contractIRegistryPermission\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"dataStoreHashesForDurationAtTimestamp\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"name\":\"dataStoreIdToSignatureHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"dataStoresForDuration\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"one_duration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"two_duration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"three_duration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"four_duration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"five_duration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"six_duration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"seven_duration\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"dataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"latestTime\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eigenLayrDelegation\",\"outputs\":[{\"internalType\":\"contractIEigenLayrDelegation\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feePerBytePerTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feeSetter\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"freezeOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"getDataStoreHashesForDurationAtTimestamp\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"}],\"name\":\"getNumDataStoresForDuration\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"feePayer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"confirmer\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalOperatorsIndex\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"header\",\"type\":\"bytes\"}],\"name\":\"initDataStore\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"contractIPauserRegistry\",\"name\":\"_pauserRegistry\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"initialOwner\",\"type\":\"address\"},{\"internalType\":\"uint16\",\"name\":\"_quorumThresholdBasisPoints\",\"type\":\"uint16\"},{\"internalType\":\"uint16\",\"name\":\"_adversaryThresholdBasisPoints\",\"type\":\"uint16\"},{\"internalType\":\"uint256\",\"name\":\"_feePerBytePerTime\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_feeSetter\",\"type\":\"address\"}],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"investmentManager\",\"outputs\":[{\"internalType\":\"contractIInvestmentManager\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestTime\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"log2NumPowersOfTau\",\"outputs\":[{\"internalType\":\"uint48\",\"name\":\"\",\"type\":\"uint48\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"numPowersOfTau\",\"outputs\":[{\"internalType\":\"uint48\",\"name\":\"\",\"type\":\"uint48\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"pause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"index\",\"type\":\"uint8\"}],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"paused\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"pauserRegistry\",\"outputs\":[{\"internalType\":\"contractIPauserRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"quorumThresholdBasisPoints\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"registry\",\"outputs\":[{\"internalType\":\"contractIQuorumRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"_adversaryThresholdBasisPoints\",\"type\":\"uint16\"}],\"name\":\"setAdversaryThresholdBasisPoints\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_feePerBytePerTime\",\"type\":\"uint256\"}],\"name\":\"setFeePerBytePerTime\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_feeSetter\",\"type\":\"address\"}],\"name\":\"setFeeSetter\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint16\",\"name\":\"_quorumThresholdBasisPoints\",\"type\":\"uint16\"}],\"name\":\"setQuorumThresholdBasisPoints\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"taskNumber\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"newPausedStatus\",\"type\":\"uint256\"}],\"name\":\"unpause\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"headerHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"durationDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"globalDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint96\",\"name\":\"fee\",\"type\":\"uint96\"},{\"internalType\":\"address\",\"name\":\"confirmer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"signatoryRecordHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIDataLayrServiceManager.DataStoreMetadata\",\"name\":\"metadata\",\"type\":\"tuple\"}],\"name\":\"verifyDataStoreMetadata\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"zeroPolynomialCommitmentMerkleRoots\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6101606040523480156200001257600080fd5b50604051620043683803806200436883398101604081905262000035916200015b565b6001600160a01b0380881660805286811660e05285811660c05284811660a05283811661012052828116610140528116610100526200007362000080565b5050505050505062000206565b600054610100900460ff1615620000ed5760405162461bcd60e51b815260206004820152602760248201527f496e697469616c697a61626c653a20636f6e747261637420697320696e697469604482015266616c697a696e6760c81b606482015260840160405180910390fd5b60005460ff908116101562000140576000805460ff191660ff9081179091556040519081527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b565b6001600160a01b03811681146200015857600080fd5b50565b600080600080600080600060e0888a0312156200017757600080fd5b8751620001848162000142565b6020890151909750620001978162000142565b6040890151909650620001aa8162000142565b6060890151909550620001bd8162000142565b6080890151909450620001d08162000142565b60a0890151909350620001e38162000142565b60c0890151909250620001f68162000142565b8091505092959891949750929550565b60805160a05160c05160e0516101005161012051610140516140cf62000299600039600081816105590152610ad30152600081816104dc0152610aa10152600081816105fa01528181610d0701526118d201526000610484015260006103f80152600061068c0152600081816105d301528181611a50015281816120730152818161222d015261242e01526140cf6000f3fe608060405234801561001057600080fd5b50600436106102955760003560e01c8063715018a611610167578063a50017a1116100ce578063dcf49ea711610087578063dcf49ea7146106d8578063deaf4498146106eb578063ed82c0ee14610752578063f2fde38b14610765578063fabc1cbc14610778578063fc2c60581461078b57600080fd5b8063a50017a11461029a578063b19805af14610674578063b2016bd414610687578063b569157b146106ae578063ba4994b1146106bc578063d21eed4f146106cf57600080fd5b80637bc9d56e116101205780637bc9d56e1461061c5780637dfd16d71461062f57806387cf3ef41461063d578063886f1195146106505780638da5cb5b14610663578063a3c7eaf01461066b57600080fd5b8063715018a61461058b57806372d18e8d1461059357806373441c4e146105a8578063772eefe3146105bb5780637b103999146105ce5780637b49d55f146105f557600080fd5b806339fe2e711161020b57806358942e73116101c457806358942e73146105065780635ac86ab7146105195780635c975abb1461054c5780635e69019c146105545780635e8b3f2d1461057b5780635f87abbb1461058357600080fd5b806339fe2e711461046c5780634b31bb101461047f5780634d53cebb146104a6578063516d8616146104c4578063573792c6146104d7578063578ae5a1146104fe57600080fd5b806331a219c51161025d57806331a219c51461032157806333223aea1461032a5780633367a3fb146103e057806333d2433a146103f35780633594b60f1461043257806338c8ee641461045957600080fd5b8063046bf4a61461029a57806307dab8b3146102be578063136439dd146102d35780631bd2b3cf146102e65780631fdab6e414610307575b600080fd5b6102a2600081565b60405165ffffffffffff90911681526020015b60405180910390f35b6102d16102cc3660046135ad565b6107ab565b005b6102d16102e13660046135cf565b6108ae565b6102f96102f43660046135f9565b610a4d565b6040519081526020016102b5565b61030f600181565b60405160ff90911681526020016102b5565b6102f9610e1081565b608b54608c546103899163ffffffff808216926401000000008304821692600160401b8104831692600160601b8204811692600160801b8304821692600160a01b8104831692600160c01b8204811692600160e01b9092048116911689565b6040805163ffffffff9a8b168152988a1660208a01529689169688019690965293871660608701529186166080860152851660a0850152841660c0840152831660e0830152909116610100820152610120016102b5565b6102f96103ee3660046135cf565b610a7f565b61041a7f000000000000000000000000000000000000000000000000000000000000000081565b6040516001600160a01b0390911681526020016102b5565b608a546104469062010000900461ffff1681565b60405161ffff90911681526020016102b5565b6102d161046736600461364c565b610a96565b6102f961047a366004613686565b610b86565b61041a7f000000000000000000000000000000000000000000000000000000000000000081565b6104af61070881565b60405163ffffffff90911681526020016102b5565b6102d16104d23660046135ad565b610bab565b61041a7f000000000000000000000000000000000000000000000000000000000000000081565b61030f600781565b6102d161051436600461380b565b610c9d565b61053c6105273660046138d9565b608954600160ff9092169190911b9081161490565b60405190151581526020016102b5565b6089546102f9565b61041a7f000000000000000000000000000000000000000000000000000000000000000081565b6104af609681565b6102f9601481565b6102d1611446565b608b54600160e01b900463ffffffff166104af565b6104af6105b63660046138d9565b61145a565b6102d16105c93660046135cf565b6115b9565b61041a7f000000000000000000000000000000000000000000000000000000000000000081565b61041a7f000000000000000000000000000000000000000000000000000000000000000081565b6102d161062a3660046138f4565b61160c565b608c5463ffffffff166104af565b608d5461041a906001600160a01b031681565b60885461041a906001600160a01b031681565b61041a61181e565b6102f961271081565b6102d161068236600461364c565b611837565b61041a7f000000000000000000000000000000000000000000000000000000000000000081565b608a546104469061ffff1681565b61053c6106ca366004613969565b611848565b6102f960655481565b6104af6106e63660046139ba565b611868565b6106fe6106f9366004613a59565b612010565b6040805163ffffffff9687168152959094166020808701919091528585019390935281516060808701919091529282015160808601529281015160a0850152015160c083015260e0820152610100016102b5565b6102f9610760366004613a9b565b612bb6565b6102d161077336600461364c565b612bf4565b6102d16107863660046135cf565b612c6a565b6102f9610799366004613ad0565b60866020526000908152604090205481565b6107b3612e09565b608a54819061ffff620100009091048116906001908316108015906107de575061270f61ffff831611155b6108035760405162461bcd60e51b81526004016107fa90613aed565b60405180910390fd5b600161ffff82161080159061081e575061270f61ffff821611155b61083a5760405162461bcd60e51b81526004016107fa90613b5e565b8061ffff168261ffff16116108615760405162461bcd60e51b81526004016107fa90613bd2565b608a805461ffff191661ffff85169081179091556040519081527fae5844e5ca560c940e41aae83424a548a030c790cd14ae00d68c8437bb2e8ec2906020015b60405180910390a1505050565b608860009054906101000a90046001600160a01b03166001600160a01b0316639fd0506d6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610901573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906109259190613c56565b6001600160a01b0316336001600160a01b0316146109965760405162461bcd60e51b815260206004820152602860248201527f6d73672e73656e646572206973206e6f74207065726d697373696f6e6564206160448201526739903830bab9b2b960c11b60648201526084016107fa565b60895481811614610a0f5760405162461bcd60e51b815260206004820152603860248201527f5061757361626c652e70617573653a20696e76616c696420617474656d70742060448201527f746f20756e70617573652066756e6374696f6e616c697479000000000000000060648201526084016107fa565b608981905560405181815233907fab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d906020015b60405180910390a250565b60876020528260005260406000206020528160005260406000208160148110610a7557600080fd5b0154925083915050565b60668160208110610a8f57600080fd5b0154905081565b336001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000161480610af55750336001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016145b610b835760405162461bcd60e51b815260206004820152605360248201527f446174614c617972536572766963654d616e616765722e667265657a654f706560448201527f7261746f723a204f6e6c79206368616c6c656e6765207265736f6c766572732060648201527263616e20736c617368206f70657261746f727360681b608482015260a4016107fa565b50565b600063ffffffff8216610b998486613c89565b610ba39190613c89565b949350505050565b610bb3612e09565b608a5461ffff168160018210801590610bd2575061270f61ffff831611155b610bee5760405162461bcd60e51b81526004016107fa90613aed565b600161ffff821610801590610c09575061270f61ffff821611155b610c255760405162461bcd60e51b81526004016107fa90613b5e565b8061ffff168261ffff1611610c4c5760405162461bcd60e51b81526004016107fa90613bd2565b608a805463ffff000019166201000061ffff8681168202929092179283905560405192041681527f1bdc513ac13a36cd49087fef52b034cb5833bd75154db5239f27daa6bde17042906020016108a1565b60895460019060029081161415610cf25760405162461bcd60e51b815260206004820152601960248201527814185d5cd8589b194e881a5b99195e081a5cc81c185d5cd959603a1b60448201526064016107fa565b6040516309d45bfd60e11b81523360048201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906313a8b7fa90602401602060405180830381865afa158015610d56573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190610d7a9190613ca8565b1515600114610dfd5760405162461bcd60e51b8152602060048201526055602482015260008051602061405a83398151915260448201527f746153746f72653a2074686973206164647265737320686173206e6f74207065606482015274726d697373696f6e20636f6e6669726d206461746160581b608482015260a4016107fa565b6000806000806000610e0f8989612010565b9450945094509450945084876000015160000151886020015189604001518a60600151604051602001610e8895949392919060e095861b6001600160e01b03199081168252600482019590955260f89390931b6001600160f81b0319166024840152602583019190915290921b16604582015260490190565b604051602081830303815290604052805190602001208314610f1c5760405162461bcd60e51b8152602060048201526053602482015260008051602061405a83398151915260448201527f746153746f72653a206d736748617368206973206e6f7420636f6e73697374656064820152726e74207769746820736561726368206461746160681b608482015260a4016107fa565b6040870151610f2e9061070890613cca565b4210610fa05760405162461bcd60e51b8152602060048201526047602482015260008051602061405a83398151915260448201527f746153746f72653a20436f6e6669726d6174696f6e2077696e646f7720686173606482015266081c185cdcd95960ca1b608482015260a4016107fa565b865160c001516001600160a01b031633146110375760405162461bcd60e51b815260206004820152605b602482015260008051602061405a83398151915260448201527f746153746f72653a2053656e646572206973206e6f7420617574686f72697a6560648201527f6420746f20636f6e6669726d2074686973206461746173746f72650000000000608482015260a4016107fa565b865160e00151156110b25760405162461bcd60e51b815260206004820152604b602482015260008051602061405a83398151915260448201527f746153746f72653a205369676e61746f72795265636f7264206d75737420626560648201526a206279746573333228302960a81b608482015260a4016107fa565b8463ffffffff1687600001516040015163ffffffff16146111485760405162461bcd60e51b8152602060048201526056602482015260008051602061405a83398151915260448201527f746153746f72653a20676c6f61626c6461746173746f7265696420697320646f6064820152756573206e6f742061677265652077697468206461746160501b608482015260a4016107fa565b8363ffffffff1687600001516060015163ffffffff16146111d55760405162461bcd60e51b815260206004820152604d602482015260008051602061405a83398151915260448201527f746153746f72653a20626c6f636b6e756d62657220646f6573206e6f7420616760648201526c7265652077697468206461746160981b608482015260a4016107fa565b60006111e48860000151612e68565b6020808a015160ff166000908152608782526040808220818d015183529092522060608a0151919250829163ffffffff166014811061122557611225613ce2565b0154146112cd5760405162461bcd60e51b8152602060048201526076602482015260008051602061405a83398151915260448201527f746153746f72653a2070726f76696465642063616c6c6461746120646f65732060648201527f6e6f74206d6174636820636f72726573706f6e64696e672073746f72656420686084820152756173682066726f6d20696e69744461746153746f726560501b60a482015260c4016107fa565b875160e00182905287516000906112e390612e68565b6020808b015160ff166000908152608782526040808220818e015183529092522060608b0151919250829163ffffffff166014811061132457611324613ce2565b0155608a546040850151855161ffff909216916113449061271090613c89565b61134e9190613d0e565b10156113ec5760405162461bcd60e51b815260206004820152606d602482015260008051602061405a83398151915260448201527f746153746f72653a207369676e61746f7269657320646f206e6f74206f776e2060648201527f6174206c65617374207468726573686f6c642070657263656e74616765206f6660848201526c20626f74682071756f72756d7360981b60a482015260c4016107fa565b608b5489515160408051600160e01b90930463ffffffff16835260208301919091527ffbb7f4f1b0b9ad9e75d69d22c364e13089418d86fcb5106792a53046c0fb33aa910160405180910390a15050505050505050505050565b61144e612e09565b6114586000612f34565b565b60008160ff1660011415611476575050608b5463ffffffff1690565b8160ff1660021415611498575050608b54640100000000900463ffffffff1690565b8160ff16600314156114b9575050608b54600160401b900463ffffffff1690565b8160ff16600414156114da575050608b54600160601b900463ffffffff1690565b8160ff16600514156114fb575050608b54600160801b900463ffffffff1690565b8160ff166006141561151c575050608b54600160a01b900463ffffffff1690565b8160ff166007141561153d575050608b54600160c01b900463ffffffff1690565b60405162461bcd60e51b8152602060048201526044602482018190527f446174614c617972536572766963654d616e616765722e6765744e756d446174908201527f6153746f726573466f724475726174696f6e3a20696e76616c696420647572616064820152633a34b7b760e11b608482015260a4016107fa565b608d546001600160a01b031633146116035760405162461bcd60e51b815260206004820152600d60248201526c37b7363ca332b2a9b2ba3a32b960991b60448201526064016107fa565b610b8381612f86565b600054610100900460ff161580801561162c5750600054600160ff909116105b806116465750303b158015611646575060005460ff166001145b6116a95760405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b60648201526084016107fa565b6000805460ff1916600117905580156116cc576000805461ff0019166101001790555b8484600161ffff8316108015906116e9575061270f61ffff831611155b6117055760405162461bcd60e51b81526004016107fa90613aed565b600161ffff821610801590611720575061270f61ffff821611155b61173c5760405162461bcd60e51b81526004016107fa90613b5e565b8061ffff168261ffff16116117635760405162461bcd60e51b81526004016107fa90613bd2565b61176e896000612fc7565b61177788612f34565b608b80546001600160e01b0316600160e01b179055608c805463ffffffff191660011790556117a585612f86565b6117ae846130c7565b5050608a805461ffff868116620100000263ffffffff19909216908816171790558015611815576000805461ff0019169055604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b50505050505050565b60006118326033546001600160a01b031690565b905090565b61183f612e09565b610b83816130c7565b600061185382612e68565b61185e868686612bb6565b1495945050505050565b608954600090600190811614156118bd5760405162461bcd60e51b815260206004820152601960248201527814185d5cd8589b194e881a5b99195e081a5cc81c185d5cd959603a1b60448201526064016107fa565b6040516309d45bfd60e11b81523360048201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b0316906313a8b7fa90602401602060405180830381865afa158015611921573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906119459190613ca8565b15156001146119c55760405162461bcd60e51b8152602060048201526052602482015260008051602061403a83398151915260448201527f746f72653a2074686973206164647265737320686173206e6f74207065726d696064820152717373696f6e20746f2070757368206461746160701b608482015260a4016107fa565b600083836040516119d7929190613d22565b6040805191829003822061010083018252600080845260208401819052918301829052606083018290526080830182905260a0830182905260c0830182905260e0830182905292509060405163b322bdc760e01b815263ffffffff808a1660048301528816602482015260009081906001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000169063b322bdc790604401602060405180830381865afa158015611a97573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190611abb9190613d32565b9050611ac8888883613130565b91506020821015611b395760405162461bcd60e51b8152602060048201526041602482015260008051602061403a83398151915260448201527f746f72653a20746f74616c4279746573203c204d494e5f53544f52455f53495a6064820152604560f81b608482015260a4016107fa565b63ee6b2800821115611bab5760405162461bcd60e51b8152602060048201526041602482015260008051602061403a83398151915260448201527f746f72653a20746f74616c4279746573203e204d41585f53544f52455f53495a6064820152604560f81b608482015260a4016107fa565b5060018a60ff1610158015611bc45750600760ff8b1611155b611c1d5760405162461bcd60e51b8152602060048201526036602482015260008051602061403a8339815191526044820152753a37b9329d1024b73b30b634b210323ab930ba34b7b760511b60648201526084016107fa565b611c2c610e1060ff8c16613c89565b92506000611c3d8260655486610b86565b9050604051806101000160405280868152602001611c5a8d61145a565b63ffffffff9081168252608b54600160e01b9004811660208301528c81166040830152431660608201526001600160601b0390921660808301526001600160a01b038d1660a0830152600060c0909201919091529150505b60148463ffffffff161015611d515760ff89166000908152608760209081526040808320428452909152902063ffffffff851660148110611cf557611cf5613ce2565b0154611d3f57611d0481612e68565b60ff8a166000908152608760209081526040808320428452909152902063ffffffff861660148110611d3857611d38613ce2565b0155611d51565b83611d4981613d4f565b945050611cb2565b60148463ffffffff161415611dfb5760405162461bcd60e51b8152602060048201526070602482015260008051602061403a83398151915260448201527f746f72653a206e756d626572206f6620696e69744461746173746f726573206660648201527f6f722074686973206475726174696f6e20616e6420626c6f636b20686173207260848201526f195858da1959081a5d1cc81b1a5b5a5d60821b60a482015260c4016107fa565b438863ffffffff161115611e7f5760405162461bcd60e51b8152602060048201526051602482015260008051602061403a83398151915260448201527f746f72653a20737065636966696564207265666572656e6365426c6f636b4e756064820152706d62657220697320696e2066757475726560781b608482015260a4016107fa565b63ffffffff4316611e9160968a613d73565b63ffffffff161015611f1f5760405162461bcd60e51b8152602060048201526057602482015260008051602061403a83398151915260448201527f746f72653a20737065636966696564207265666572656e6365426c6f636b4e7560648201527f6d62657220697320746f6f2066617220696e2070617374000000000000000000608482015260a4016107fa565b6040805160808101825282815260ff8b166020820152428183015263ffffffff8616606082015290517f25a833fbdbcbd72479fd89b970eb4715d9718313f196da0477220e8cd44425d890611f7b908e9084908b908b90613dc4565b60405180910390a16000611f8f8442613d73565b608c5490915063ffffffff9081169082161115611fbc57608c805463ffffffff191663ffffffff83161790555b611fc58b613153565b608b8054601c90611fe290600160e01b900463ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff1602179055505050505050979650505050505050565b60008060006120406040518060800160405280600081526020016000815260200160008152602001600081525090565b60405163102068d360e01b8152602087013560d01c60048201819052602688013560e01c945087359350600091889083907f00000000000000000000000000000000000000000000000000000000000000006001600160a01b03169063102068d390602401608060405180830381865afa1580156120c2573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906120e69190613e9f565b90506120f28188613347565b604080820180516001600160601b039081169288019290925251811686526060808301805183169188019190915251166020860152602a82013560e090811c9850602e830135901c9250612147603283613cca565b915060008367ffffffffffffffff81111561216457612164613708565b60405190808252806020026020018201604052801561218d578160200160208202803683370190505b50905061219861353b565b600085156122f65784358083526020808701358185018190526040805192830193909352818301526044870196919091013560e01c90600090606001604051602081830303815290604052805190602001209050808560008151811061220057612200613ce2565b6020908102919091010152604051632173fa2360e01b81526004810182905263ffffffff831660248201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690632173fa2390604401608060405180830381865afa15801561227c573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906122a09190613e9f565b95506122ac868d613347565b85604001516001600160601b03168a6000018181516122cb9190613f32565b905250606086015160208b0180516001600160601b03909216916122f0908390613f32565b90525050505b60015b8681101561257e57853560408481018290526020808901356060808801829052835180840195909552848401919091528251808503840181529301825282519201919091206044880197919091013560e01c9085612358600185613f32565b8151811061236857612368613ce2565b602002602001015160001c8160001c116123ee5760405162461bcd60e51b815260206004820152604d602482015260008051602061407a83398151915260448201527f7265733a205075626b657920686173686573206d75737420626520696e20617360648201526c31b2b73234b7339037b93232b960991b608482015260a4016107fa565b8086848151811061240157612401613ce2565b6020908102919091010152604051632173fa2360e01b81526004810182905263ffffffff831660248201527f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031690632173fa2390604401608060405180830381865afa15801561247d573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906124a19190613e9f565b96506124ad878e613347565b86604001516001600160601b03168b6000018181516124cc9190613f32565b905250606087015160208c0180516001600160601b03909216916124f1908390613f32565b90525060408560808160066107d05a03fa935083801561251057612512565bfe5b50836125745760405162461bcd60e51b815260206004820152603f602482015260008051602061407a83398151915260448201527f7265733a206e6f6e207369676e6572206164646974696f6e206661696c65640060648201526084016107fa565b50506001016122f9565b506004850135604083015260248501356060830152604490940193851561265257602082015160008051602061401a833981519152906125be9082613f32565b6125c89190613f49565b6020830152604080830160808460066107d05a03fa9050806126525760405162461bcd60e51b8152602060048201526049602482015260008051602061407a83398151915260448201527f7265733a20616767726567617465206e6f6e207369676e6572206164646974696064820152681bdb8819985a5b195960ba1b608482015260a4016107fa565b61265b896133af565b836006602002018460076020908102919091019290925291909152853561012084015285810135610100840152604086013561016084015260608601356101408401526080860135835260a08601359083015260c0909401938160006020020151826001602002015183600260200201518460036020020151856006602002015186600760200201518760086020020151886009602002015189600a60200201518a600b60209081029190910151604080519283019b909b52998101989098526060880196909652608087019490945260a086019290925260c085015260e08401526101008301526101208201526101408101919091526101600160408051601f1981840301815291905280516020909101208260046020020152604082810160608160076107d05a03fa9050806128065760405162461bcd60e51b8152602060048201526054602482015260008051602061407a83398151915260448201527f7265733a20616767726567617465207369676e6572207075626c6963206b6579606482015273081c985b991bdb481cda1a599d0819985a5b195960621b608482015260a4016107fa565b60408260808460066107d05a03fa90508061289d5760405162461bcd60e51b815260206004820152605e602482015260008051602061407a83398151915260448201527f7265733a20616767726567617465207369676e6572207075626c6963206b657960648201527f20616e64207369676e6174757265206164646974696f6e206661696c65640000608482015260a4016107fa565b6001826002602002015260028260036020020152604060808301606084830160076107d05a03fa9050806129325760405162461bcd60e51b8152602060048201526042602482015260008051602061407a83398151915260448201527f7265733a2067656e657261746f722072616e646f6d207368696674206661696c606482015261195960f21b608482015260a4016107fa565b604060c08301608080850160066107d05a03fa9050806129ce5760405162461bcd60e51b8152602060048201526057602482015260008051602061407a83398151915260448201527f7265733a2067656e657261746f722072616e646f6d20736869667420616e642060648201527f47312068617368206164646974696f6e206661696c6564000000000000000000608482015260a4016107fa565b7f198e9393920d483a7260bfb731fb5d25f1aa493335a9e71297e485b7aef312c260408301527f1800deef121f1e76426a00665e5c4479674322d4f75edadd46debd5cd992f6ed60608301527f275dc4a288d1afb3cbb1ac09187524c7db36395df7be3b99e673b13a075a65ec60808301527f1d9befcd05a5323e6da4d435f3b617cdb3af83285c2df711ef39c01571827f9d60a08301526020826101808160086107d05a03fa905080612ae45760405162461bcd60e51b8152602060048201526043602482015260008051602061407a83398151915260448201527f7265733a2070616972696e6720707265636f6d70696c652063616c6c206661696064820152621b195960ea1b608482015260a4016107fa565b8151600114612b495760405162461bcd60e51b8152602060048201526039602482015260008051602061407a83398151915260448201527f7265733a2050616972696e6720756e7375636365737366756c0000000000000060648201526084016107fa565b7f34d57e230be557a52d94166eb9035810e61ac973182a92b09e6b0e99110665a9898c8a600001518b6020015187604051612b88959493929190613f5d565b60405180910390a1612ba48b848a600001518b6020015161341b565b96505050505050509295509295909350565b60ff83166000908152608760209081526040808320858452909152812063ffffffff831660148110612bea57612bea613ce2565b0154949350505050565b612bfc612e09565b6001600160a01b038116612c615760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b60648201526084016107fa565b610b8381612f34565b608860009054906101000a90046001600160a01b03166001600160a01b031663eab66d7a6040518163ffffffff1660e01b8152600401602060405180830381865afa158015612cbd573d6000803e3d6000fd5b505050506040513d601f19601f82011682018060405250810190612ce19190613c56565b6001600160a01b0316336001600160a01b031614612d545760405162461bcd60e51b815260206004820152602a60248201527f6d73672e73656e646572206973206e6f74207065726d697373696f6e6564206160448201526939903ab73830bab9b2b960b11b60648201526084016107fa565b608954198119608954191614612dd25760405162461bcd60e51b815260206004820152603860248201527f5061757361626c652e756e70617573653a20696e76616c696420617474656d7060448201527f7420746f2070617573652066756e6374696f6e616c697479000000000000000060648201526084016107fa565b608981905560405181815233907f3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c90602001610a42565b33612e1261181e565b6001600160a01b0316146114585760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e657260448201526064016107fa565b600080826000015183602001518460400151856060015186608001518760a001518860c001518960e00151604051602001612f1598979695949392919097885260e096871b6001600160e01b031990811660208a015295871b8616602489015293861b851660288801529190941b909216602c85015260a09290921b6001600160a01b031916603084015260601b6bffffffffffffffffffffffff1916603c830152605082015260700190565b60408051601f1981840301815291905280516020909101209392505050565b603380546001600160a01b038381166001600160a01b0319831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b60655460408051918252602082018390527fcd1b2c2a220284accd1f9effd811cdecb6beaa4638618b48bbea07ce7ae16996910160405180910390a1606555565b6088546001600160a01b0316158015612fe857506001600160a01b03821615155b61306a5760405162461bcd60e51b815260206004820152604760248201527f5061757361626c652e5f696e697469616c697a655061757365723a205f696e6960448201527f7469616c697a6550617573657228292063616e206f6e6c792062652063616c6c6064820152666564206f6e636560c81b608482015260a4016107fa565b608981905560405181815233907fab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d9060200160405180910390a250608880546001600160a01b0319166001600160a01b0392909216919091179055565b608d54604080516001600160a01b03928316815291831660208301527f774b126b94b3cc801460a024dd575406c3ebf27affd7c36198a53ac6655f056d910160405180910390a1608d80546001600160a01b0319166001600160a01b0392909216919091179055565b6000604084013560e01c600101820261314a601f82613c89565b95945050505050565b8060ff166001141561319457608b80546000906131759063ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff1602179055505b8060ff16600214156131dd57608b80546004906131be90640100000000900463ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff1602179055505b8060ff166003141561322557608b805460089061320690600160401b900463ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff1602179055505b8060ff166004141561326d57608b8054600c9061324e90600160601b900463ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff1602179055505b8060ff16600514156132b557608b805460109061329690600160801b900463ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff1602179055505b8060ff16600614156132fd57608b80546014906132de90600160a01b900463ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff1602179055505b8060ff1660071415610b8357608b805460189061332690600160c01b900463ffffffff16613d4f565b91906101000a81548163ffffffff021916908363ffffffff16021790555050565b815163ffffffff808316911611156133ab5760405162461bcd60e51b815260206004820152602160248201527f50726f7669646564207374616b6520696e64657820697320746f6f206561726c6044820152607960f81b60648201526084016107fa565b5050565b6000808080806133cd60008051602061401a83398151915287613f49565b90505b6133d981613454565b909350915060008051602061401a833981519152828309831415613401579590945092505050565b60008051602061401a8339815191526001820890506133d0565b6000848484846040516020016134349493929190613fc6565b604051602081830303815290604052805190602001209050949350505050565b6000808060008051602061401a833981519152600360008051602061401a8339815191528660008051602061401a8339815191528889090908905060006134ca827f0c19139cb84c680a6e14116da060561765e05aa45a1c72a34f082305b61f3f5260008051602061401a8339815191526134d6565b91959194509092505050565b6000806134e161355a565b6134e9613578565b602080825281810181905260408201819052606082018890526080820187905260a082018690528260c08360056107d05a03fa925082801561251057508261353057600080fd5b505195945050505050565b604051806101800160405280600c906020820280368337509192915050565b60405180602001604052806001906020820280368337509192915050565b6040518060c001604052806006906020820280368337509192915050565b803561ffff811681146135a857600080fd5b919050565b6000602082840312156135bf57600080fd5b6135c882613596565b9392505050565b6000602082840312156135e157600080fd5b5035919050565b803560ff811681146135a857600080fd5b60008060006060848603121561360e57600080fd5b613617846135e8565b95602085013595506040909401359392505050565b6001600160a01b0381168114610b8357600080fd5b80356135a88161362c565b60006020828403121561365e57600080fd5b81356135c88161362c565b63ffffffff81168114610b8357600080fd5b80356135a881613669565b60008060006060848603121561369b57600080fd5b833592506020840135915060408401356136b481613669565b809150509250925092565b60008083601f8401126136d157600080fd5b50813567ffffffffffffffff8111156136e957600080fd5b60208301915083602082850101111561370157600080fd5b9250929050565b634e487b7160e01b600052604160045260246000fd5b6001600160601b0381168114610b8357600080fd5b80356135a88161371e565b600061010080838503121561375257600080fd5b6040519081019067ffffffffffffffff8211818310171561378357634e487b7160e01b600052604160045260246000fd5b81604052809250833581526020840135915061379e82613669565b8160208201526137b06040850161367b565b60408201526137c16060850161367b565b60608201526137d26080850161367b565b60808201526137e360a08501613733565b60a08201526137f460c08501613641565b60c082015260e084013560e0820152505092915050565b600080600083850361018081121561382257600080fd5b843567ffffffffffffffff8082111561383a57600080fd5b613846888389016136bf565b90965094506101609150601f19830182131561386157600080fd5b604051925060808301838110828211171561388c57634e487b7160e01b600052604160045260246000fd5b6040525061389d876020880161373e565b82526138ac61012087016135e8565b602083015261014086013560408301528501356138c881613669565b606082015292959194509192509050565b6000602082840312156138eb57600080fd5b6135c8826135e8565b60008060008060008060c0878903121561390d57600080fd5b86356139188161362c565b955060208701356139288161362c565b945061393660408801613596565b935061394460608801613596565b92506080870135915060a087013561395b8161362c565b809150509295509295509295565b600080600080610160858703121561398057600080fd5b613989856135e8565b93506020850135925060408501356139a081613669565b91506139af866060870161373e565b905092959194509250565b600080600080600080600060c0888a0312156139d557600080fd5b87356139e08161362c565b965060208801356139f08161362c565b95506139fe604089016135e8565b94506060880135613a0e81613669565b93506080880135613a1e81613669565b925060a088013567ffffffffffffffff811115613a3a57600080fd5b613a468a828b016136bf565b989b979a50959850939692959293505050565b60008060208385031215613a6c57600080fd5b823567ffffffffffffffff811115613a8357600080fd5b613a8f858286016136bf565b90969095509350505050565b600080600060608486031215613ab057600080fd5b613ab9846135e8565b92506020840135915060408401356136b481613669565b600060208284031215613ae257600080fd5b81356135c881613669565b6020808252604b908201527f446174614c617972536572766963654d616e616765722e76616c69645468726560408201527f73686f6c64733a20696e76616c6964205f71756f72756d5468726573686f6c6460608201526a4261736973506f696e747360a81b608082015260a00190565b6020808252604e908201527f446174614c617972536572766963654d616e616765722e76616c69645468726560408201527f73686f6c64733a20696e76616c6964205f61647665727361727954687265736860608201526d6f6c644261736973506f696e747360901b608082015260a00190565b602080825260609082018190527f446174614c617972536572766963654d616e616765722e76616c69645468726560408301527f73686f6c64733a2051756f72756d207468726573686f6c64206d757374206265908201527f207374726963746c792067726561746572207468616e20616476657273617279608082015260a00190565b600060208284031215613c6857600080fd5b81516135c88161362c565b634e487b7160e01b600052601160045260246000fd5b6000816000190483118215151615613ca357613ca3613c73565b500290565b600060208284031215613cba57600080fd5b815180151581146135c857600080fd5b60008219821115613cdd57613cdd613c73565b500190565b634e487b7160e01b600052603260045260246000fd5b634e487b7160e01b600052601260045260246000fd5b600082613d1d57613d1d613cf8565b500490565b8183823760009101908152919050565b600060208284031215613d4457600080fd5b81516135c881613669565b600063ffffffff80831681811415613d6957613d69613c73565b6001019392505050565b600063ffffffff808316818516808303821115613d9257613d92613c73565b01949350505050565b81835281816020850137506000828201602090810191909152601f909101601f19169091010190565b60006101a060018060a01b0387168352855180516020850152602081015163ffffffff808216604087015280604084015116606087015280606084015116608087015250506080810151613e2060a086018263ffffffff169052565b5060a08101516001600160601b031660c0858101919091528101516001600160a01b031660e0808601919091520151610100840152602086015160ff166101208401526040860151610140840152606086015163ffffffff166101608401526101808301819052613e948184018587613d9b565b979650505050505050565b600060808284031215613eb157600080fd5b6040516080810181811067ffffffffffffffff82111715613ee257634e487b7160e01b600052604160045260246000fd5b6040528251613ef081613669565b81526020830151613f0081613669565b60208201526040830151613f138161371e565b60408201526060830151613f268161371e565b60608201529392505050565b600082821015613f4457613f44613c73565b500390565b600082613f5857613f58613cf8565b500690565b600060a08201878352602063ffffffff88168185015286604085015285606085015260a0608085015281855180845260c086019150828701935060005b81811015613fb657845183529383019391830191600101613f9a565b50909a9950505050505050505050565b63ffffffff60e01b8560e01b1681526000600482018551602080880160005b8381101561400157815185529382019390820190600101613fe5565b5050958252509384019290925250506040019291505056fe30644e72e131a029b85045b68181585d97816a916871ca8d3c208c16d87cfd47446174614c617972536572766963654d616e616765722e696e69744461746153446174614c617972536572766963654d616e616765722e636f6e6669726d4461424c535369676e6174757265436865636b65722e636865636b5369676e617475a2646970667358221220e56ab13e217c70ebaae14a55cfabe56fb45eff605dd308b0bb7d17e1d8081b7764736f6c634300080c0033",
}

// ContractDataLayrServiceManagerABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractDataLayrServiceManagerMetaData.ABI instead.
var ContractDataLayrServiceManagerABI = ContractDataLayrServiceManagerMetaData.ABI

// ContractDataLayrServiceManagerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ContractDataLayrServiceManagerMetaData.Bin instead.
var ContractDataLayrServiceManagerBin = ContractDataLayrServiceManagerMetaData.Bin

// DeployContractDataLayrServiceManager deploys a new Ethereum contract, binding an instance of ContractDataLayrServiceManager to it.
func DeployContractDataLayrServiceManager(auth *bind.TransactOpts, backend bind.ContractBackend, _registry common.Address, _investmentManager common.Address, _eigenLayrDelegation common.Address, _collateralToken common.Address, _dataLayrChallenge common.Address, _dataLayrBombVerifier common.Address, _dataPermissionManager common.Address) (common.Address, *types.Transaction, *ContractDataLayrServiceManager, error) {
	parsed, err := ContractDataLayrServiceManagerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ContractDataLayrServiceManagerBin), backend, _registry, _investmentManager, _eigenLayrDelegation, _collateralToken, _dataLayrChallenge, _dataLayrBombVerifier, _dataPermissionManager)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ContractDataLayrServiceManager{ContractDataLayrServiceManagerCaller: ContractDataLayrServiceManagerCaller{contract: contract}, ContractDataLayrServiceManagerTransactor: ContractDataLayrServiceManagerTransactor{contract: contract}, ContractDataLayrServiceManagerFilterer: ContractDataLayrServiceManagerFilterer{contract: contract}}, nil
}

// ContractDataLayrServiceManager is an auto generated Go binding around an Ethereum contract.
type ContractDataLayrServiceManager struct {
	ContractDataLayrServiceManagerCaller     // Read-only binding to the contract
	ContractDataLayrServiceManagerTransactor // Write-only binding to the contract
	ContractDataLayrServiceManagerFilterer   // Log filterer for contract events
}

// ContractDataLayrServiceManagerCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractDataLayrServiceManagerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractDataLayrServiceManagerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractDataLayrServiceManagerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractDataLayrServiceManagerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractDataLayrServiceManagerSession struct {
	Contract     *ContractDataLayrServiceManager // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                   // Call options to use throughout this session
	TransactOpts bind.TransactOpts               // Transaction auth options to use throughout this session
}

// ContractDataLayrServiceManagerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractDataLayrServiceManagerCallerSession struct {
	Contract *ContractDataLayrServiceManagerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                         // Call options to use throughout this session
}

// ContractDataLayrServiceManagerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractDataLayrServiceManagerTransactorSession struct {
	Contract     *ContractDataLayrServiceManagerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                         // Transaction auth options to use throughout this session
}

// ContractDataLayrServiceManagerRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerRaw struct {
	Contract *ContractDataLayrServiceManager // Generic contract binding to access the raw methods on
}

// ContractDataLayrServiceManagerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerCallerRaw struct {
	Contract *ContractDataLayrServiceManagerCaller // Generic read-only contract binding to access the raw methods on
}

// ContractDataLayrServiceManagerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerTransactorRaw struct {
	Contract *ContractDataLayrServiceManagerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractDataLayrServiceManager creates a new instance of ContractDataLayrServiceManager, bound to a specific deployed contract.
func NewContractDataLayrServiceManager(address common.Address, backend bind.ContractBackend) (*ContractDataLayrServiceManager, error) {
	contract, err := bindContractDataLayrServiceManager(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManager{ContractDataLayrServiceManagerCaller: ContractDataLayrServiceManagerCaller{contract: contract}, ContractDataLayrServiceManagerTransactor: ContractDataLayrServiceManagerTransactor{contract: contract}, ContractDataLayrServiceManagerFilterer: ContractDataLayrServiceManagerFilterer{contract: contract}}, nil
}

// NewContractDataLayrServiceManagerCaller creates a new read-only instance of ContractDataLayrServiceManager, bound to a specific deployed contract.
func NewContractDataLayrServiceManagerCaller(address common.Address, caller bind.ContractCaller) (*ContractDataLayrServiceManagerCaller, error) {
	contract, err := bindContractDataLayrServiceManager(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerCaller{contract: contract}, nil
}

// NewContractDataLayrServiceManagerTransactor creates a new write-only instance of ContractDataLayrServiceManager, bound to a specific deployed contract.
func NewContractDataLayrServiceManagerTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractDataLayrServiceManagerTransactor, error) {
	contract, err := bindContractDataLayrServiceManager(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerTransactor{contract: contract}, nil
}

// NewContractDataLayrServiceManagerFilterer creates a new log filterer instance of ContractDataLayrServiceManager, bound to a specific deployed contract.
func NewContractDataLayrServiceManagerFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractDataLayrServiceManagerFilterer, error) {
	contract, err := bindContractDataLayrServiceManager(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerFilterer{contract: contract}, nil
}

// bindContractDataLayrServiceManager binds a generic wrapper to an already deployed contract.
func bindContractDataLayrServiceManager(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractDataLayrServiceManagerABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractDataLayrServiceManager.Contract.ContractDataLayrServiceManagerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.ContractDataLayrServiceManagerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.ContractDataLayrServiceManagerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractDataLayrServiceManager.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.contract.Transact(opts, method, params...)
}

// BIPMULTIPLIER is a free data retrieval call binding the contract method 0xa3c7eaf0.
//
// Solidity: function BIP_MULTIPLIER() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) BIPMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "BIP_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BIPMULTIPLIER is a free data retrieval call binding the contract method 0xa3c7eaf0.
//
// Solidity: function BIP_MULTIPLIER() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) BIPMULTIPLIER() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.BIPMULTIPLIER(&_ContractDataLayrServiceManager.CallOpts)
}

// BIPMULTIPLIER is a free data retrieval call binding the contract method 0xa3c7eaf0.
//
// Solidity: function BIP_MULTIPLIER() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) BIPMULTIPLIER() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.BIPMULTIPLIER(&_ContractDataLayrServiceManager.CallOpts)
}

// BLOCKSTALEMEASURE is a free data retrieval call binding the contract method 0x5e8b3f2d.
//
// Solidity: function BLOCK_STALE_MEASURE() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) BLOCKSTALEMEASURE(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "BLOCK_STALE_MEASURE")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// BLOCKSTALEMEASURE is a free data retrieval call binding the contract method 0x5e8b3f2d.
//
// Solidity: function BLOCK_STALE_MEASURE() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) BLOCKSTALEMEASURE() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.BLOCKSTALEMEASURE(&_ContractDataLayrServiceManager.CallOpts)
}

// BLOCKSTALEMEASURE is a free data retrieval call binding the contract method 0x5e8b3f2d.
//
// Solidity: function BLOCK_STALE_MEASURE() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) BLOCKSTALEMEASURE() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.BLOCKSTALEMEASURE(&_ContractDataLayrServiceManager.CallOpts)
}

// DURATIONSCALE is a free data retrieval call binding the contract method 0x31a219c5.
//
// Solidity: function DURATION_SCALE() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) DURATIONSCALE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "DURATION_SCALE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DURATIONSCALE is a free data retrieval call binding the contract method 0x31a219c5.
//
// Solidity: function DURATION_SCALE() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) DURATIONSCALE() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.DURATIONSCALE(&_ContractDataLayrServiceManager.CallOpts)
}

// DURATIONSCALE is a free data retrieval call binding the contract method 0x31a219c5.
//
// Solidity: function DURATION_SCALE() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) DURATIONSCALE() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.DURATIONSCALE(&_ContractDataLayrServiceManager.CallOpts)
}

// MAXDATASTOREDURATION is a free data retrieval call binding the contract method 0x578ae5a1.
//
// Solidity: function MAX_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) MAXDATASTOREDURATION(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "MAX_DATASTORE_DURATION")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// MAXDATASTOREDURATION is a free data retrieval call binding the contract method 0x578ae5a1.
//
// Solidity: function MAX_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) MAXDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManager.Contract.MAXDATASTOREDURATION(&_ContractDataLayrServiceManager.CallOpts)
}

// MAXDATASTOREDURATION is a free data retrieval call binding the contract method 0x578ae5a1.
//
// Solidity: function MAX_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) MAXDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManager.Contract.MAXDATASTOREDURATION(&_ContractDataLayrServiceManager.CallOpts)
}

// MINDATASTOREDURATION is a free data retrieval call binding the contract method 0x1fdab6e4.
//
// Solidity: function MIN_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) MINDATASTOREDURATION(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "MIN_DATASTORE_DURATION")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// MINDATASTOREDURATION is a free data retrieval call binding the contract method 0x1fdab6e4.
//
// Solidity: function MIN_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) MINDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManager.Contract.MINDATASTOREDURATION(&_ContractDataLayrServiceManager.CallOpts)
}

// MINDATASTOREDURATION is a free data retrieval call binding the contract method 0x1fdab6e4.
//
// Solidity: function MIN_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) MINDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManager.Contract.MINDATASTOREDURATION(&_ContractDataLayrServiceManager.CallOpts)
}

// NUMDSPERBLOCKPERDURATION is a free data retrieval call binding the contract method 0x5f87abbb.
//
// Solidity: function NUM_DS_PER_BLOCK_PER_DURATION() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) NUMDSPERBLOCKPERDURATION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "NUM_DS_PER_BLOCK_PER_DURATION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NUMDSPERBLOCKPERDURATION is a free data retrieval call binding the contract method 0x5f87abbb.
//
// Solidity: function NUM_DS_PER_BLOCK_PER_DURATION() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) NUMDSPERBLOCKPERDURATION() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.NUMDSPERBLOCKPERDURATION(&_ContractDataLayrServiceManager.CallOpts)
}

// NUMDSPERBLOCKPERDURATION is a free data retrieval call binding the contract method 0x5f87abbb.
//
// Solidity: function NUM_DS_PER_BLOCK_PER_DURATION() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) NUMDSPERBLOCKPERDURATION() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.NUMDSPERBLOCKPERDURATION(&_ContractDataLayrServiceManager.CallOpts)
}

// AdversaryThresholdBasisPoints is a free data retrieval call binding the contract method 0x3594b60f.
//
// Solidity: function adversaryThresholdBasisPoints() view returns(uint16)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) AdversaryThresholdBasisPoints(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "adversaryThresholdBasisPoints")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// AdversaryThresholdBasisPoints is a free data retrieval call binding the contract method 0x3594b60f.
//
// Solidity: function adversaryThresholdBasisPoints() view returns(uint16)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) AdversaryThresholdBasisPoints() (uint16, error) {
	return _ContractDataLayrServiceManager.Contract.AdversaryThresholdBasisPoints(&_ContractDataLayrServiceManager.CallOpts)
}

// AdversaryThresholdBasisPoints is a free data retrieval call binding the contract method 0x3594b60f.
//
// Solidity: function adversaryThresholdBasisPoints() view returns(uint16)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) AdversaryThresholdBasisPoints() (uint16, error) {
	return _ContractDataLayrServiceManager.Contract.AdversaryThresholdBasisPoints(&_ContractDataLayrServiceManager.CallOpts)
}

// CalculateFee is a free data retrieval call binding the contract method 0x39fe2e71.
//
// Solidity: function calculateFee(uint256 totalBytes, uint256 _feePerBytePerTime, uint32 storePeriodLength) pure returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) CalculateFee(opts *bind.CallOpts, totalBytes *big.Int, _feePerBytePerTime *big.Int, storePeriodLength uint32) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "calculateFee", totalBytes, _feePerBytePerTime, storePeriodLength)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// CalculateFee is a free data retrieval call binding the contract method 0x39fe2e71.
//
// Solidity: function calculateFee(uint256 totalBytes, uint256 _feePerBytePerTime, uint32 storePeriodLength) pure returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) CalculateFee(totalBytes *big.Int, _feePerBytePerTime *big.Int, storePeriodLength uint32) (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.CalculateFee(&_ContractDataLayrServiceManager.CallOpts, totalBytes, _feePerBytePerTime, storePeriodLength)
}

// CalculateFee is a free data retrieval call binding the contract method 0x39fe2e71.
//
// Solidity: function calculateFee(uint256 totalBytes, uint256 _feePerBytePerTime, uint32 storePeriodLength) pure returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) CalculateFee(totalBytes *big.Int, _feePerBytePerTime *big.Int, storePeriodLength uint32) (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.CalculateFee(&_ContractDataLayrServiceManager.CallOpts, totalBytes, _feePerBytePerTime, storePeriodLength)
}

// CollateralToken is a free data retrieval call binding the contract method 0xb2016bd4.
//
// Solidity: function collateralToken() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) CollateralToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "collateralToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CollateralToken is a free data retrieval call binding the contract method 0xb2016bd4.
//
// Solidity: function collateralToken() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) CollateralToken() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.CollateralToken(&_ContractDataLayrServiceManager.CallOpts)
}

// CollateralToken is a free data retrieval call binding the contract method 0xb2016bd4.
//
// Solidity: function collateralToken() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) CollateralToken() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.CollateralToken(&_ContractDataLayrServiceManager.CallOpts)
}

// ConfirmDataStoreTimeout is a free data retrieval call binding the contract method 0x4d53cebb.
//
// Solidity: function confirmDataStoreTimeout() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) ConfirmDataStoreTimeout(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "confirmDataStoreTimeout")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// ConfirmDataStoreTimeout is a free data retrieval call binding the contract method 0x4d53cebb.
//
// Solidity: function confirmDataStoreTimeout() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) ConfirmDataStoreTimeout() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.ConfirmDataStoreTimeout(&_ContractDataLayrServiceManager.CallOpts)
}

// ConfirmDataStoreTimeout is a free data retrieval call binding the contract method 0x4d53cebb.
//
// Solidity: function confirmDataStoreTimeout() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) ConfirmDataStoreTimeout() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.ConfirmDataStoreTimeout(&_ContractDataLayrServiceManager.CallOpts)
}

// DataLayrBombVerifier is a free data retrieval call binding the contract method 0x5e69019c.
//
// Solidity: function dataLayrBombVerifier() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) DataLayrBombVerifier(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "dataLayrBombVerifier")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DataLayrBombVerifier is a free data retrieval call binding the contract method 0x5e69019c.
//
// Solidity: function dataLayrBombVerifier() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) DataLayrBombVerifier() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.DataLayrBombVerifier(&_ContractDataLayrServiceManager.CallOpts)
}

// DataLayrBombVerifier is a free data retrieval call binding the contract method 0x5e69019c.
//
// Solidity: function dataLayrBombVerifier() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) DataLayrBombVerifier() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.DataLayrBombVerifier(&_ContractDataLayrServiceManager.CallOpts)
}

// DataLayrChallenge is a free data retrieval call binding the contract method 0x573792c6.
//
// Solidity: function dataLayrChallenge() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) DataLayrChallenge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "dataLayrChallenge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DataLayrChallenge is a free data retrieval call binding the contract method 0x573792c6.
//
// Solidity: function dataLayrChallenge() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) DataLayrChallenge() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.DataLayrChallenge(&_ContractDataLayrServiceManager.CallOpts)
}

// DataLayrChallenge is a free data retrieval call binding the contract method 0x573792c6.
//
// Solidity: function dataLayrChallenge() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) DataLayrChallenge() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.DataLayrChallenge(&_ContractDataLayrServiceManager.CallOpts)
}

// DataPermissionManager is a free data retrieval call binding the contract method 0x7b49d55f.
//
// Solidity: function dataPermissionManager() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) DataPermissionManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "dataPermissionManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DataPermissionManager is a free data retrieval call binding the contract method 0x7b49d55f.
//
// Solidity: function dataPermissionManager() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) DataPermissionManager() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.DataPermissionManager(&_ContractDataLayrServiceManager.CallOpts)
}

// DataPermissionManager is a free data retrieval call binding the contract method 0x7b49d55f.
//
// Solidity: function dataPermissionManager() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) DataPermissionManager() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.DataPermissionManager(&_ContractDataLayrServiceManager.CallOpts)
}

// DataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0x1bd2b3cf.
//
// Solidity: function dataStoreHashesForDurationAtTimestamp(uint8 , uint256 , uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) DataStoreHashesForDurationAtTimestamp(opts *bind.CallOpts, arg0 uint8, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "dataStoreHashesForDurationAtTimestamp", arg0, arg1, arg2)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0x1bd2b3cf.
//
// Solidity: function dataStoreHashesForDurationAtTimestamp(uint8 , uint256 , uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) DataStoreHashesForDurationAtTimestamp(arg0 uint8, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.DataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManager.CallOpts, arg0, arg1, arg2)
}

// DataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0x1bd2b3cf.
//
// Solidity: function dataStoreHashesForDurationAtTimestamp(uint8 , uint256 , uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) DataStoreHashesForDurationAtTimestamp(arg0 uint8, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.DataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManager.CallOpts, arg0, arg1, arg2)
}

// DataStoreIdToSignatureHash is a free data retrieval call binding the contract method 0xfc2c6058.
//
// Solidity: function dataStoreIdToSignatureHash(uint32 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) DataStoreIdToSignatureHash(opts *bind.CallOpts, arg0 uint32) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "dataStoreIdToSignatureHash", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DataStoreIdToSignatureHash is a free data retrieval call binding the contract method 0xfc2c6058.
//
// Solidity: function dataStoreIdToSignatureHash(uint32 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) DataStoreIdToSignatureHash(arg0 uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.DataStoreIdToSignatureHash(&_ContractDataLayrServiceManager.CallOpts, arg0)
}

// DataStoreIdToSignatureHash is a free data retrieval call binding the contract method 0xfc2c6058.
//
// Solidity: function dataStoreIdToSignatureHash(uint32 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) DataStoreIdToSignatureHash(arg0 uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.DataStoreIdToSignatureHash(&_ContractDataLayrServiceManager.CallOpts, arg0)
}

// DataStoresForDuration is a free data retrieval call binding the contract method 0x33223aea.
//
// Solidity: function dataStoresForDuration() view returns(uint32 one_duration, uint32 two_duration, uint32 three_duration, uint32 four_duration, uint32 five_duration, uint32 six_duration, uint32 seven_duration, uint32 dataStoreId, uint32 latestTime)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) DataStoresForDuration(opts *bind.CallOpts) (struct {
	OneDuration   uint32
	TwoDuration   uint32
	ThreeDuration uint32
	FourDuration  uint32
	FiveDuration  uint32
	SixDuration   uint32
	SevenDuration uint32
	DataStoreId   uint32
	LatestTime    uint32
}, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "dataStoresForDuration")

	outstruct := new(struct {
		OneDuration   uint32
		TwoDuration   uint32
		ThreeDuration uint32
		FourDuration  uint32
		FiveDuration  uint32
		SixDuration   uint32
		SevenDuration uint32
		DataStoreId   uint32
		LatestTime    uint32
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.OneDuration = *abi.ConvertType(out[0], new(uint32)).(*uint32)
	outstruct.TwoDuration = *abi.ConvertType(out[1], new(uint32)).(*uint32)
	outstruct.ThreeDuration = *abi.ConvertType(out[2], new(uint32)).(*uint32)
	outstruct.FourDuration = *abi.ConvertType(out[3], new(uint32)).(*uint32)
	outstruct.FiveDuration = *abi.ConvertType(out[4], new(uint32)).(*uint32)
	outstruct.SixDuration = *abi.ConvertType(out[5], new(uint32)).(*uint32)
	outstruct.SevenDuration = *abi.ConvertType(out[6], new(uint32)).(*uint32)
	outstruct.DataStoreId = *abi.ConvertType(out[7], new(uint32)).(*uint32)
	outstruct.LatestTime = *abi.ConvertType(out[8], new(uint32)).(*uint32)

	return *outstruct, err

}

// DataStoresForDuration is a free data retrieval call binding the contract method 0x33223aea.
//
// Solidity: function dataStoresForDuration() view returns(uint32 one_duration, uint32 two_duration, uint32 three_duration, uint32 four_duration, uint32 five_duration, uint32 six_duration, uint32 seven_duration, uint32 dataStoreId, uint32 latestTime)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) DataStoresForDuration() (struct {
	OneDuration   uint32
	TwoDuration   uint32
	ThreeDuration uint32
	FourDuration  uint32
	FiveDuration  uint32
	SixDuration   uint32
	SevenDuration uint32
	DataStoreId   uint32
	LatestTime    uint32
}, error) {
	return _ContractDataLayrServiceManager.Contract.DataStoresForDuration(&_ContractDataLayrServiceManager.CallOpts)
}

// DataStoresForDuration is a free data retrieval call binding the contract method 0x33223aea.
//
// Solidity: function dataStoresForDuration() view returns(uint32 one_duration, uint32 two_duration, uint32 three_duration, uint32 four_duration, uint32 five_duration, uint32 six_duration, uint32 seven_duration, uint32 dataStoreId, uint32 latestTime)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) DataStoresForDuration() (struct {
	OneDuration   uint32
	TwoDuration   uint32
	ThreeDuration uint32
	FourDuration  uint32
	FiveDuration  uint32
	SixDuration   uint32
	SevenDuration uint32
	DataStoreId   uint32
	LatestTime    uint32
}, error) {
	return _ContractDataLayrServiceManager.Contract.DataStoresForDuration(&_ContractDataLayrServiceManager.CallOpts)
}

// EigenLayrDelegation is a free data retrieval call binding the contract method 0x33d2433a.
//
// Solidity: function eigenLayrDelegation() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) EigenLayrDelegation(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "eigenLayrDelegation")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EigenLayrDelegation is a free data retrieval call binding the contract method 0x33d2433a.
//
// Solidity: function eigenLayrDelegation() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) EigenLayrDelegation() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.EigenLayrDelegation(&_ContractDataLayrServiceManager.CallOpts)
}

// EigenLayrDelegation is a free data retrieval call binding the contract method 0x33d2433a.
//
// Solidity: function eigenLayrDelegation() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) EigenLayrDelegation() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.EigenLayrDelegation(&_ContractDataLayrServiceManager.CallOpts)
}

// FeePerBytePerTime is a free data retrieval call binding the contract method 0xd21eed4f.
//
// Solidity: function feePerBytePerTime() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) FeePerBytePerTime(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "feePerBytePerTime")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeePerBytePerTime is a free data retrieval call binding the contract method 0xd21eed4f.
//
// Solidity: function feePerBytePerTime() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) FeePerBytePerTime() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.FeePerBytePerTime(&_ContractDataLayrServiceManager.CallOpts)
}

// FeePerBytePerTime is a free data retrieval call binding the contract method 0xd21eed4f.
//
// Solidity: function feePerBytePerTime() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) FeePerBytePerTime() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.FeePerBytePerTime(&_ContractDataLayrServiceManager.CallOpts)
}

// FeeSetter is a free data retrieval call binding the contract method 0x87cf3ef4.
//
// Solidity: function feeSetter() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) FeeSetter(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "feeSetter")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// FeeSetter is a free data retrieval call binding the contract method 0x87cf3ef4.
//
// Solidity: function feeSetter() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) FeeSetter() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.FeeSetter(&_ContractDataLayrServiceManager.CallOpts)
}

// FeeSetter is a free data retrieval call binding the contract method 0x87cf3ef4.
//
// Solidity: function feeSetter() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) FeeSetter() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.FeeSetter(&_ContractDataLayrServiceManager.CallOpts)
}

// GetDataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0xed82c0ee.
//
// Solidity: function getDataStoreHashesForDurationAtTimestamp(uint8 duration, uint256 timestamp, uint32 index) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) GetDataStoreHashesForDurationAtTimestamp(opts *bind.CallOpts, duration uint8, timestamp *big.Int, index uint32) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "getDataStoreHashesForDurationAtTimestamp", duration, timestamp, index)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetDataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0xed82c0ee.
//
// Solidity: function getDataStoreHashesForDurationAtTimestamp(uint8 duration, uint256 timestamp, uint32 index) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) GetDataStoreHashesForDurationAtTimestamp(duration uint8, timestamp *big.Int, index uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.GetDataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManager.CallOpts, duration, timestamp, index)
}

// GetDataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0xed82c0ee.
//
// Solidity: function getDataStoreHashesForDurationAtTimestamp(uint8 duration, uint256 timestamp, uint32 index) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) GetDataStoreHashesForDurationAtTimestamp(duration uint8, timestamp *big.Int, index uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.GetDataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManager.CallOpts, duration, timestamp, index)
}

// GetNumDataStoresForDuration is a free data retrieval call binding the contract method 0x73441c4e.
//
// Solidity: function getNumDataStoresForDuration(uint8 duration) view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) GetNumDataStoresForDuration(opts *bind.CallOpts, duration uint8) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "getNumDataStoresForDuration", duration)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetNumDataStoresForDuration is a free data retrieval call binding the contract method 0x73441c4e.
//
// Solidity: function getNumDataStoresForDuration(uint8 duration) view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) GetNumDataStoresForDuration(duration uint8) (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.GetNumDataStoresForDuration(&_ContractDataLayrServiceManager.CallOpts, duration)
}

// GetNumDataStoresForDuration is a free data retrieval call binding the contract method 0x73441c4e.
//
// Solidity: function getNumDataStoresForDuration(uint8 duration) view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) GetNumDataStoresForDuration(duration uint8) (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.GetNumDataStoresForDuration(&_ContractDataLayrServiceManager.CallOpts, duration)
}

// InvestmentManager is a free data retrieval call binding the contract method 0x4b31bb10.
//
// Solidity: function investmentManager() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) InvestmentManager(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "investmentManager")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// InvestmentManager is a free data retrieval call binding the contract method 0x4b31bb10.
//
// Solidity: function investmentManager() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) InvestmentManager() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.InvestmentManager(&_ContractDataLayrServiceManager.CallOpts)
}

// InvestmentManager is a free data retrieval call binding the contract method 0x4b31bb10.
//
// Solidity: function investmentManager() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) InvestmentManager() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.InvestmentManager(&_ContractDataLayrServiceManager.CallOpts)
}

// LatestTime is a free data retrieval call binding the contract method 0x7dfd16d7.
//
// Solidity: function latestTime() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) LatestTime(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "latestTime")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// LatestTime is a free data retrieval call binding the contract method 0x7dfd16d7.
//
// Solidity: function latestTime() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) LatestTime() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.LatestTime(&_ContractDataLayrServiceManager.CallOpts)
}

// LatestTime is a free data retrieval call binding the contract method 0x7dfd16d7.
//
// Solidity: function latestTime() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) LatestTime() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.LatestTime(&_ContractDataLayrServiceManager.CallOpts)
}

// Log2NumPowersOfTau is a free data retrieval call binding the contract method 0xa50017a1.
//
// Solidity: function log2NumPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) Log2NumPowersOfTau(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "log2NumPowersOfTau")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Log2NumPowersOfTau is a free data retrieval call binding the contract method 0xa50017a1.
//
// Solidity: function log2NumPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Log2NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.Log2NumPowersOfTau(&_ContractDataLayrServiceManager.CallOpts)
}

// Log2NumPowersOfTau is a free data retrieval call binding the contract method 0xa50017a1.
//
// Solidity: function log2NumPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) Log2NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.Log2NumPowersOfTau(&_ContractDataLayrServiceManager.CallOpts)
}

// NumPowersOfTau is a free data retrieval call binding the contract method 0x046bf4a6.
//
// Solidity: function numPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) NumPowersOfTau(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "numPowersOfTau")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumPowersOfTau is a free data retrieval call binding the contract method 0x046bf4a6.
//
// Solidity: function numPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.NumPowersOfTau(&_ContractDataLayrServiceManager.CallOpts)
}

// NumPowersOfTau is a free data retrieval call binding the contract method 0x046bf4a6.
//
// Solidity: function numPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.NumPowersOfTau(&_ContractDataLayrServiceManager.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Owner() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.Owner(&_ContractDataLayrServiceManager.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) Owner() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.Owner(&_ContractDataLayrServiceManager.CallOpts)
}

// Paused is a free data retrieval call binding the contract method 0x5ac86ab7.
//
// Solidity: function paused(uint8 index) view returns(bool)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) Paused(opts *bind.CallOpts, index uint8) (bool, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "paused", index)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// Paused is a free data retrieval call binding the contract method 0x5ac86ab7.
//
// Solidity: function paused(uint8 index) view returns(bool)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Paused(index uint8) (bool, error) {
	return _ContractDataLayrServiceManager.Contract.Paused(&_ContractDataLayrServiceManager.CallOpts, index)
}

// Paused is a free data retrieval call binding the contract method 0x5ac86ab7.
//
// Solidity: function paused(uint8 index) view returns(bool)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) Paused(index uint8) (bool, error) {
	return _ContractDataLayrServiceManager.Contract.Paused(&_ContractDataLayrServiceManager.CallOpts, index)
}

// Paused0 is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) Paused0(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "paused0")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Paused0 is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Paused0() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.Paused0(&_ContractDataLayrServiceManager.CallOpts)
}

// Paused0 is a free data retrieval call binding the contract method 0x5c975abb.
//
// Solidity: function paused() view returns(uint256)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) Paused0() (*big.Int, error) {
	return _ContractDataLayrServiceManager.Contract.Paused0(&_ContractDataLayrServiceManager.CallOpts)
}

// PauserRegistry is a free data retrieval call binding the contract method 0x886f1195.
//
// Solidity: function pauserRegistry() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) PauserRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "pauserRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// PauserRegistry is a free data retrieval call binding the contract method 0x886f1195.
//
// Solidity: function pauserRegistry() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) PauserRegistry() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.PauserRegistry(&_ContractDataLayrServiceManager.CallOpts)
}

// PauserRegistry is a free data retrieval call binding the contract method 0x886f1195.
//
// Solidity: function pauserRegistry() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) PauserRegistry() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.PauserRegistry(&_ContractDataLayrServiceManager.CallOpts)
}

// QuorumThresholdBasisPoints is a free data retrieval call binding the contract method 0xb569157b.
//
// Solidity: function quorumThresholdBasisPoints() view returns(uint16)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) QuorumThresholdBasisPoints(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "quorumThresholdBasisPoints")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// QuorumThresholdBasisPoints is a free data retrieval call binding the contract method 0xb569157b.
//
// Solidity: function quorumThresholdBasisPoints() view returns(uint16)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) QuorumThresholdBasisPoints() (uint16, error) {
	return _ContractDataLayrServiceManager.Contract.QuorumThresholdBasisPoints(&_ContractDataLayrServiceManager.CallOpts)
}

// QuorumThresholdBasisPoints is a free data retrieval call binding the contract method 0xb569157b.
//
// Solidity: function quorumThresholdBasisPoints() view returns(uint16)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) QuorumThresholdBasisPoints() (uint16, error) {
	return _ContractDataLayrServiceManager.Contract.QuorumThresholdBasisPoints(&_ContractDataLayrServiceManager.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) Registry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "registry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Registry() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.Registry(&_ContractDataLayrServiceManager.CallOpts)
}

// Registry is a free data retrieval call binding the contract method 0x7b103999.
//
// Solidity: function registry() view returns(address)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) Registry() (common.Address, error) {
	return _ContractDataLayrServiceManager.Contract.Registry(&_ContractDataLayrServiceManager.CallOpts)
}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) TaskNumber(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "taskNumber")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) TaskNumber() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.TaskNumber(&_ContractDataLayrServiceManager.CallOpts)
}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) TaskNumber() (uint32, error) {
	return _ContractDataLayrServiceManager.Contract.TaskNumber(&_ContractDataLayrServiceManager.CallOpts)
}

// VerifyDataStoreMetadata is a free data retrieval call binding the contract method 0xba4994b1.
//
// Solidity: function verifyDataStoreMetadata(uint8 duration, uint256 timestamp, uint32 index, (bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32) metadata) view returns(bool)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) VerifyDataStoreMetadata(opts *bind.CallOpts, duration uint8, timestamp *big.Int, index uint32, metadata IDataLayrServiceManagerDataStoreMetadata) (bool, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "verifyDataStoreMetadata", duration, timestamp, index, metadata)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyDataStoreMetadata is a free data retrieval call binding the contract method 0xba4994b1.
//
// Solidity: function verifyDataStoreMetadata(uint8 duration, uint256 timestamp, uint32 index, (bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32) metadata) view returns(bool)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) VerifyDataStoreMetadata(duration uint8, timestamp *big.Int, index uint32, metadata IDataLayrServiceManagerDataStoreMetadata) (bool, error) {
	return _ContractDataLayrServiceManager.Contract.VerifyDataStoreMetadata(&_ContractDataLayrServiceManager.CallOpts, duration, timestamp, index, metadata)
}

// VerifyDataStoreMetadata is a free data retrieval call binding the contract method 0xba4994b1.
//
// Solidity: function verifyDataStoreMetadata(uint8 duration, uint256 timestamp, uint32 index, (bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32) metadata) view returns(bool)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) VerifyDataStoreMetadata(duration uint8, timestamp *big.Int, index uint32, metadata IDataLayrServiceManagerDataStoreMetadata) (bool, error) {
	return _ContractDataLayrServiceManager.Contract.VerifyDataStoreMetadata(&_ContractDataLayrServiceManager.CallOpts, duration, timestamp, index, metadata)
}

// ZeroPolynomialCommitmentMerkleRoots is a free data retrieval call binding the contract method 0x3367a3fb.
//
// Solidity: function zeroPolynomialCommitmentMerkleRoots(uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCaller) ZeroPolynomialCommitmentMerkleRoots(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManager.contract.Call(opts, &out, "zeroPolynomialCommitmentMerkleRoots", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ZeroPolynomialCommitmentMerkleRoots is a free data retrieval call binding the contract method 0x3367a3fb.
//
// Solidity: function zeroPolynomialCommitmentMerkleRoots(uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) ZeroPolynomialCommitmentMerkleRoots(arg0 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.ZeroPolynomialCommitmentMerkleRoots(&_ContractDataLayrServiceManager.CallOpts, arg0)
}

// ZeroPolynomialCommitmentMerkleRoots is a free data retrieval call binding the contract method 0x3367a3fb.
//
// Solidity: function zeroPolynomialCommitmentMerkleRoots(uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerCallerSession) ZeroPolynomialCommitmentMerkleRoots(arg0 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManager.Contract.ZeroPolynomialCommitmentMerkleRoots(&_ContractDataLayrServiceManager.CallOpts, arg0)
}

// CheckSignatures is a paid mutator transaction binding the contract method 0xdeaf4498.
//
// Solidity: function checkSignatures(bytes data) returns(uint32 taskNumberToConfirm, uint32 referenceBlockNumber, bytes32 msgHash, (uint256,uint256,uint256,uint256) signedTotals, bytes32 compressedSignatoryRecord)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) CheckSignatures(opts *bind.TransactOpts, data []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "checkSignatures", data)
}

// CheckSignatures is a paid mutator transaction binding the contract method 0xdeaf4498.
//
// Solidity: function checkSignatures(bytes data) returns(uint32 taskNumberToConfirm, uint32 referenceBlockNumber, bytes32 msgHash, (uint256,uint256,uint256,uint256) signedTotals, bytes32 compressedSignatoryRecord)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) CheckSignatures(data []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.CheckSignatures(&_ContractDataLayrServiceManager.TransactOpts, data)
}

// CheckSignatures is a paid mutator transaction binding the contract method 0xdeaf4498.
//
// Solidity: function checkSignatures(bytes data) returns(uint32 taskNumberToConfirm, uint32 referenceBlockNumber, bytes32 msgHash, (uint256,uint256,uint256,uint256) signedTotals, bytes32 compressedSignatoryRecord)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) CheckSignatures(data []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.CheckSignatures(&_ContractDataLayrServiceManager.TransactOpts, data)
}

// ConfirmDataStore is a paid mutator transaction binding the contract method 0x58942e73.
//
// Solidity: function confirmDataStore(bytes data, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) ConfirmDataStore(opts *bind.TransactOpts, data []byte, searchData IDataLayrServiceManagerDataStoreSearchData) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "confirmDataStore", data, searchData)
}

// ConfirmDataStore is a paid mutator transaction binding the contract method 0x58942e73.
//
// Solidity: function confirmDataStore(bytes data, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) ConfirmDataStore(data []byte, searchData IDataLayrServiceManagerDataStoreSearchData) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.ConfirmDataStore(&_ContractDataLayrServiceManager.TransactOpts, data, searchData)
}

// ConfirmDataStore is a paid mutator transaction binding the contract method 0x58942e73.
//
// Solidity: function confirmDataStore(bytes data, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) ConfirmDataStore(data []byte, searchData IDataLayrServiceManagerDataStoreSearchData) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.ConfirmDataStore(&_ContractDataLayrServiceManager.TransactOpts, data, searchData)
}

// FreezeOperator is a paid mutator transaction binding the contract method 0x38c8ee64.
//
// Solidity: function freezeOperator(address operator) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) FreezeOperator(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "freezeOperator", operator)
}

// FreezeOperator is a paid mutator transaction binding the contract method 0x38c8ee64.
//
// Solidity: function freezeOperator(address operator) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) FreezeOperator(operator common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.FreezeOperator(&_ContractDataLayrServiceManager.TransactOpts, operator)
}

// FreezeOperator is a paid mutator transaction binding the contract method 0x38c8ee64.
//
// Solidity: function freezeOperator(address operator) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) FreezeOperator(operator common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.FreezeOperator(&_ContractDataLayrServiceManager.TransactOpts, operator)
}

// InitDataStore is a paid mutator transaction binding the contract method 0xdcf49ea7.
//
// Solidity: function initDataStore(address feePayer, address confirmer, uint8 duration, uint32 referenceBlockNumber, uint32 totalOperatorsIndex, bytes header) returns(uint32 index)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) InitDataStore(opts *bind.TransactOpts, feePayer common.Address, confirmer common.Address, duration uint8, referenceBlockNumber uint32, totalOperatorsIndex uint32, header []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "initDataStore", feePayer, confirmer, duration, referenceBlockNumber, totalOperatorsIndex, header)
}

// InitDataStore is a paid mutator transaction binding the contract method 0xdcf49ea7.
//
// Solidity: function initDataStore(address feePayer, address confirmer, uint8 duration, uint32 referenceBlockNumber, uint32 totalOperatorsIndex, bytes header) returns(uint32 index)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) InitDataStore(feePayer common.Address, confirmer common.Address, duration uint8, referenceBlockNumber uint32, totalOperatorsIndex uint32, header []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.InitDataStore(&_ContractDataLayrServiceManager.TransactOpts, feePayer, confirmer, duration, referenceBlockNumber, totalOperatorsIndex, header)
}

// InitDataStore is a paid mutator transaction binding the contract method 0xdcf49ea7.
//
// Solidity: function initDataStore(address feePayer, address confirmer, uint8 duration, uint32 referenceBlockNumber, uint32 totalOperatorsIndex, bytes header) returns(uint32 index)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) InitDataStore(feePayer common.Address, confirmer common.Address, duration uint8, referenceBlockNumber uint32, totalOperatorsIndex uint32, header []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.InitDataStore(&_ContractDataLayrServiceManager.TransactOpts, feePayer, confirmer, duration, referenceBlockNumber, totalOperatorsIndex, header)
}

// Initialize is a paid mutator transaction binding the contract method 0x7bc9d56e.
//
// Solidity: function initialize(address _pauserRegistry, address initialOwner, uint16 _quorumThresholdBasisPoints, uint16 _adversaryThresholdBasisPoints, uint256 _feePerBytePerTime, address _feeSetter) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) Initialize(opts *bind.TransactOpts, _pauserRegistry common.Address, initialOwner common.Address, _quorumThresholdBasisPoints uint16, _adversaryThresholdBasisPoints uint16, _feePerBytePerTime *big.Int, _feeSetter common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "initialize", _pauserRegistry, initialOwner, _quorumThresholdBasisPoints, _adversaryThresholdBasisPoints, _feePerBytePerTime, _feeSetter)
}

// Initialize is a paid mutator transaction binding the contract method 0x7bc9d56e.
//
// Solidity: function initialize(address _pauserRegistry, address initialOwner, uint16 _quorumThresholdBasisPoints, uint16 _adversaryThresholdBasisPoints, uint256 _feePerBytePerTime, address _feeSetter) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Initialize(_pauserRegistry common.Address, initialOwner common.Address, _quorumThresholdBasisPoints uint16, _adversaryThresholdBasisPoints uint16, _feePerBytePerTime *big.Int, _feeSetter common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.Initialize(&_ContractDataLayrServiceManager.TransactOpts, _pauserRegistry, initialOwner, _quorumThresholdBasisPoints, _adversaryThresholdBasisPoints, _feePerBytePerTime, _feeSetter)
}

// Initialize is a paid mutator transaction binding the contract method 0x7bc9d56e.
//
// Solidity: function initialize(address _pauserRegistry, address initialOwner, uint16 _quorumThresholdBasisPoints, uint16 _adversaryThresholdBasisPoints, uint256 _feePerBytePerTime, address _feeSetter) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) Initialize(_pauserRegistry common.Address, initialOwner common.Address, _quorumThresholdBasisPoints uint16, _adversaryThresholdBasisPoints uint16, _feePerBytePerTime *big.Int, _feeSetter common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.Initialize(&_ContractDataLayrServiceManager.TransactOpts, _pauserRegistry, initialOwner, _quorumThresholdBasisPoints, _adversaryThresholdBasisPoints, _feePerBytePerTime, _feeSetter)
}

// Pause is a paid mutator transaction binding the contract method 0x136439dd.
//
// Solidity: function pause(uint256 newPausedStatus) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) Pause(opts *bind.TransactOpts, newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "pause", newPausedStatus)
}

// Pause is a paid mutator transaction binding the contract method 0x136439dd.
//
// Solidity: function pause(uint256 newPausedStatus) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Pause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.Pause(&_ContractDataLayrServiceManager.TransactOpts, newPausedStatus)
}

// Pause is a paid mutator transaction binding the contract method 0x136439dd.
//
// Solidity: function pause(uint256 newPausedStatus) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) Pause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.Pause(&_ContractDataLayrServiceManager.TransactOpts, newPausedStatus)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.RenounceOwnership(&_ContractDataLayrServiceManager.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.RenounceOwnership(&_ContractDataLayrServiceManager.TransactOpts)
}

// SetAdversaryThresholdBasisPoints is a paid mutator transaction binding the contract method 0x516d8616.
//
// Solidity: function setAdversaryThresholdBasisPoints(uint16 _adversaryThresholdBasisPoints) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) SetAdversaryThresholdBasisPoints(opts *bind.TransactOpts, _adversaryThresholdBasisPoints uint16) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "setAdversaryThresholdBasisPoints", _adversaryThresholdBasisPoints)
}

// SetAdversaryThresholdBasisPoints is a paid mutator transaction binding the contract method 0x516d8616.
//
// Solidity: function setAdversaryThresholdBasisPoints(uint16 _adversaryThresholdBasisPoints) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) SetAdversaryThresholdBasisPoints(_adversaryThresholdBasisPoints uint16) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetAdversaryThresholdBasisPoints(&_ContractDataLayrServiceManager.TransactOpts, _adversaryThresholdBasisPoints)
}

// SetAdversaryThresholdBasisPoints is a paid mutator transaction binding the contract method 0x516d8616.
//
// Solidity: function setAdversaryThresholdBasisPoints(uint16 _adversaryThresholdBasisPoints) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) SetAdversaryThresholdBasisPoints(_adversaryThresholdBasisPoints uint16) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetAdversaryThresholdBasisPoints(&_ContractDataLayrServiceManager.TransactOpts, _adversaryThresholdBasisPoints)
}

// SetFeePerBytePerTime is a paid mutator transaction binding the contract method 0x772eefe3.
//
// Solidity: function setFeePerBytePerTime(uint256 _feePerBytePerTime) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) SetFeePerBytePerTime(opts *bind.TransactOpts, _feePerBytePerTime *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "setFeePerBytePerTime", _feePerBytePerTime)
}

// SetFeePerBytePerTime is a paid mutator transaction binding the contract method 0x772eefe3.
//
// Solidity: function setFeePerBytePerTime(uint256 _feePerBytePerTime) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) SetFeePerBytePerTime(_feePerBytePerTime *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetFeePerBytePerTime(&_ContractDataLayrServiceManager.TransactOpts, _feePerBytePerTime)
}

// SetFeePerBytePerTime is a paid mutator transaction binding the contract method 0x772eefe3.
//
// Solidity: function setFeePerBytePerTime(uint256 _feePerBytePerTime) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) SetFeePerBytePerTime(_feePerBytePerTime *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetFeePerBytePerTime(&_ContractDataLayrServiceManager.TransactOpts, _feePerBytePerTime)
}

// SetFeeSetter is a paid mutator transaction binding the contract method 0xb19805af.
//
// Solidity: function setFeeSetter(address _feeSetter) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) SetFeeSetter(opts *bind.TransactOpts, _feeSetter common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "setFeeSetter", _feeSetter)
}

// SetFeeSetter is a paid mutator transaction binding the contract method 0xb19805af.
//
// Solidity: function setFeeSetter(address _feeSetter) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) SetFeeSetter(_feeSetter common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetFeeSetter(&_ContractDataLayrServiceManager.TransactOpts, _feeSetter)
}

// SetFeeSetter is a paid mutator transaction binding the contract method 0xb19805af.
//
// Solidity: function setFeeSetter(address _feeSetter) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) SetFeeSetter(_feeSetter common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetFeeSetter(&_ContractDataLayrServiceManager.TransactOpts, _feeSetter)
}

// SetQuorumThresholdBasisPoints is a paid mutator transaction binding the contract method 0x07dab8b3.
//
// Solidity: function setQuorumThresholdBasisPoints(uint16 _quorumThresholdBasisPoints) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) SetQuorumThresholdBasisPoints(opts *bind.TransactOpts, _quorumThresholdBasisPoints uint16) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "setQuorumThresholdBasisPoints", _quorumThresholdBasisPoints)
}

// SetQuorumThresholdBasisPoints is a paid mutator transaction binding the contract method 0x07dab8b3.
//
// Solidity: function setQuorumThresholdBasisPoints(uint16 _quorumThresholdBasisPoints) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) SetQuorumThresholdBasisPoints(_quorumThresholdBasisPoints uint16) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetQuorumThresholdBasisPoints(&_ContractDataLayrServiceManager.TransactOpts, _quorumThresholdBasisPoints)
}

// SetQuorumThresholdBasisPoints is a paid mutator transaction binding the contract method 0x07dab8b3.
//
// Solidity: function setQuorumThresholdBasisPoints(uint16 _quorumThresholdBasisPoints) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) SetQuorumThresholdBasisPoints(_quorumThresholdBasisPoints uint16) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.SetQuorumThresholdBasisPoints(&_ContractDataLayrServiceManager.TransactOpts, _quorumThresholdBasisPoints)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.TransferOwnership(&_ContractDataLayrServiceManager.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.TransferOwnership(&_ContractDataLayrServiceManager.TransactOpts, newOwner)
}

// Unpause is a paid mutator transaction binding the contract method 0xfabc1cbc.
//
// Solidity: function unpause(uint256 newPausedStatus) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactor) Unpause(opts *bind.TransactOpts, newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.contract.Transact(opts, "unpause", newPausedStatus)
}

// Unpause is a paid mutator transaction binding the contract method 0xfabc1cbc.
//
// Solidity: function unpause(uint256 newPausedStatus) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerSession) Unpause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.Unpause(&_ContractDataLayrServiceManager.TransactOpts, newPausedStatus)
}

// Unpause is a paid mutator transaction binding the contract method 0xfabc1cbc.
//
// Solidity: function unpause(uint256 newPausedStatus) returns()
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerTransactorSession) Unpause(newPausedStatus *big.Int) (*types.Transaction, error) {
	return _ContractDataLayrServiceManager.Contract.Unpause(&_ContractDataLayrServiceManager.TransactOpts, newPausedStatus)
}

// ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdatedIterator is returned from FilterAdversaryThresholdBasisPointsUpdated and is used to iterate over the raw logs and unpacked data for AdversaryThresholdBasisPointsUpdated events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdatedIterator struct {
	Event *ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated represents a AdversaryThresholdBasisPointsUpdated event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated struct {
	AdversaryThresholdBasisPoints uint16
	Raw                           types.Log // Blockchain specific contextual infos
}

// FilterAdversaryThresholdBasisPointsUpdated is a free log retrieval operation binding the contract event 0x1bdc513ac13a36cd49087fef52b034cb5833bd75154db5239f27daa6bde17042.
//
// Solidity: event AdversaryThresholdBasisPointsUpdated(uint16 adversaryThresholdBasisPoints)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterAdversaryThresholdBasisPointsUpdated(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdatedIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "AdversaryThresholdBasisPointsUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdatedIterator{contract: _ContractDataLayrServiceManager.contract, event: "AdversaryThresholdBasisPointsUpdated", logs: logs, sub: sub}, nil
}

// WatchAdversaryThresholdBasisPointsUpdated is a free log subscription operation binding the contract event 0x1bdc513ac13a36cd49087fef52b034cb5833bd75154db5239f27daa6bde17042.
//
// Solidity: event AdversaryThresholdBasisPointsUpdated(uint16 adversaryThresholdBasisPoints)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchAdversaryThresholdBasisPointsUpdated(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "AdversaryThresholdBasisPointsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "AdversaryThresholdBasisPointsUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseAdversaryThresholdBasisPointsUpdated is a log parse operation binding the contract event 0x1bdc513ac13a36cd49087fef52b034cb5833bd75154db5239f27daa6bde17042.
//
// Solidity: event AdversaryThresholdBasisPointsUpdated(uint16 adversaryThresholdBasisPoints)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseAdversaryThresholdBasisPointsUpdated(log types.Log) (*ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated, error) {
	event := new(ContractDataLayrServiceManagerAdversaryThresholdBasisPointsUpdated)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "AdversaryThresholdBasisPointsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerBombVerifierSetIterator is returned from FilterBombVerifierSet and is used to iterate over the raw logs and unpacked data for BombVerifierSet events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerBombVerifierSetIterator struct {
	Event *ContractDataLayrServiceManagerBombVerifierSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerBombVerifierSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerBombVerifierSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerBombVerifierSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerBombVerifierSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerBombVerifierSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerBombVerifierSet represents a BombVerifierSet event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerBombVerifierSet struct {
	PreviousAddress common.Address
	NewAddress      common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterBombVerifierSet is a free log retrieval operation binding the contract event 0x875303dc4b1493d311d0dc6908455605e5fd8deae1190665f8f5a4365c58fe58.
//
// Solidity: event BombVerifierSet(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterBombVerifierSet(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerBombVerifierSetIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "BombVerifierSet")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerBombVerifierSetIterator{contract: _ContractDataLayrServiceManager.contract, event: "BombVerifierSet", logs: logs, sub: sub}, nil
}

// WatchBombVerifierSet is a free log subscription operation binding the contract event 0x875303dc4b1493d311d0dc6908455605e5fd8deae1190665f8f5a4365c58fe58.
//
// Solidity: event BombVerifierSet(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchBombVerifierSet(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerBombVerifierSet) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "BombVerifierSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerBombVerifierSet)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "BombVerifierSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseBombVerifierSet is a log parse operation binding the contract event 0x875303dc4b1493d311d0dc6908455605e5fd8deae1190665f8f5a4365c58fe58.
//
// Solidity: event BombVerifierSet(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseBombVerifierSet(log types.Log) (*ContractDataLayrServiceManagerBombVerifierSet, error) {
	event := new(ContractDataLayrServiceManagerBombVerifierSet)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "BombVerifierSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerConfirmDataStoreIterator is returned from FilterConfirmDataStore and is used to iterate over the raw logs and unpacked data for ConfirmDataStore events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerConfirmDataStoreIterator struct {
	Event *ContractDataLayrServiceManagerConfirmDataStore // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerConfirmDataStoreIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerConfirmDataStore)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerConfirmDataStore)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerConfirmDataStoreIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerConfirmDataStoreIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerConfirmDataStore represents a ConfirmDataStore event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerConfirmDataStore struct {
	DataStoreId uint32
	HeaderHash  [32]byte
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterConfirmDataStore is a free log retrieval operation binding the contract event 0xfbb7f4f1b0b9ad9e75d69d22c364e13089418d86fcb5106792a53046c0fb33aa.
//
// Solidity: event ConfirmDataStore(uint32 dataStoreId, bytes32 headerHash)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterConfirmDataStore(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerConfirmDataStoreIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "ConfirmDataStore")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerConfirmDataStoreIterator{contract: _ContractDataLayrServiceManager.contract, event: "ConfirmDataStore", logs: logs, sub: sub}, nil
}

// WatchConfirmDataStore is a free log subscription operation binding the contract event 0xfbb7f4f1b0b9ad9e75d69d22c364e13089418d86fcb5106792a53046c0fb33aa.
//
// Solidity: event ConfirmDataStore(uint32 dataStoreId, bytes32 headerHash)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchConfirmDataStore(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerConfirmDataStore) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "ConfirmDataStore")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerConfirmDataStore)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "ConfirmDataStore", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseConfirmDataStore is a log parse operation binding the contract event 0xfbb7f4f1b0b9ad9e75d69d22c364e13089418d86fcb5106792a53046c0fb33aa.
//
// Solidity: event ConfirmDataStore(uint32 dataStoreId, bytes32 headerHash)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseConfirmDataStore(log types.Log) (*ContractDataLayrServiceManagerConfirmDataStore, error) {
	event := new(ContractDataLayrServiceManagerConfirmDataStore)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "ConfirmDataStore", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerFeePerBytePerTimeSetIterator is returned from FilterFeePerBytePerTimeSet and is used to iterate over the raw logs and unpacked data for FeePerBytePerTimeSet events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerFeePerBytePerTimeSetIterator struct {
	Event *ContractDataLayrServiceManagerFeePerBytePerTimeSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerFeePerBytePerTimeSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerFeePerBytePerTimeSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerFeePerBytePerTimeSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerFeePerBytePerTimeSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerFeePerBytePerTimeSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerFeePerBytePerTimeSet represents a FeePerBytePerTimeSet event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerFeePerBytePerTimeSet struct {
	PreviousValue *big.Int
	NewValue      *big.Int
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterFeePerBytePerTimeSet is a free log retrieval operation binding the contract event 0xcd1b2c2a220284accd1f9effd811cdecb6beaa4638618b48bbea07ce7ae16996.
//
// Solidity: event FeePerBytePerTimeSet(uint256 previousValue, uint256 newValue)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterFeePerBytePerTimeSet(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerFeePerBytePerTimeSetIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "FeePerBytePerTimeSet")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerFeePerBytePerTimeSetIterator{contract: _ContractDataLayrServiceManager.contract, event: "FeePerBytePerTimeSet", logs: logs, sub: sub}, nil
}

// WatchFeePerBytePerTimeSet is a free log subscription operation binding the contract event 0xcd1b2c2a220284accd1f9effd811cdecb6beaa4638618b48bbea07ce7ae16996.
//
// Solidity: event FeePerBytePerTimeSet(uint256 previousValue, uint256 newValue)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchFeePerBytePerTimeSet(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerFeePerBytePerTimeSet) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "FeePerBytePerTimeSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerFeePerBytePerTimeSet)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "FeePerBytePerTimeSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFeePerBytePerTimeSet is a log parse operation binding the contract event 0xcd1b2c2a220284accd1f9effd811cdecb6beaa4638618b48bbea07ce7ae16996.
//
// Solidity: event FeePerBytePerTimeSet(uint256 previousValue, uint256 newValue)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseFeePerBytePerTimeSet(log types.Log) (*ContractDataLayrServiceManagerFeePerBytePerTimeSet, error) {
	event := new(ContractDataLayrServiceManagerFeePerBytePerTimeSet)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "FeePerBytePerTimeSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerFeeSetterChangedIterator is returned from FilterFeeSetterChanged and is used to iterate over the raw logs and unpacked data for FeeSetterChanged events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerFeeSetterChangedIterator struct {
	Event *ContractDataLayrServiceManagerFeeSetterChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerFeeSetterChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerFeeSetterChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerFeeSetterChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerFeeSetterChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerFeeSetterChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerFeeSetterChanged represents a FeeSetterChanged event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerFeeSetterChanged struct {
	PreviousAddress common.Address
	NewAddress      common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterFeeSetterChanged is a free log retrieval operation binding the contract event 0x774b126b94b3cc801460a024dd575406c3ebf27affd7c36198a53ac6655f056d.
//
// Solidity: event FeeSetterChanged(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterFeeSetterChanged(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerFeeSetterChangedIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "FeeSetterChanged")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerFeeSetterChangedIterator{contract: _ContractDataLayrServiceManager.contract, event: "FeeSetterChanged", logs: logs, sub: sub}, nil
}

// WatchFeeSetterChanged is a free log subscription operation binding the contract event 0x774b126b94b3cc801460a024dd575406c3ebf27affd7c36198a53ac6655f056d.
//
// Solidity: event FeeSetterChanged(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchFeeSetterChanged(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerFeeSetterChanged) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "FeeSetterChanged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerFeeSetterChanged)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "FeeSetterChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFeeSetterChanged is a log parse operation binding the contract event 0x774b126b94b3cc801460a024dd575406c3ebf27affd7c36198a53ac6655f056d.
//
// Solidity: event FeeSetterChanged(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseFeeSetterChanged(log types.Log) (*ContractDataLayrServiceManagerFeeSetterChanged, error) {
	event := new(ContractDataLayrServiceManagerFeeSetterChanged)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "FeeSetterChanged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerInitDataStoreIterator is returned from FilterInitDataStore and is used to iterate over the raw logs and unpacked data for InitDataStore events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerInitDataStoreIterator struct {
	Event *ContractDataLayrServiceManagerInitDataStore // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerInitDataStoreIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerInitDataStore)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerInitDataStore)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerInitDataStoreIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerInitDataStoreIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerInitDataStore represents a InitDataStore event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerInitDataStore struct {
	FeePayer   common.Address
	SearchData IDataLayrServiceManagerDataStoreSearchData
	Header     []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterInitDataStore is a free log retrieval operation binding the contract event 0x25a833fbdbcbd72479fd89b970eb4715d9718313f196da0477220e8cd44425d8.
//
// Solidity: event InitDataStore(address feePayer, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData, bytes header)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterInitDataStore(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerInitDataStoreIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "InitDataStore")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerInitDataStoreIterator{contract: _ContractDataLayrServiceManager.contract, event: "InitDataStore", logs: logs, sub: sub}, nil
}

// WatchInitDataStore is a free log subscription operation binding the contract event 0x25a833fbdbcbd72479fd89b970eb4715d9718313f196da0477220e8cd44425d8.
//
// Solidity: event InitDataStore(address feePayer, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData, bytes header)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchInitDataStore(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerInitDataStore) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "InitDataStore")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerInitDataStore)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "InitDataStore", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitDataStore is a log parse operation binding the contract event 0x25a833fbdbcbd72479fd89b970eb4715d9718313f196da0477220e8cd44425d8.
//
// Solidity: event InitDataStore(address feePayer, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData, bytes header)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseInitDataStore(log types.Log) (*ContractDataLayrServiceManagerInitDataStore, error) {
	event := new(ContractDataLayrServiceManagerInitDataStore)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "InitDataStore", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerInitializedIterator struct {
	Event *ContractDataLayrServiceManagerInitialized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerInitialized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerInitialized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerInitialized represents a Initialized event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterInitialized(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerInitializedIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerInitializedIterator{contract: _ContractDataLayrServiceManager.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerInitialized) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerInitialized)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "Initialized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseInitialized is a log parse operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseInitialized(log types.Log) (*ContractDataLayrServiceManagerInitialized, error) {
	event := new(ContractDataLayrServiceManagerInitialized)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerOwnershipTransferredIterator struct {
	Event *ContractDataLayrServiceManagerOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerOwnershipTransferred represents a OwnershipTransferred event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ContractDataLayrServiceManagerOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerOwnershipTransferredIterator{contract: _ContractDataLayrServiceManager.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerOwnershipTransferred)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseOwnershipTransferred(log types.Log) (*ContractDataLayrServiceManagerOwnershipTransferred, error) {
	event := new(ContractDataLayrServiceManagerOwnershipTransferred)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerPausedIterator is returned from FilterPaused and is used to iterate over the raw logs and unpacked data for Paused events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerPausedIterator struct {
	Event *ContractDataLayrServiceManagerPaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerPausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerPaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerPaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerPausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerPausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerPaused represents a Paused event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerPaused struct {
	Account         common.Address
	NewPausedStatus *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterPaused is a free log retrieval operation binding the contract event 0xab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d.
//
// Solidity: event Paused(address indexed account, uint256 newPausedStatus)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterPaused(opts *bind.FilterOpts, account []common.Address) (*ContractDataLayrServiceManagerPausedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "Paused", accountRule)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerPausedIterator{contract: _ContractDataLayrServiceManager.contract, event: "Paused", logs: logs, sub: sub}, nil
}

// WatchPaused is a free log subscription operation binding the contract event 0xab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d.
//
// Solidity: event Paused(address indexed account, uint256 newPausedStatus)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchPaused(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerPaused, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "Paused", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerPaused)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "Paused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaused is a log parse operation binding the contract event 0xab40a374bc51de372200a8bc981af8c9ecdc08dfdaef0bb6e09f88f3c616ef3d.
//
// Solidity: event Paused(address indexed account, uint256 newPausedStatus)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParsePaused(log types.Log) (*ContractDataLayrServiceManagerPaused, error) {
	event := new(ContractDataLayrServiceManagerPaused)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "Paused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerPaymentManagerSetIterator is returned from FilterPaymentManagerSet and is used to iterate over the raw logs and unpacked data for PaymentManagerSet events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerPaymentManagerSetIterator struct {
	Event *ContractDataLayrServiceManagerPaymentManagerSet // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerPaymentManagerSetIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerPaymentManagerSet)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerPaymentManagerSet)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerPaymentManagerSetIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerPaymentManagerSetIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerPaymentManagerSet represents a PaymentManagerSet event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerPaymentManagerSet struct {
	PreviousAddress common.Address
	NewAddress      common.Address
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterPaymentManagerSet is a free log retrieval operation binding the contract event 0xa3044efb81dffce20bbf49cae117f167852a973364ae504dfade51a8d022c95a.
//
// Solidity: event PaymentManagerSet(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterPaymentManagerSet(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerPaymentManagerSetIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "PaymentManagerSet")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerPaymentManagerSetIterator{contract: _ContractDataLayrServiceManager.contract, event: "PaymentManagerSet", logs: logs, sub: sub}, nil
}

// WatchPaymentManagerSet is a free log subscription operation binding the contract event 0xa3044efb81dffce20bbf49cae117f167852a973364ae504dfade51a8d022c95a.
//
// Solidity: event PaymentManagerSet(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchPaymentManagerSet(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerPaymentManagerSet) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "PaymentManagerSet")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerPaymentManagerSet)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "PaymentManagerSet", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParsePaymentManagerSet is a log parse operation binding the contract event 0xa3044efb81dffce20bbf49cae117f167852a973364ae504dfade51a8d022c95a.
//
// Solidity: event PaymentManagerSet(address previousAddress, address newAddress)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParsePaymentManagerSet(log types.Log) (*ContractDataLayrServiceManagerPaymentManagerSet, error) {
	event := new(ContractDataLayrServiceManagerPaymentManagerSet)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "PaymentManagerSet", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdatedIterator is returned from FilterQuorumThresholdBasisPointsUpdated and is used to iterate over the raw logs and unpacked data for QuorumThresholdBasisPointsUpdated events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdatedIterator struct {
	Event *ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated represents a QuorumThresholdBasisPointsUpdated event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated struct {
	QuorumTHresholdBasisPoints uint16
	Raw                        types.Log // Blockchain specific contextual infos
}

// FilterQuorumThresholdBasisPointsUpdated is a free log retrieval operation binding the contract event 0xae5844e5ca560c940e41aae83424a548a030c790cd14ae00d68c8437bb2e8ec2.
//
// Solidity: event QuorumThresholdBasisPointsUpdated(uint16 quorumTHresholdBasisPoints)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterQuorumThresholdBasisPointsUpdated(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdatedIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "QuorumThresholdBasisPointsUpdated")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdatedIterator{contract: _ContractDataLayrServiceManager.contract, event: "QuorumThresholdBasisPointsUpdated", logs: logs, sub: sub}, nil
}

// WatchQuorumThresholdBasisPointsUpdated is a free log subscription operation binding the contract event 0xae5844e5ca560c940e41aae83424a548a030c790cd14ae00d68c8437bb2e8ec2.
//
// Solidity: event QuorumThresholdBasisPointsUpdated(uint16 quorumTHresholdBasisPoints)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchQuorumThresholdBasisPointsUpdated(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "QuorumThresholdBasisPointsUpdated")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "QuorumThresholdBasisPointsUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseQuorumThresholdBasisPointsUpdated is a log parse operation binding the contract event 0xae5844e5ca560c940e41aae83424a548a030c790cd14ae00d68c8437bb2e8ec2.
//
// Solidity: event QuorumThresholdBasisPointsUpdated(uint16 quorumTHresholdBasisPoints)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseQuorumThresholdBasisPointsUpdated(log types.Log) (*ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated, error) {
	event := new(ContractDataLayrServiceManagerQuorumThresholdBasisPointsUpdated)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "QuorumThresholdBasisPointsUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerSignatoryRecordIterator is returned from FilterSignatoryRecord and is used to iterate over the raw logs and unpacked data for SignatoryRecord events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerSignatoryRecordIterator struct {
	Event *ContractDataLayrServiceManagerSignatoryRecord // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerSignatoryRecordIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerSignatoryRecord)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerSignatoryRecord)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerSignatoryRecordIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerSignatoryRecordIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerSignatoryRecord represents a SignatoryRecord event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerSignatoryRecord struct {
	MsgHash                 [32]byte
	TaskNumber              uint32
	SignedStakeFirstQuorum  *big.Int
	SignedStakeSecondQuorum *big.Int
	PubkeyHashes            [][32]byte
	Raw                     types.Log // Blockchain specific contextual infos
}

// FilterSignatoryRecord is a free log retrieval operation binding the contract event 0x34d57e230be557a52d94166eb9035810e61ac973182a92b09e6b0e99110665a9.
//
// Solidity: event SignatoryRecord(bytes32 msgHash, uint32 taskNumber, uint256 signedStakeFirstQuorum, uint256 signedStakeSecondQuorum, bytes32[] pubkeyHashes)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterSignatoryRecord(opts *bind.FilterOpts) (*ContractDataLayrServiceManagerSignatoryRecordIterator, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "SignatoryRecord")
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerSignatoryRecordIterator{contract: _ContractDataLayrServiceManager.contract, event: "SignatoryRecord", logs: logs, sub: sub}, nil
}

// WatchSignatoryRecord is a free log subscription operation binding the contract event 0x34d57e230be557a52d94166eb9035810e61ac973182a92b09e6b0e99110665a9.
//
// Solidity: event SignatoryRecord(bytes32 msgHash, uint32 taskNumber, uint256 signedStakeFirstQuorum, uint256 signedStakeSecondQuorum, bytes32[] pubkeyHashes)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchSignatoryRecord(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerSignatoryRecord) (event.Subscription, error) {

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "SignatoryRecord")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerSignatoryRecord)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "SignatoryRecord", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseSignatoryRecord is a log parse operation binding the contract event 0x34d57e230be557a52d94166eb9035810e61ac973182a92b09e6b0e99110665a9.
//
// Solidity: event SignatoryRecord(bytes32 msgHash, uint32 taskNumber, uint256 signedStakeFirstQuorum, uint256 signedStakeSecondQuorum, bytes32[] pubkeyHashes)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseSignatoryRecord(log types.Log) (*ContractDataLayrServiceManagerSignatoryRecord, error) {
	event := new(ContractDataLayrServiceManagerSignatoryRecord)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "SignatoryRecord", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ContractDataLayrServiceManagerUnpausedIterator is returned from FilterUnpaused and is used to iterate over the raw logs and unpacked data for Unpaused events raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerUnpausedIterator struct {
	Event *ContractDataLayrServiceManagerUnpaused // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ContractDataLayrServiceManagerUnpausedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ContractDataLayrServiceManagerUnpaused)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ContractDataLayrServiceManagerUnpaused)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ContractDataLayrServiceManagerUnpausedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ContractDataLayrServiceManagerUnpausedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ContractDataLayrServiceManagerUnpaused represents a Unpaused event raised by the ContractDataLayrServiceManager contract.
type ContractDataLayrServiceManagerUnpaused struct {
	Account         common.Address
	NewPausedStatus *big.Int
	Raw             types.Log // Blockchain specific contextual infos
}

// FilterUnpaused is a free log retrieval operation binding the contract event 0x3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c.
//
// Solidity: event Unpaused(address indexed account, uint256 newPausedStatus)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) FilterUnpaused(opts *bind.FilterOpts, account []common.Address) (*ContractDataLayrServiceManagerUnpausedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractDataLayrServiceManager.contract.FilterLogs(opts, "Unpaused", accountRule)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerUnpausedIterator{contract: _ContractDataLayrServiceManager.contract, event: "Unpaused", logs: logs, sub: sub}, nil
}

// WatchUnpaused is a free log subscription operation binding the contract event 0x3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c.
//
// Solidity: event Unpaused(address indexed account, uint256 newPausedStatus)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) WatchUnpaused(opts *bind.WatchOpts, sink chan<- *ContractDataLayrServiceManagerUnpaused, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _ContractDataLayrServiceManager.contract.WatchLogs(opts, "Unpaused", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ContractDataLayrServiceManagerUnpaused)
				if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "Unpaused", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUnpaused is a log parse operation binding the contract event 0x3582d1828e26bf56bd801502bc021ac0bc8afb57c826e4986b45593c8fad389c.
//
// Solidity: event Unpaused(address indexed account, uint256 newPausedStatus)
func (_ContractDataLayrServiceManager *ContractDataLayrServiceManagerFilterer) ParseUnpaused(log types.Log) (*ContractDataLayrServiceManagerUnpaused, error) {
	event := new(ContractDataLayrServiceManagerUnpaused)
	if err := _ContractDataLayrServiceManager.contract.UnpackLog(event, "Unpaused", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
