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

// ContractDataLayrServiceManagerStorageMetaData contains all meta data concerning the ContractDataLayrServiceManagerStorage contract.
var ContractDataLayrServiceManagerStorageMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[],\"name\":\"BIP_MULTIPLIER\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"BLOCK_STALE_MEASURE\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"DURATION_SCALE\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MAX_DATASTORE_DURATION\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_DATASTORE_DURATION\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"NUM_DS_PER_BLOCK_PER_DURATION\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"collateralToken\",\"outputs\":[{\"internalType\":\"contractIERC20\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"data\",\"type\":\"bytes\"},{\"components\":[{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"headerHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"durationDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"globalDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint96\",\"name\":\"fee\",\"type\":\"uint96\"},{\"internalType\":\"address\",\"name\":\"confirmer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"signatoryRecordHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIDataLayrServiceManager.DataStoreMetadata\",\"name\":\"metadata\",\"type\":\"tuple\"},{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"internalType\":\"structIDataLayrServiceManager.DataStoreSearchData\",\"name\":\"searchData\",\"type\":\"tuple\"}],\"name\":\"confirmDataStore\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"confirmDataStoreTimeout\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"dataStoreHashesForDurationAtTimestamp\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"name\":\"dataStoreIdToSignatureHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"eigenLayrDelegation\",\"outputs\":[{\"internalType\":\"contractIEigenLayrDelegation\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"feePerBytePerTime\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"operator\",\"type\":\"address\"}],\"name\":\"freezeOperator\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"}],\"name\":\"getDataStoreHashesForDurationAtTimestamp\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"}],\"name\":\"getNumDataStoresForDuration\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"feePayer\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"confirmer\",\"type\":\"address\"},{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"totalOperatorsIndex\",\"type\":\"uint32\"},{\"internalType\":\"bytes\",\"name\":\"header\",\"type\":\"bytes\"}],\"name\":\"initDataStore\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"latestTime\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"log2NumPowersOfTau\",\"outputs\":[{\"internalType\":\"uint48\",\"name\":\"\",\"type\":\"uint48\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"numPowersOfTau\",\"outputs\":[{\"internalType\":\"uint48\",\"name\":\"\",\"type\":\"uint48\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"taskNumber\",\"outputs\":[{\"internalType\":\"uint32\",\"name\":\"\",\"type\":\"uint32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint8\",\"name\":\"duration\",\"type\":\"uint8\"},{\"internalType\":\"uint256\",\"name\":\"timestamp\",\"type\":\"uint256\"},{\"internalType\":\"uint32\",\"name\":\"index\",\"type\":\"uint32\"},{\"components\":[{\"internalType\":\"bytes32\",\"name\":\"headerHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint32\",\"name\":\"durationDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"globalDataStoreId\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"referenceBlockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint32\",\"name\":\"blockNumber\",\"type\":\"uint32\"},{\"internalType\":\"uint96\",\"name\":\"fee\",\"type\":\"uint96\"},{\"internalType\":\"address\",\"name\":\"confirmer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"signatoryRecordHash\",\"type\":\"bytes32\"}],\"internalType\":\"structIDataLayrServiceManager.DataStoreMetadata\",\"name\":\"metadata\",\"type\":\"tuple\"}],\"name\":\"verifyDataStoreMetadata\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"zeroPolynomialCommitmentMerkleRoots\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
}

// ContractDataLayrServiceManagerStorageABI is the input ABI used to generate the binding from.
// Deprecated: Use ContractDataLayrServiceManagerStorageMetaData.ABI instead.
var ContractDataLayrServiceManagerStorageABI = ContractDataLayrServiceManagerStorageMetaData.ABI

// ContractDataLayrServiceManagerStorage is an auto generated Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerStorage struct {
	ContractDataLayrServiceManagerStorageCaller     // Read-only binding to the contract
	ContractDataLayrServiceManagerStorageTransactor // Write-only binding to the contract
	ContractDataLayrServiceManagerStorageFilterer   // Log filterer for contract events
}

