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

// SequencerFeeVaultMetaData contains all meta data concerning the SequencerFeeVault contract.
var SequencerFeeVaultMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_recipient\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"}],\"name\":\"Withdrawal\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MIN_WITHDRAWAL_AMOUNT\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RECIPIENT\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1FeeWallet\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalProcessed\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"withdraw\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"stateMutability\":\"payable\",\"type\":\"receive\"}]",
	Bin: "0x61012060405234801561001157600080fd5b506040516107033803806107038339810160408190526100309161005d565b678ac7230489e800006080526001600160a01b031660a052600160c081905260e05260006101005261008d565b60006020828403121561006f57600080fd5b81516001600160a01b038116811461008657600080fd5b9392505050565b60805160a05160c05160e0516101005161062c6100d760003960006102560152600061022d015260006102040152600081816087015261017301526000610142015261062c6000f3fe6080604052600436106100695760003560e01c806384411d651161004357806384411d651461010c578063d3e5792b14610130578063d4ff92181461016457600080fd5b80630d9019e1146100755780633ccfd60b146100d357806354fd4d50146100ea57600080fd5b3661007057005b600080fd5b34801561008157600080fd5b506100a97f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020015b60405180910390f35b3480156100df57600080fd5b506100e8610197565b005b3480156100f657600080fd5b506100ff6101fd565b6040516100ca919061040d565b34801561011857600080fd5b5061012260005481565b6040519081526020016100ca565b34801561013c57600080fd5b506101227f000000000000000000000000000000000000000000000000000000000000000081565b34801561017057600080fd5b507f00000000000000000000000000000000000000000000000000000000000000006100a9565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601e60248201527f746869732066756e6374696f6e206e65656420746f20646973637573732e0000604482015260640160405180910390fd5b60606102287f00000000000000000000000000000000000000000000000000000000000000006102a0565b6102517f00000000000000000000000000000000000000000000000000000000000000006102a0565b61027a7f00000000000000000000000000000000000000000000000000000000000000006102a0565b60405160200161028c9392919061045e565b604051602081830303815290604052905090565b6060816000036102e357505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b811561030d57806102f781610503565b91506103069050600a8361056a565b91506102e7565b60008167ffffffffffffffff8111156103285761032861057e565b6040519080825280601f01601f191660200182016040528015610352576020820181803683370190505b5090505b84156103d5576103676001836105ad565b9150610374600a866105c4565b61037f9060306105d8565b60f81b818381518110610394576103946105f0565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a9053506103ce600a8661056a565b9450610356565b949350505050565b60005b838110156103f85781810151838201526020016103e0565b83811115610407576000848401525b50505050565b602081526000825180602084015261042c8160408501602087016103dd565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169190910160400192915050565b600084516104708184602089016103dd565b80830190507f2e0000000000000000000000000000000000000000000000000000000000000080825285516104ac816001850160208a016103dd565b600192019182015283516104c78160028401602088016103dd565b0160020195945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203610534576105346104d4565b5060010190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b6000826105795761057961053b565b500490565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b6000828210156105bf576105bf6104d4565b500390565b6000826105d3576105d361053b565b500690565b600082198211156105eb576105eb6104d4565b500190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fdfea164736f6c634300080f000a",
}

// SequencerFeeVaultABI is the input ABI used to generate the binding from.
// Deprecated: Use SequencerFeeVaultMetaData.ABI instead.
var SequencerFeeVaultABI = SequencerFeeVaultMetaData.ABI

// SequencerFeeVaultBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SequencerFeeVaultMetaData.Bin instead.
var SequencerFeeVaultBin = SequencerFeeVaultMetaData.Bin

// DeploySequencerFeeVault deploys a new Ethereum contract, binding an instance of SequencerFeeVault to it.
func DeploySequencerFeeVault(auth *bind.TransactOpts, backend bind.ContractBackend, _recipient common.Address) (common.Address, *types.Transaction, *SequencerFeeVault, error) {
	parsed, err := SequencerFeeVaultMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SequencerFeeVaultBin), backend, _recipient)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SequencerFeeVault{SequencerFeeVaultCaller: SequencerFeeVaultCaller{contract: contract}, SequencerFeeVaultTransactor: SequencerFeeVaultTransactor{contract: contract}, SequencerFeeVaultFilterer: SequencerFeeVaultFilterer{contract: contract}}, nil
}