// ContractDataLayrServiceManagerStorageCaller is an auto generated read-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerStorageCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractDataLayrServiceManagerStorageTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerStorageTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractDataLayrServiceManagerStorageFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ContractDataLayrServiceManagerStorageFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ContractDataLayrServiceManagerStorageSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ContractDataLayrServiceManagerStorageSession struct {
	Contract     *ContractDataLayrServiceManagerStorage // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                          // Call options to use throughout this session
	TransactOpts bind.TransactOpts                      // Transaction auth options to use throughout this session
}

// ContractDataLayrServiceManagerStorageCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ContractDataLayrServiceManagerStorageCallerSession struct {
	Contract *ContractDataLayrServiceManagerStorageCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                                // Call options to use throughout this session
}

// ContractDataLayrServiceManagerStorageTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ContractDataLayrServiceManagerStorageTransactorSession struct {
	Contract     *ContractDataLayrServiceManagerStorageTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                                // Transaction auth options to use throughout this session
}

// ContractDataLayrServiceManagerStorageRaw is an auto generated low-level Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerStorageRaw struct {
	Contract *ContractDataLayrServiceManagerStorage // Generic contract binding to access the raw methods on
}

// ContractDataLayrServiceManagerStorageCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerStorageCallerRaw struct {
	Contract *ContractDataLayrServiceManagerStorageCaller // Generic read-only contract binding to access the raw methods on
}

// ContractDataLayrServiceManagerStorageTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ContractDataLayrServiceManagerStorageTransactorRaw struct {
	Contract *ContractDataLayrServiceManagerStorageTransactor // Generic write-only contract binding to access the raw methods on
}

// NewContractDataLayrServiceManagerStorage creates a new instance of ContractDataLayrServiceManagerStorage, bound to a specific deployed contract.
func NewContractDataLayrServiceManagerStorage(address common.Address, backend bind.ContractBackend) (*ContractDataLayrServiceManagerStorage, error) {
	contract, err := bindContractDataLayrServiceManagerStorage(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerStorage{ContractDataLayrServiceManagerStorageCaller: ContractDataLayrServiceManagerStorageCaller{contract: contract}, ContractDataLayrServiceManagerStorageTransactor: ContractDataLayrServiceManagerStorageTransactor{contract: contract}, ContractDataLayrServiceManagerStorageFilterer: ContractDataLayrServiceManagerStorageFilterer{contract: contract}}, nil
}

// NewContractDataLayrServiceManagerStorageCaller creates a new read-only instance of ContractDataLayrServiceManagerStorage, bound to a specific deployed contract.
func NewContractDataLayrServiceManagerStorageCaller(address common.Address, caller bind.ContractCaller) (*ContractDataLayrServiceManagerStorageCaller, error) {
	contract, err := bindContractDataLayrServiceManagerStorage(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerStorageCaller{contract: contract}, nil
}

// NewContractDataLayrServiceManagerStorageTransactor creates a new write-only instance of ContractDataLayrServiceManagerStorage, bound to a specific deployed contract.
func NewContractDataLayrServiceManagerStorageTransactor(address common.Address, transactor bind.ContractTransactor) (*ContractDataLayrServiceManagerStorageTransactor, error) {
	contract, err := bindContractDataLayrServiceManagerStorage(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerStorageTransactor{contract: contract}, nil
}

// NewContractDataLayrServiceManagerStorageFilterer creates a new log filterer instance of ContractDataLayrServiceManagerStorage, bound to a specific deployed contract.
func NewContractDataLayrServiceManagerStorageFilterer(address common.Address, filterer bind.ContractFilterer) (*ContractDataLayrServiceManagerStorageFilterer, error) {
	contract, err := bindContractDataLayrServiceManagerStorage(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ContractDataLayrServiceManagerStorageFilterer{contract: contract}, nil
}

// bindContractDataLayrServiceManagerStorage binds a generic wrapper to an already deployed contract.
func bindContractDataLayrServiceManagerStorage(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ContractDataLayrServiceManagerStorageABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractDataLayrServiceManagerStorage.Contract.ContractDataLayrServiceManagerStorageCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ContractDataLayrServiceManagerStorageTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ContractDataLayrServiceManagerStorageTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ContractDataLayrServiceManagerStorage.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.contract.Transact(opts, method, params...)
}

// BIPMULTIPLIER is a free data retrieval call binding the contract method 0xa3c7eaf0.
//
// Solidity: function BIP_MULTIPLIER() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) BIPMULTIPLIER(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "BIP_MULTIPLIER")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BIPMULTIPLIER is a free data retrieval call binding the contract method 0xa3c7eaf0.
//
// Solidity: function BIP_MULTIPLIER() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) BIPMULTIPLIER() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.BIPMULTIPLIER(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// BIPMULTIPLIER is a free data retrieval call binding the contract method 0xa3c7eaf0.
//
// Solidity: function BIP_MULTIPLIER() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) BIPMULTIPLIER() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.BIPMULTIPLIER(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// BLOCKSTALEMEASURE is a free data retrieval call binding the contract method 0x5e8b3f2d.
//
// Solidity: function BLOCK_STALE_MEASURE() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) BLOCKSTALEMEASURE(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "BLOCK_STALE_MEASURE")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// BLOCKSTALEMEASURE is a free data retrieval call binding the contract method 0x5e8b3f2d.
//
// Solidity: function BLOCK_STALE_MEASURE() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) BLOCKSTALEMEASURE() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.BLOCKSTALEMEASURE(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// BLOCKSTALEMEASURE is a free data retrieval call binding the contract method 0x5e8b3f2d.
//
// Solidity: function BLOCK_STALE_MEASURE() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) BLOCKSTALEMEASURE() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.BLOCKSTALEMEASURE(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// DURATIONSCALE is a free data retrieval call binding the contract method 0x31a219c5.
//
// Solidity: function DURATION_SCALE() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) DURATIONSCALE(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "DURATION_SCALE")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DURATIONSCALE is a free data retrieval call binding the contract method 0x31a219c5.
//
// Solidity: function DURATION_SCALE() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) DURATIONSCALE() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.DURATIONSCALE(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// DURATIONSCALE is a free data retrieval call binding the contract method 0x31a219c5.
//
// Solidity: function DURATION_SCALE() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) DURATIONSCALE() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.DURATIONSCALE(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// MAXDATASTOREDURATION is a free data retrieval call binding the contract method 0x578ae5a1.
//
// Solidity: function MAX_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) MAXDATASTOREDURATION(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "MAX_DATASTORE_DURATION")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// MAXDATASTOREDURATION is a free data retrieval call binding the contract method 0x578ae5a1.
//
// Solidity: function MAX_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) MAXDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.MAXDATASTOREDURATION(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// MAXDATASTOREDURATION is a free data retrieval call binding the contract method 0x578ae5a1.
//
// Solidity: function MAX_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) MAXDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.MAXDATASTOREDURATION(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// MINDATASTOREDURATION is a free data retrieval call binding the contract method 0x1fdab6e4.
//
// Solidity: function MIN_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) MINDATASTOREDURATION(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "MIN_DATASTORE_DURATION")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// MINDATASTOREDURATION is a free data retrieval call binding the contract method 0x1fdab6e4.
//
// Solidity: function MIN_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) MINDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.MINDATASTOREDURATION(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// MINDATASTOREDURATION is a free data retrieval call binding the contract method 0x1fdab6e4.
//
// Solidity: function MIN_DATASTORE_DURATION() view returns(uint8)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) MINDATASTOREDURATION() (uint8, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.MINDATASTOREDURATION(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// NUMDSPERBLOCKPERDURATION is a free data retrieval call binding the contract method 0x5f87abbb.
//
// Solidity: function NUM_DS_PER_BLOCK_PER_DURATION() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) NUMDSPERBLOCKPERDURATION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "NUM_DS_PER_BLOCK_PER_DURATION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NUMDSPERBLOCKPERDURATION is a free data retrieval call binding the contract method 0x5f87abbb.
//
// Solidity: function NUM_DS_PER_BLOCK_PER_DURATION() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) NUMDSPERBLOCKPERDURATION() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.NUMDSPERBLOCKPERDURATION(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// NUMDSPERBLOCKPERDURATION is a free data retrieval call binding the contract method 0x5f87abbb.
//
// Solidity: function NUM_DS_PER_BLOCK_PER_DURATION() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) NUMDSPERBLOCKPERDURATION() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.NUMDSPERBLOCKPERDURATION(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// CollateralToken is a free data retrieval call binding the contract method 0xb2016bd4.
//
// Solidity: function collateralToken() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) CollateralToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "collateralToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// CollateralToken is a free data retrieval call binding the contract method 0xb2016bd4.
//
// Solidity: function collateralToken() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) CollateralToken() (common.Address, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.CollateralToken(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// CollateralToken is a free data retrieval call binding the contract method 0xb2016bd4.
//
// Solidity: function collateralToken() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) CollateralToken() (common.Address, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.CollateralToken(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// ConfirmDataStoreTimeout is a free data retrieval call binding the contract method 0x4d53cebb.
//
// Solidity: function confirmDataStoreTimeout() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) ConfirmDataStoreTimeout(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "confirmDataStoreTimeout")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// ConfirmDataStoreTimeout is a free data retrieval call binding the contract method 0x4d53cebb.
//
// Solidity: function confirmDataStoreTimeout() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) ConfirmDataStoreTimeout() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ConfirmDataStoreTimeout(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// ConfirmDataStoreTimeout is a free data retrieval call binding the contract method 0x4d53cebb.
//
// Solidity: function confirmDataStoreTimeout() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) ConfirmDataStoreTimeout() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ConfirmDataStoreTimeout(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// DataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0x1bd2b3cf.
//
// Solidity: function dataStoreHashesForDurationAtTimestamp(uint8 , uint256 , uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) DataStoreHashesForDurationAtTimestamp(opts *bind.CallOpts, arg0 uint8, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "dataStoreHashesForDurationAtTimestamp", arg0, arg1, arg2)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0x1bd2b3cf.
//
// Solidity: function dataStoreHashesForDurationAtTimestamp(uint8 , uint256 , uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) DataStoreHashesForDurationAtTimestamp(arg0 uint8, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.DataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManagerStorage.CallOpts, arg0, arg1, arg2)
}

// DataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0x1bd2b3cf.
//
// Solidity: function dataStoreHashesForDurationAtTimestamp(uint8 , uint256 , uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) DataStoreHashesForDurationAtTimestamp(arg0 uint8, arg1 *big.Int, arg2 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.DataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManagerStorage.CallOpts, arg0, arg1, arg2)
}

// DataStoreIdToSignatureHash is a free data retrieval call binding the contract method 0xfc2c6058.
//
// Solidity: function dataStoreIdToSignatureHash(uint32 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) DataStoreIdToSignatureHash(opts *bind.CallOpts, arg0 uint32) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "dataStoreIdToSignatureHash", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DataStoreIdToSignatureHash is a free data retrieval call binding the contract method 0xfc2c6058.
//
// Solidity: function dataStoreIdToSignatureHash(uint32 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) DataStoreIdToSignatureHash(arg0 uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.DataStoreIdToSignatureHash(&_ContractDataLayrServiceManagerStorage.CallOpts, arg0)
}

// DataStoreIdToSignatureHash is a free data retrieval call binding the contract method 0xfc2c6058.
//
// Solidity: function dataStoreIdToSignatureHash(uint32 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) DataStoreIdToSignatureHash(arg0 uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.DataStoreIdToSignatureHash(&_ContractDataLayrServiceManagerStorage.CallOpts, arg0)
}