// SequencerFeeVault is an auto generated Go binding around an Ethereum contract.
type SequencerFeeVault struct {
	SequencerFeeVaultCaller     // Read-only binding to the contract
	SequencerFeeVaultTransactor // Write-only binding to the contract
	SequencerFeeVaultFilterer   // Log filterer for contract events
}

// SequencerFeeVaultCaller is an auto generated read-only Go binding around an Ethereum contract.
type SequencerFeeVaultCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SequencerFeeVaultTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SequencerFeeVaultTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SequencerFeeVaultFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SequencerFeeVaultFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SequencerFeeVaultSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SequencerFeeVaultSession struct {
	Contract     *SequencerFeeVault // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// SequencerFeeVaultCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SequencerFeeVaultCallerSession struct {
	Contract *SequencerFeeVaultCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// SequencerFeeVaultTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SequencerFeeVaultTransactorSession struct {
	Contract     *SequencerFeeVaultTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// SequencerFeeVaultRaw is an auto generated low-level Go binding around an Ethereum contract.
type SequencerFeeVaultRaw struct {
	Contract *SequencerFeeVault // Generic contract binding to access the raw methods on
}

// SequencerFeeVaultCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SequencerFeeVaultCallerRaw struct {
	Contract *SequencerFeeVaultCaller // Generic read-only contract binding to access the raw methods on
}

// SequencerFeeVaultTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SequencerFeeVaultTransactorRaw struct {
	Contract *SequencerFeeVaultTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSequencerFeeVault creates a new instance of SequencerFeeVault, bound to a specific deployed contract.