// EigenLayrDelegation is a free data retrieval call binding the contract method 0x33d2433a.
//
// Solidity: function eigenLayrDelegation() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) EigenLayrDelegation(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "eigenLayrDelegation")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// EigenLayrDelegation is a free data retrieval call binding the contract method 0x33d2433a.
//
// Solidity: function eigenLayrDelegation() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) EigenLayrDelegation() (common.Address, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.EigenLayrDelegation(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// EigenLayrDelegation is a free data retrieval call binding the contract method 0x33d2433a.
//
// Solidity: function eigenLayrDelegation() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) EigenLayrDelegation() (common.Address, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.EigenLayrDelegation(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// FeePerBytePerTime is a free data retrieval call binding the contract method 0xd21eed4f.
//
// Solidity: function feePerBytePerTime() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) FeePerBytePerTime(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "feePerBytePerTime")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FeePerBytePerTime is a free data retrieval call binding the contract method 0xd21eed4f.
//
// Solidity: function feePerBytePerTime() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) FeePerBytePerTime() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.FeePerBytePerTime(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// FeePerBytePerTime is a free data retrieval call binding the contract method 0xd21eed4f.
//
// Solidity: function feePerBytePerTime() view returns(uint256)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) FeePerBytePerTime() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.FeePerBytePerTime(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// GetDataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0xed82c0ee.
//
// Solidity: function getDataStoreHashesForDurationAtTimestamp(uint8 duration, uint256 timestamp, uint32 index) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) GetDataStoreHashesForDurationAtTimestamp(opts *bind.CallOpts, duration uint8, timestamp *big.Int, index uint32) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "getDataStoreHashesForDurationAtTimestamp", duration, timestamp, index)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// GetDataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0xed82c0ee.
//
// Solidity: function getDataStoreHashesForDurationAtTimestamp(uint8 duration, uint256 timestamp, uint32 index) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) GetDataStoreHashesForDurationAtTimestamp(duration uint8, timestamp *big.Int, index uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.GetDataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManagerStorage.CallOpts, duration, timestamp, index)
}

// GetDataStoreHashesForDurationAtTimestamp is a free data retrieval call binding the contract method 0xed82c0ee.
//
// Solidity: function getDataStoreHashesForDurationAtTimestamp(uint8 duration, uint256 timestamp, uint32 index) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) GetDataStoreHashesForDurationAtTimestamp(duration uint8, timestamp *big.Int, index uint32) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.GetDataStoreHashesForDurationAtTimestamp(&_ContractDataLayrServiceManagerStorage.CallOpts, duration, timestamp, index)
}

// GetNumDataStoresForDuration is a free data retrieval call binding the contract method 0x73441c4e.
//
// Solidity: function getNumDataStoresForDuration(uint8 duration) view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) GetNumDataStoresForDuration(opts *bind.CallOpts, duration uint8) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "getNumDataStoresForDuration", duration)

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// GetNumDataStoresForDuration is a free data retrieval call binding the contract method 0x73441c4e.
//
// Solidity: function getNumDataStoresForDuration(uint8 duration) view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) GetNumDataStoresForDuration(duration uint8) (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.GetNumDataStoresForDuration(&_ContractDataLayrServiceManagerStorage.CallOpts, duration)
}

// GetNumDataStoresForDuration is a free data retrieval call binding the contract method 0x73441c4e.
//
// Solidity: function getNumDataStoresForDuration(uint8 duration) view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) GetNumDataStoresForDuration(duration uint8) (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.GetNumDataStoresForDuration(&_ContractDataLayrServiceManagerStorage.CallOpts, duration)
}

// LatestTime is a free data retrieval call binding the contract method 0x7dfd16d7.
//
// Solidity: function latestTime() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) LatestTime(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "latestTime")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// LatestTime is a free data retrieval call binding the contract method 0x7dfd16d7.
//
// Solidity: function latestTime() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) LatestTime() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.LatestTime(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// LatestTime is a free data retrieval call binding the contract method 0x7dfd16d7.
//
// Solidity: function latestTime() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) LatestTime() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.LatestTime(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// Log2NumPowersOfTau is a free data retrieval call binding the contract method 0xa50017a1.
//
// Solidity: function log2NumPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) Log2NumPowersOfTau(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "log2NumPowersOfTau")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Log2NumPowersOfTau is a free data retrieval call binding the contract method 0xa50017a1.
//
// Solidity: function log2NumPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) Log2NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.Log2NumPowersOfTau(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// Log2NumPowersOfTau is a free data retrieval call binding the contract method 0xa50017a1.
//
// Solidity: function log2NumPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) Log2NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.Log2NumPowersOfTau(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// NumPowersOfTau is a free data retrieval call binding the contract method 0x046bf4a6.
//
// Solidity: function numPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) NumPowersOfTau(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "numPowersOfTau")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// NumPowersOfTau is a free data retrieval call binding the contract method 0x046bf4a6.
//
// Solidity: function numPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.NumPowersOfTau(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// NumPowersOfTau is a free data retrieval call binding the contract method 0x046bf4a6.
//
// Solidity: function numPowersOfTau() view returns(uint48)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) NumPowersOfTau() (*big.Int, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.NumPowersOfTau(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) Owner() (common.Address, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.Owner(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) Owner() (common.Address, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.Owner(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) TaskNumber(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "taskNumber")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) TaskNumber() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.TaskNumber(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// TaskNumber is a free data retrieval call binding the contract method 0x72d18e8d.
//
// Solidity: function taskNumber() view returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) TaskNumber() (uint32, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.TaskNumber(&_ContractDataLayrServiceManagerStorage.CallOpts)
}