func NewSequencerFeeVault(address common.Address, backend bind.ContractBackend) (*SequencerFeeVault, error) {
	contract, err := bindSequencerFeeVault(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SequencerFeeVault{SequencerFeeVaultCaller: SequencerFeeVaultCaller{contract: contract}, SequencerFeeVaultTransactor: SequencerFeeVaultTransactor{contract: contract}, SequencerFeeVaultFilterer: SequencerFeeVaultFilterer{contract: contract}}, nil
}

// NewSequencerFeeVaultCaller creates a new read-only instance of SequencerFeeVault, bound to a specific deployed contract.
func NewSequencerFeeVaultCaller(address common.Address, caller bind.ContractCaller) (*SequencerFeeVaultCaller, error) {
	contract, err := bindSequencerFeeVault(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SequencerFeeVaultCaller{contract: contract}, nil
}

// NewSequencerFeeVaultTransactor creates a new write-only instance of SequencerFeeVault, bound to a specific deployed contract.
func NewSequencerFeeVaultTransactor(address common.Address, transactor bind.ContractTransactor) (*SequencerFeeVaultTransactor, error) {
	contract, err := bindSequencerFeeVault(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SequencerFeeVaultTransactor{contract: contract}, nil
}

// NewSequencerFeeVaultFilterer creates a new log filterer instance of SequencerFeeVault, bound to a specific deployed contract.
func NewSequencerFeeVaultFilterer(address common.Address, filterer bind.ContractFilterer) (*SequencerFeeVaultFilterer, error) {
	contract, err := bindSequencerFeeVault(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SequencerFeeVaultFilterer{contract: contract}, nil
}

// bindSequencerFeeVault binds a generic wrapper to an already deployed contract.
func bindSequencerFeeVault(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SequencerFeeVaultABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SequencerFeeVault *SequencerFeeVaultRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SequencerFeeVault.Contract.SequencerFeeVaultCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SequencerFeeVault *SequencerFeeVaultRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.SequencerFeeVaultTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SequencerFeeVault *SequencerFeeVaultRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.SequencerFeeVaultTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SequencerFeeVault *SequencerFeeVaultCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SequencerFeeVault.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SequencerFeeVault *SequencerFeeVaultTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SequencerFeeVault *SequencerFeeVaultTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.contract.Transact(opts, method, params...)
}

// MINWITHDRAWALAMOUNT is a free data retrieval call binding the contract method 0xd3e5792b.
//
// Solidity: function MIN_WITHDRAWAL_AMOUNT() view returns(uint256)
func (_SequencerFeeVault *SequencerFeeVaultCaller) MINWITHDRAWALAMOUNT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SequencerFeeVault.contract.Call(opts, &out, "MIN_WITHDRAWAL_AMOUNT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MINWITHDRAWALAMOUNT is a free data retrieval call binding the contract method 0xd3e5792b.
//
// Solidity: function MIN_WITHDRAWAL_AMOUNT() view returns(uint256)
func (_SequencerFeeVault *SequencerFeeVaultSession) MINWITHDRAWALAMOUNT() (*big.Int, error) {
	return _SequencerFeeVault.Contract.MINWITHDRAWALAMOUNT(&_SequencerFeeVault.CallOpts)
}

// MINWITHDRAWALAMOUNT is a free data retrieval call binding the contract method 0xd3e5792b.
//
// Solidity: function MIN_WITHDRAWAL_AMOUNT() view returns(uint256)
func (_SequencerFeeVault *SequencerFeeVaultCallerSession) MINWITHDRAWALAMOUNT() (*big.Int, error) {
	return _SequencerFeeVault.Contract.MINWITHDRAWALAMOUNT(&_SequencerFeeVault.CallOpts)
}

// RECIPIENT is a free data retrieval call binding the contract method 0x0d9019e1.
//
// Solidity: function RECIPIENT() view returns(address)
func (_SequencerFeeVault *SequencerFeeVaultCaller) RECIPIENT(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SequencerFeeVault.contract.Call(opts, &out, "RECIPIENT")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RECIPIENT is a free data retrieval call binding the contract method 0x0d9019e1.
//
// Solidity: function RECIPIENT() view returns(address)
func (_SequencerFeeVault *SequencerFeeVaultSession) RECIPIENT() (common.Address, error) {
	return _SequencerFeeVault.Contract.RECIPIENT(&_SequencerFeeVault.CallOpts)
}

// RECIPIENT is a free data retrieval call binding the contract method 0x0d9019e1.
//
// Solidity: function RECIPIENT() view returns(address)
func (_SequencerFeeVault *SequencerFeeVaultCallerSession) RECIPIENT() (common.Address, error) {
	return _SequencerFeeVault.Contract.RECIPIENT(&_SequencerFeeVault.CallOpts)
}

// L1FeeWallet is a free data retrieval call binding the contract method 0xd4ff9218.
//
// Solidity: function l1FeeWallet() view returns(address)
func (_SequencerFeeVault *SequencerFeeVaultCaller) L1FeeWallet(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SequencerFeeVault.contract.Call(opts, &out, "l1FeeWallet")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1FeeWallet is a free data retrieval call binding the contract method 0xd4ff9218.
//
// Solidity: function l1FeeWallet() view returns(address)
func (_SequencerFeeVault *SequencerFeeVaultSession) L1FeeWallet() (common.Address, error) {
	return _SequencerFeeVault.Contract.L1FeeWallet(&_SequencerFeeVault.CallOpts)
}

// L1FeeWallet is a free data retrieval call binding the contract method 0xd4ff9218.
//
// Solidity: function l1FeeWallet() view returns(address)
func (_SequencerFeeVault *SequencerFeeVaultCallerSession) L1FeeWallet() (common.Address, error) {
	return _SequencerFeeVault.Contract.L1FeeWallet(&_SequencerFeeVault.CallOpts)
}

// TotalProcessed is a free data retrieval call binding the contract method 0x84411d65.
//
// Solidity: function totalProcessed() view returns(uint256)
func (_SequencerFeeVault *SequencerFeeVaultCaller) TotalProcessed(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SequencerFeeVault.contract.Call(opts, &out, "totalProcessed")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalProcessed is a free data retrieval call binding the contract method 0x84411d65.
//
// Solidity: function totalProcessed() view returns(uint256)
func (_SequencerFeeVault *SequencerFeeVaultSession) TotalProcessed() (*big.Int, error) {
	return _SequencerFeeVault.Contract.TotalProcessed(&_SequencerFeeVault.CallOpts)
}

// TotalProcessed is a free data retrieval call binding the contract method 0x84411d65.
//
// Solidity: function totalProcessed() view returns(uint256)
func (_SequencerFeeVault *SequencerFeeVaultCallerSession) TotalProcessed() (*big.Int, error) {
	return _SequencerFeeVault.Contract.TotalProcessed(&_SequencerFeeVault.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SequencerFeeVault *SequencerFeeVaultCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SequencerFeeVault.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SequencerFeeVault *SequencerFeeVaultSession) Version() (string, error) {
	return _SequencerFeeVault.Contract.Version(&_SequencerFeeVault.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SequencerFeeVault *SequencerFeeVaultCallerSession) Version() (string, error) {
	return _SequencerFeeVault.Contract.Version(&_SequencerFeeVault.CallOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_SequencerFeeVault *SequencerFeeVaultTransactor) Withdraw(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SequencerFeeVault.contract.Transact(opts, "withdraw")
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_SequencerFeeVault *SequencerFeeVaultSession) Withdraw() (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.Withdraw(&_SequencerFeeVault.TransactOpts)
}

// Withdraw is a paid mutator transaction binding the contract method 0x3ccfd60b.
//
// Solidity: function withdraw() returns()
func (_SequencerFeeVault *SequencerFeeVaultTransactorSession) Withdraw() (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.Withdraw(&_SequencerFeeVault.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_SequencerFeeVault *SequencerFeeVaultTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SequencerFeeVault.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_SequencerFeeVault *SequencerFeeVaultSession) Receive() (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.Receive(&_SequencerFeeVault.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_SequencerFeeVault *SequencerFeeVaultTransactorSession) Receive() (*types.Transaction, error) {
	return _SequencerFeeVault.Contract.Receive(&_SequencerFeeVault.TransactOpts)
}

// SequencerFeeVaultWithdrawalIterator is returned from FilterWithdrawal and is used to iterate over the raw logs and unpacked data for Withdrawal events raised by the SequencerFeeVault contract.
type SequencerFeeVaultWithdrawalIterator struct {
	Event *SequencerFeeVaultWithdrawal // Event containing the contract specifics and raw log

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
func (it *SequencerFeeVaultWithdrawalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SequencerFeeVaultWithdrawal)
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
		it.Event = new(SequencerFeeVaultWithdrawal)
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
func (it *SequencerFeeVaultWithdrawalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SequencerFeeVaultWithdrawalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SequencerFeeVaultWithdrawal represents a Withdrawal event raised by the SequencerFeeVault contract.
type SequencerFeeVaultWithdrawal struct {
	Value *big.Int
	To    common.Address
	From  common.Address
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterWithdrawal is a free log retrieval operation binding the contract event 0xc8a211cc64b6ed1b50595a9fcb1932b6d1e5a6e8ef15b60e5b1f988ea9086bba.
//
// Solidity: event Withdrawal(uint256 value, address to, address from)
func (_SequencerFeeVault *SequencerFeeVaultFilterer) FilterWithdrawal(opts *bind.FilterOpts) (*SequencerFeeVaultWithdrawalIterator, error) {

	logs, sub, err := _SequencerFeeVault.contract.FilterLogs(opts, "Withdrawal")
	if err != nil {
		return nil, err
	}
	return &SequencerFeeVaultWithdrawalIterator{contract: _SequencerFeeVault.contract, event: "Withdrawal", logs: logs, sub: sub}, nil
}

// WatchWithdrawal is a free log subscription operation binding the contract event 0xc8a211cc64b6ed1b50595a9fcb1932b6d1e5a6e8ef15b60e5b1f988ea9086bba.
//
// Solidity: event Withdrawal(uint256 value, address to, address from)
func (_SequencerFeeVault *SequencerFeeVaultFilterer) WatchWithdrawal(opts *bind.WatchOpts, sink chan<- *SequencerFeeVaultWithdrawal) (event.Subscription, error) {

	logs, sub, err := _SequencerFeeVault.contract.WatchLogs(opts, "Withdrawal")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SequencerFeeVaultWithdrawal)
				if err := _SequencerFeeVault.contract.UnpackLog(event, "Withdrawal", log); err != nil {
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

// ParseWithdrawal is a log parse operation binding the contract event 0xc8a211cc64b6ed1b50595a9fcb1932b6d1e5a6e8ef15b60e5b1f988ea9086bba.
//
// Solidity: event Withdrawal(uint256 value, address to, address from)
func (_SequencerFeeVault *SequencerFeeVaultFilterer) ParseWithdrawal(log types.Log) (*SequencerFeeVaultWithdrawal, error) {
	event := new(SequencerFeeVaultWithdrawal)
	if err := _SequencerFeeVault.contract.UnpackLog(event, "Withdrawal", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