// VerifyDataStoreMetadata is a free data retrieval call binding the contract method 0xba4994b1.
//
// Solidity: function verifyDataStoreMetadata(uint8 duration, uint256 timestamp, uint32 index, (bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32) metadata) view returns(bool)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) VerifyDataStoreMetadata(opts *bind.CallOpts, duration uint8, timestamp *big.Int, index uint32, metadata IDataLayrServiceManagerDataStoreMetadata) (bool, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "verifyDataStoreMetadata", duration, timestamp, index, metadata)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// VerifyDataStoreMetadata is a free data retrieval call binding the contract method 0xba4994b1.
//
// Solidity: function verifyDataStoreMetadata(uint8 duration, uint256 timestamp, uint32 index, (bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32) metadata) view returns(bool)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) VerifyDataStoreMetadata(duration uint8, timestamp *big.Int, index uint32, metadata IDataLayrServiceManagerDataStoreMetadata) (bool, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.VerifyDataStoreMetadata(&_ContractDataLayrServiceManagerStorage.CallOpts, duration, timestamp, index, metadata)
}

// VerifyDataStoreMetadata is a free data retrieval call binding the contract method 0xba4994b1.
//
// Solidity: function verifyDataStoreMetadata(uint8 duration, uint256 timestamp, uint32 index, (bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32) metadata) view returns(bool)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) VerifyDataStoreMetadata(duration uint8, timestamp *big.Int, index uint32, metadata IDataLayrServiceManagerDataStoreMetadata) (bool, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.VerifyDataStoreMetadata(&_ContractDataLayrServiceManagerStorage.CallOpts, duration, timestamp, index, metadata)
}

// ZeroPolynomialCommitmentMerkleRoots is a free data retrieval call binding the contract method 0x3367a3fb.
//
// Solidity: function zeroPolynomialCommitmentMerkleRoots(uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCaller) ZeroPolynomialCommitmentMerkleRoots(opts *bind.CallOpts, arg0 *big.Int) ([32]byte, error) {
	var out []interface{}
	err := _ContractDataLayrServiceManagerStorage.contract.Call(opts, &out, "zeroPolynomialCommitmentMerkleRoots", arg0)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ZeroPolynomialCommitmentMerkleRoots is a free data retrieval call binding the contract method 0x3367a3fb.
//
// Solidity: function zeroPolynomialCommitmentMerkleRoots(uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) ZeroPolynomialCommitmentMerkleRoots(arg0 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ZeroPolynomialCommitmentMerkleRoots(&_ContractDataLayrServiceManagerStorage.CallOpts, arg0)
}

// ZeroPolynomialCommitmentMerkleRoots is a free data retrieval call binding the contract method 0x3367a3fb.
//
// Solidity: function zeroPolynomialCommitmentMerkleRoots(uint256 ) view returns(bytes32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageCallerSession) ZeroPolynomialCommitmentMerkleRoots(arg0 *big.Int) ([32]byte, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ZeroPolynomialCommitmentMerkleRoots(&_ContractDataLayrServiceManagerStorage.CallOpts, arg0)
}

// ConfirmDataStore is a paid mutator transaction binding the contract method 0x58942e73.
//
// Solidity: function confirmDataStore(bytes data, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData) returns()
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactor) ConfirmDataStore(opts *bind.TransactOpts, data []byte, searchData IDataLayrServiceManagerDataStoreSearchData) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.contract.Transact(opts, "confirmDataStore", data, searchData)
}

// ConfirmDataStore is a paid mutator transaction binding the contract method 0x58942e73.
//
// Solidity: function confirmDataStore(bytes data, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData) returns()
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) ConfirmDataStore(data []byte, searchData IDataLayrServiceManagerDataStoreSearchData) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ConfirmDataStore(&_ContractDataLayrServiceManagerStorage.TransactOpts, data, searchData)
}

// ConfirmDataStore is a paid mutator transaction binding the contract method 0x58942e73.
//
// Solidity: function confirmDataStore(bytes data, ((bytes32,uint32,uint32,uint32,uint32,uint96,address,bytes32),uint8,uint256,uint32) searchData) returns()
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactorSession) ConfirmDataStore(data []byte, searchData IDataLayrServiceManagerDataStoreSearchData) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.ConfirmDataStore(&_ContractDataLayrServiceManagerStorage.TransactOpts, data, searchData)
}

// FreezeOperator is a paid mutator transaction binding the contract method 0x38c8ee64.
//
// Solidity: function freezeOperator(address operator) returns()
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactor) FreezeOperator(opts *bind.TransactOpts, operator common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.contract.Transact(opts, "freezeOperator", operator)
}

// FreezeOperator is a paid mutator transaction binding the contract method 0x38c8ee64.
//
// Solidity: function freezeOperator(address operator) returns()
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) FreezeOperator(operator common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.FreezeOperator(&_ContractDataLayrServiceManagerStorage.TransactOpts, operator)
}

// FreezeOperator is a paid mutator transaction binding the contract method 0x38c8ee64.
//
// Solidity: function freezeOperator(address operator) returns()
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactorSession) FreezeOperator(operator common.Address) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.FreezeOperator(&_ContractDataLayrServiceManagerStorage.TransactOpts, operator)
}

// InitDataStore is a paid mutator transaction binding the contract method 0xdcf49ea7.
//
// Solidity: function initDataStore(address feePayer, address confirmer, uint8 duration, uint32 blockNumber, uint32 totalOperatorsIndex, bytes header) returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactor) InitDataStore(opts *bind.TransactOpts, feePayer common.Address, confirmer common.Address, duration uint8, blockNumber uint32, totalOperatorsIndex uint32, header []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.contract.Transact(opts, "initDataStore", feePayer, confirmer, duration, blockNumber, totalOperatorsIndex, header)
}

// InitDataStore is a paid mutator transaction binding the contract method 0xdcf49ea7.
//
// Solidity: function initDataStore(address feePayer, address confirmer, uint8 duration, uint32 blockNumber, uint32 totalOperatorsIndex, bytes header) returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageSession) InitDataStore(feePayer common.Address, confirmer common.Address, duration uint8, blockNumber uint32, totalOperatorsIndex uint32, header []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.InitDataStore(&_ContractDataLayrServiceManagerStorage.TransactOpts, feePayer, confirmer, duration, blockNumber, totalOperatorsIndex, header)
}

// InitDataStore is a paid mutator transaction binding the contract method 0xdcf49ea7.
//
// Solidity: function initDataStore(address feePayer, address confirmer, uint8 duration, uint32 blockNumber, uint32 totalOperatorsIndex, bytes header) returns(uint32)
func (_ContractDataLayrServiceManagerStorage *ContractDataLayrServiceManagerStorageTransactorSession) InitDataStore(feePayer common.Address, confirmer common.Address, duration uint8, blockNumber uint32, totalOperatorsIndex uint32, header []byte) (*types.Transaction, error) {
	return _ContractDataLayrServiceManagerStorage.Contract.InitDataStore(&_ContractDataLayrServiceManagerStorage.TransactOpts, feePayer, confirmer, duration, blockNumber, totalOperatorsIndex, header)
}
