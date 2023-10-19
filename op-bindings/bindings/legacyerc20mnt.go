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
	_ = abi.ConvertType
)

// LegacyERC20MNTMetaData contains all meta data concerning the LegacyERC20MNT contract.
var LegacyERC20MNTMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_l1mnt\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Approval\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Burn\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amount\",\"type\":\"uint256\"}],\"name\":\"Mint\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"to\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"value\",\"type\":\"uint256\"}],\"name\":\"Transfer\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"BRIDGE\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"REMOTE_TOKEN\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"spender\",\"type\":\"address\"}],\"name\":\"allowance\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"approve\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_who\",\"type\":\"address\"}],\"name\":\"balanceOf\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"bridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"burn\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"decimals\",\"outputs\":[{\"internalType\":\"uint8\",\"name\":\"\",\"type\":\"uint8\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"decreaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"increaseAllowance\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1Token\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l2Bridge\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"mint\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"name\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"remoteToken\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes4\",\"name\":\"_interfaceId\",\"type\":\"bytes4\"}],\"name\":\"supportsInterface\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbol\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"totalSupply\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transfer\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"transferFrom\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6101206040523480156200001257600080fd5b506040516200112b3803806200112b8339810160408190526200003591620000de565b734200000000000000000000000000000000000010816040518060400160405280600c81526020016b26b0b73a3632902a37b5b2b760a11b8152506040518060400160405280600381526020016213539560ea1b815250600160008084848160039081620000a49190620001b5565b506004620000b38282620001b5565b50505060809290925260a05260c05250506001600160a01b0390811660e05216610100525062000281565b600060208284031215620000f157600080fd5b81516001600160a01b03811681146200010957600080fd5b9392505050565b634e487b7160e01b600052604160045260246000fd5b600181811c908216806200013b57607f821691505b6020821081036200015c57634e487b7160e01b600052602260045260246000fd5b50919050565b601f821115620001b057600081815260208120601f850160051c810160208610156200018b5750805b601f850160051c820191505b81811015620001ac5782815560010162000197565b5050505b505050565b81516001600160401b03811115620001d157620001d162000110565b620001e981620001e2845462000126565b8462000162565b602080601f831160018114620002215760008415620002085750858301515b600019600386901b1c1916600185901b178555620001ac565b600085815260208120601f198616915b82811015620002525788860151825594840194600190910190840162000231565b5085821015620002715787850151600019600388901b60f8161c191681555b5050505050600190811b01905550565b60805160a05160c05160e05161010051610e57620002d4600039600081816102e7015261037c0152600081816101a9015261030d0152600061078201526000610759015260006107300152610e576000f3fe608060405234801561001057600080fd5b50600436106101775760003560e01c806370a08231116100d8578063ae1f6aaf1161008c578063dd62ed3e11610066578063dd62ed3e14610331578063e78cea92146102e5578063ee9a31a21461037757600080fd5b8063ae1f6aaf146102e5578063c01e1bd61461030b578063d6c0b2c41461030b57600080fd5b80639dc29fac116100bd5780639dc29fac146102ac578063a457c2d7146102bf578063a9059cbb146102d257600080fd5b806370a082311461027c57806395d89b41146102a457600080fd5b806323b872dd1161012f5780633950935111610114578063395093511461024c57806340c10f191461025f57806354fd4d501461027457600080fd5b806323b872dd1461022a578063313ce5671461023d57600080fd5b806306fdde031161016057806306fdde03146101f0578063095ea7b31461020557806318160ddd1461021857600080fd5b806301ffc9a71461017c578063033964be146101a4575b600080fd5b61018f61018a366004610a8f565b61039e565b60405190151581526020015b60405180910390f35b6101cb7f000000000000000000000000000000000000000000000000000000000000000081565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200161019b565b6101f861048f565b60405161019b9190610b08565b61018f610213366004610b82565b610521565b6002545b60405190815260200161019b565b61018f610238366004610bac565b6105b1565b6040516012815260200161019b565b61018f61025a366004610b82565b61063c565b61027261026d366004610b82565b6106c7565b005b6101f8610729565b61021c61028a366004610be8565b73ffffffffffffffffffffffffffffffffffffffff163190565b6101f86107cc565b6102726102ba366004610b82565b6107db565b61018f6102cd366004610b82565b61083d565b61018f6102e0366004610b82565b6108c8565b7f00000000000000000000000000000000000000000000000000000000000000006101cb565b7f00000000000000000000000000000000000000000000000000000000000000006101cb565b61021c61033f366004610c03565b73ffffffffffffffffffffffffffffffffffffffff918216600090815260016020908152604080832093909416825291909152205490565b6101cb7f000000000000000000000000000000000000000000000000000000000000000081565b60007f01ffc9a7000000000000000000000000000000000000000000000000000000007f1d1d8b63000000000000000000000000000000000000000000000000000000007fec4fc8e3000000000000000000000000000000000000000000000000000000007fffffffff00000000000000000000000000000000000000000000000000000000851683148061045757507fffffffff00000000000000000000000000000000000000000000000000000000858116908316145b8061048657507fffffffff00000000000000000000000000000000000000000000000000000000858116908216145b95945050505050565b60606003805461049e90610c36565b80601f01602080910402602001604051908101604052809291908181526020018280546104ca90610c36565b80156105175780601f106104ec57610100808354040283529160200191610517565b820191906000526020600020905b8154815290600101906020018083116104fa57829003601f168201915b5050505050905090565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602360248201527f4c656761637945524332304d4e543a20617070726f766520697320646973616260448201527f6c6564000000000000000000000000000000000000000000000000000000000060648201526000906084015b60405180910390fd5b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602860248201527f4c656761637945524332304d4e543a207472616e7366657246726f6d2069732060448201527f64697361626c656400000000000000000000000000000000000000000000000060648201526000906084016105a8565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f4c656761637945524332304d4e543a20696e637265617365416c6c6f77616e6360448201527f652069732064697361626c65640000000000000000000000000000000000000060648201526000906084016105a8565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4c656761637945524332304d4e543a206d696e742069732064697361626c656460448201526064016105a8565b60606107547f0000000000000000000000000000000000000000000000000000000000000000610952565b61077d7f0000000000000000000000000000000000000000000000000000000000000000610952565b6107a67f0000000000000000000000000000000000000000000000000000000000000000610952565b6040516020016107b893929190610c89565b604051602081830303815290604052905090565b60606004805461049e90610c36565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4c656761637945524332304d4e543a206275726e2069732064697361626c656460448201526064016105a8565b6040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f4c656761637945524332304d4e543a206465637265617365416c6c6f77616e6360448201527f652069732064697361626c65640000000000000000000000000000000000000060648201526000906084016105a8565b6040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526024808201527f4c656761637945524332304d4e543a207472616e73666572206973206469736160448201527f626c65640000000000000000000000000000000000000000000000000000000060648201526000906084016105a8565b60608160000361099557505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b81156109bf57806109a981610d2e565b91506109b89050600a83610d95565b9150610999565b60008167ffffffffffffffff8111156109da576109da610da9565b6040519080825280601f01601f191660200182016040528015610a04576020820181803683370190505b5090505b8415610a8757610a19600183610dd8565b9150610a26600a86610def565b610a31906030610e03565b60f81b818381518110610a4657610a46610e1b565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350610a80600a86610d95565b9450610a08565b949350505050565b600060208284031215610aa157600080fd5b81357fffffffff0000000000000000000000000000000000000000000000000000000081168114610ad157600080fd5b9392505050565b60005b83811015610af3578181015183820152602001610adb565b83811115610b02576000848401525b50505050565b6020815260008251806020840152610b27816040850160208701610ad8565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169190910160400192915050565b803573ffffffffffffffffffffffffffffffffffffffff81168114610b7d57600080fd5b919050565b60008060408385031215610b9557600080fd5b610b9e83610b59565b946020939093013593505050565b600080600060608486031215610bc157600080fd5b610bca84610b59565b9250610bd860208501610b59565b9150604084013590509250925092565b600060208284031215610bfa57600080fd5b610ad182610b59565b60008060408385031215610c1657600080fd5b610c1f83610b59565b9150610c2d60208401610b59565b90509250929050565b600181811c90821680610c4a57607f821691505b602082108103610c83577f4e487b7100000000000000000000000000000000000000000000000000000000600052602260045260246000fd5b50919050565b60008451610c9b818460208901610ad8565b80830190507f2e000000000000000000000000000000000000000000000000000000000000008082528551610cd7816001850160208a01610ad8565b60019201918201528351610cf2816002840160208801610ad8565b0160020195945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203610d5f57610d5f610cff565b5060010190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600082610da457610da4610d66565b500490565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600082821015610dea57610dea610cff565b500390565b600082610dfe57610dfe610d66565b500690565b60008219821115610e1657610e16610cff565b500190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fdfea164736f6c634300080f000a",
}

// LegacyERC20MNTABI is the input ABI used to generate the binding from.
// Deprecated: Use LegacyERC20MNTMetaData.ABI instead.
var LegacyERC20MNTABI = LegacyERC20MNTMetaData.ABI

// LegacyERC20MNTBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use LegacyERC20MNTMetaData.Bin instead.
var LegacyERC20MNTBin = LegacyERC20MNTMetaData.Bin

// DeployLegacyERC20MNT deploys a new Ethereum contract, binding an instance of LegacyERC20MNT to it.
func DeployLegacyERC20MNT(auth *bind.TransactOpts, backend bind.ContractBackend, _l1mnt common.Address) (common.Address, *types.Transaction, *LegacyERC20MNT, error) {
	parsed, err := LegacyERC20MNTMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(LegacyERC20MNTBin), backend, _l1mnt)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &LegacyERC20MNT{LegacyERC20MNTCaller: LegacyERC20MNTCaller{contract: contract}, LegacyERC20MNTTransactor: LegacyERC20MNTTransactor{contract: contract}, LegacyERC20MNTFilterer: LegacyERC20MNTFilterer{contract: contract}}, nil
}

// LegacyERC20MNT is an auto generated Go binding around an Ethereum contract.
type LegacyERC20MNT struct {
	LegacyERC20MNTCaller     // Read-only binding to the contract
	LegacyERC20MNTTransactor // Write-only binding to the contract
	LegacyERC20MNTFilterer   // Log filterer for contract events
}

// LegacyERC20MNTCaller is an auto generated read-only Go binding around an Ethereum contract.
type LegacyERC20MNTCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LegacyERC20MNTTransactor is an auto generated write-only Go binding around an Ethereum contract.
type LegacyERC20MNTTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LegacyERC20MNTFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type LegacyERC20MNTFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// LegacyERC20MNTSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type LegacyERC20MNTSession struct {
	Contract     *LegacyERC20MNT   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// LegacyERC20MNTCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type LegacyERC20MNTCallerSession struct {
	Contract *LegacyERC20MNTCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// LegacyERC20MNTTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type LegacyERC20MNTTransactorSession struct {
	Contract     *LegacyERC20MNTTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// LegacyERC20MNTRaw is an auto generated low-level Go binding around an Ethereum contract.
type LegacyERC20MNTRaw struct {
	Contract *LegacyERC20MNT // Generic contract binding to access the raw methods on
}

// LegacyERC20MNTCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type LegacyERC20MNTCallerRaw struct {
	Contract *LegacyERC20MNTCaller // Generic read-only contract binding to access the raw methods on
}

// LegacyERC20MNTTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type LegacyERC20MNTTransactorRaw struct {
	Contract *LegacyERC20MNTTransactor // Generic write-only contract binding to access the raw methods on
}

// NewLegacyERC20MNT creates a new instance of LegacyERC20MNT, bound to a specific deployed contract.
func NewLegacyERC20MNT(address common.Address, backend bind.ContractBackend) (*LegacyERC20MNT, error) {
	contract, err := bindLegacyERC20MNT(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNT{LegacyERC20MNTCaller: LegacyERC20MNTCaller{contract: contract}, LegacyERC20MNTTransactor: LegacyERC20MNTTransactor{contract: contract}, LegacyERC20MNTFilterer: LegacyERC20MNTFilterer{contract: contract}}, nil
}

// NewLegacyERC20MNTCaller creates a new read-only instance of LegacyERC20MNT, bound to a specific deployed contract.
func NewLegacyERC20MNTCaller(address common.Address, caller bind.ContractCaller) (*LegacyERC20MNTCaller, error) {
	contract, err := bindLegacyERC20MNT(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNTCaller{contract: contract}, nil
}

// NewLegacyERC20MNTTransactor creates a new write-only instance of LegacyERC20MNT, bound to a specific deployed contract.
func NewLegacyERC20MNTTransactor(address common.Address, transactor bind.ContractTransactor) (*LegacyERC20MNTTransactor, error) {
	contract, err := bindLegacyERC20MNT(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNTTransactor{contract: contract}, nil
}

// NewLegacyERC20MNTFilterer creates a new log filterer instance of LegacyERC20MNT, bound to a specific deployed contract.
func NewLegacyERC20MNTFilterer(address common.Address, filterer bind.ContractFilterer) (*LegacyERC20MNTFilterer, error) {
	contract, err := bindLegacyERC20MNT(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNTFilterer{contract: contract}, nil
}

// bindLegacyERC20MNT binds a generic wrapper to an already deployed contract.
func bindLegacyERC20MNT(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := LegacyERC20MNTMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LegacyERC20MNT *LegacyERC20MNTRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LegacyERC20MNT.Contract.LegacyERC20MNTCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LegacyERC20MNT *LegacyERC20MNTRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.LegacyERC20MNTTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LegacyERC20MNT *LegacyERC20MNTRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.LegacyERC20MNTTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_LegacyERC20MNT *LegacyERC20MNTCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _LegacyERC20MNT.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_LegacyERC20MNT *LegacyERC20MNTTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_LegacyERC20MNT *LegacyERC20MNTTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.contract.Transact(opts, method, params...)
}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) BRIDGE(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "BRIDGE")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTSession) BRIDGE() (common.Address, error) {
	return _LegacyERC20MNT.Contract.BRIDGE(&_LegacyERC20MNT.CallOpts)
}

// BRIDGE is a free data retrieval call binding the contract method 0xee9a31a2.
//
// Solidity: function BRIDGE() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) BRIDGE() (common.Address, error) {
	return _LegacyERC20MNT.Contract.BRIDGE(&_LegacyERC20MNT.CallOpts)
}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) REMOTETOKEN(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "REMOTE_TOKEN")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTSession) REMOTETOKEN() (common.Address, error) {
	return _LegacyERC20MNT.Contract.REMOTETOKEN(&_LegacyERC20MNT.CallOpts)
}

// REMOTETOKEN is a free data retrieval call binding the contract method 0x033964be.
//
// Solidity: function REMOTE_TOKEN() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) REMOTETOKEN() (common.Address, error) {
	return _LegacyERC20MNT.Contract.REMOTETOKEN(&_LegacyERC20MNT.CallOpts)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "allowance", owner, spender)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _LegacyERC20MNT.Contract.Allowance(&_LegacyERC20MNT.CallOpts, owner, spender)
}

// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
//
// Solidity: function allowance(address owner, address spender) view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) Allowance(owner common.Address, spender common.Address) (*big.Int, error) {
	return _LegacyERC20MNT.Contract.Allowance(&_LegacyERC20MNT.CallOpts, owner, spender)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _who) view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) BalanceOf(opts *bind.CallOpts, _who common.Address) (*big.Int, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "balanceOf", _who)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _who) view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTSession) BalanceOf(_who common.Address) (*big.Int, error) {
	return _LegacyERC20MNT.Contract.BalanceOf(&_LegacyERC20MNT.CallOpts, _who)
}

// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
//
// Solidity: function balanceOf(address _who) view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) BalanceOf(_who common.Address) (*big.Int, error) {
	return _LegacyERC20MNT.Contract.BalanceOf(&_LegacyERC20MNT.CallOpts, _who)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Bridge() (common.Address, error) {
	return _LegacyERC20MNT.Contract.Bridge(&_LegacyERC20MNT.CallOpts)
}

// Bridge is a free data retrieval call binding the contract method 0xe78cea92.
//
// Solidity: function bridge() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) Bridge() (common.Address, error) {
	return _LegacyERC20MNT.Contract.Bridge(&_LegacyERC20MNT.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) Decimals(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Decimals() (uint8, error) {
	return _LegacyERC20MNT.Contract.Decimals(&_LegacyERC20MNT.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() view returns(uint8)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) Decimals() (uint8, error) {
	return _LegacyERC20MNT.Contract.Decimals(&_LegacyERC20MNT.CallOpts)
}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) L1Token(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "l1Token")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTSession) L1Token() (common.Address, error) {
	return _LegacyERC20MNT.Contract.L1Token(&_LegacyERC20MNT.CallOpts)
}

// L1Token is a free data retrieval call binding the contract method 0xc01e1bd6.
//
// Solidity: function l1Token() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) L1Token() (common.Address, error) {
	return _LegacyERC20MNT.Contract.L1Token(&_LegacyERC20MNT.CallOpts)
}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) L2Bridge(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "l2Bridge")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTSession) L2Bridge() (common.Address, error) {
	return _LegacyERC20MNT.Contract.L2Bridge(&_LegacyERC20MNT.CallOpts)
}

// L2Bridge is a free data retrieval call binding the contract method 0xae1f6aaf.
//
// Solidity: function l2Bridge() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) L2Bridge() (common.Address, error) {
	return _LegacyERC20MNT.Contract.L2Bridge(&_LegacyERC20MNT.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) Name(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "name")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Name() (string, error) {
	return _LegacyERC20MNT.Contract.Name(&_LegacyERC20MNT.CallOpts)
}

// Name is a free data retrieval call binding the contract method 0x06fdde03.
//
// Solidity: function name() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) Name() (string, error) {
	return _LegacyERC20MNT.Contract.Name(&_LegacyERC20MNT.CallOpts)
}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) RemoteToken(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "remoteToken")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTSession) RemoteToken() (common.Address, error) {
	return _LegacyERC20MNT.Contract.RemoteToken(&_LegacyERC20MNT.CallOpts)
}

// RemoteToken is a free data retrieval call binding the contract method 0xd6c0b2c4.
//
// Solidity: function remoteToken() view returns(address)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) RemoteToken() (common.Address, error) {
	return _LegacyERC20MNT.Contract.RemoteToken(&_LegacyERC20MNT.CallOpts)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) SupportsInterface(opts *bind.CallOpts, _interfaceId [4]byte) (bool, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "supportsInterface", _interfaceId)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _LegacyERC20MNT.Contract.SupportsInterface(&_LegacyERC20MNT.CallOpts, _interfaceId)
}

// SupportsInterface is a free data retrieval call binding the contract method 0x01ffc9a7.
//
// Solidity: function supportsInterface(bytes4 _interfaceId) pure returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) SupportsInterface(_interfaceId [4]byte) (bool, error) {
	return _LegacyERC20MNT.Contract.SupportsInterface(&_LegacyERC20MNT.CallOpts, _interfaceId)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) Symbol(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "symbol")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Symbol() (string, error) {
	return _LegacyERC20MNT.Contract.Symbol(&_LegacyERC20MNT.CallOpts)
}

// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
//
// Solidity: function symbol() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) Symbol() (string, error) {
	return _LegacyERC20MNT.Contract.Symbol(&_LegacyERC20MNT.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) TotalSupply(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "totalSupply")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTSession) TotalSupply() (*big.Int, error) {
	return _LegacyERC20MNT.Contract.TotalSupply(&_LegacyERC20MNT.CallOpts)
}

// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
//
// Solidity: function totalSupply() view returns(uint256)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) TotalSupply() (*big.Int, error) {
	return _LegacyERC20MNT.Contract.TotalSupply(&_LegacyERC20MNT.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _LegacyERC20MNT.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Version() (string, error) {
	return _LegacyERC20MNT.Contract.Version(&_LegacyERC20MNT.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_LegacyERC20MNT *LegacyERC20MNTCallerSession) Version() (string, error) {
	return _LegacyERC20MNT.Contract.Version(&_LegacyERC20MNT.CallOpts)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactor) Approve(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.contract.Transact(opts, "approve", arg0, arg1)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Approve(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Approve(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// Approve is a paid mutator transaction binding the contract method 0x095ea7b3.
//
// Solidity: function approve(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactorSession) Approve(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Approve(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address , uint256 ) returns()
func (_LegacyERC20MNT *LegacyERC20MNTTransactor) Burn(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.contract.Transact(opts, "burn", arg0, arg1)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address , uint256 ) returns()
func (_LegacyERC20MNT *LegacyERC20MNTSession) Burn(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Burn(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// Burn is a paid mutator transaction binding the contract method 0x9dc29fac.
//
// Solidity: function burn(address , uint256 ) returns()
func (_LegacyERC20MNT *LegacyERC20MNTTransactorSession) Burn(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Burn(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactor) DecreaseAllowance(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.contract.Transact(opts, "decreaseAllowance", arg0, arg1)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTSession) DecreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.DecreaseAllowance(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// DecreaseAllowance is a paid mutator transaction binding the contract method 0xa457c2d7.
//
// Solidity: function decreaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactorSession) DecreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.DecreaseAllowance(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactor) IncreaseAllowance(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.contract.Transact(opts, "increaseAllowance", arg0, arg1)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTSession) IncreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.IncreaseAllowance(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// IncreaseAllowance is a paid mutator transaction binding the contract method 0x39509351.
//
// Solidity: function increaseAllowance(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactorSession) IncreaseAllowance(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.IncreaseAllowance(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address , uint256 ) returns()
func (_LegacyERC20MNT *LegacyERC20MNTTransactor) Mint(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.contract.Transact(opts, "mint", arg0, arg1)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address , uint256 ) returns()
func (_LegacyERC20MNT *LegacyERC20MNTSession) Mint(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Mint(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// Mint is a paid mutator transaction binding the contract method 0x40c10f19.
//
// Solidity: function mint(address , uint256 ) returns()
func (_LegacyERC20MNT *LegacyERC20MNTTransactorSession) Mint(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Mint(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactor) Transfer(opts *bind.TransactOpts, arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.contract.Transact(opts, "transfer", arg0, arg1)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTSession) Transfer(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Transfer(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// Transfer is a paid mutator transaction binding the contract method 0xa9059cbb.
//
// Solidity: function transfer(address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactorSession) Transfer(arg0 common.Address, arg1 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.Transfer(&_LegacyERC20MNT.TransactOpts, arg0, arg1)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address , address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactor) TransferFrom(opts *bind.TransactOpts, arg0 common.Address, arg1 common.Address, arg2 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.contract.Transact(opts, "transferFrom", arg0, arg1, arg2)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address , address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTSession) TransferFrom(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.TransferFrom(&_LegacyERC20MNT.TransactOpts, arg0, arg1, arg2)
}

// TransferFrom is a paid mutator transaction binding the contract method 0x23b872dd.
//
// Solidity: function transferFrom(address , address , uint256 ) returns(bool)
func (_LegacyERC20MNT *LegacyERC20MNTTransactorSession) TransferFrom(arg0 common.Address, arg1 common.Address, arg2 *big.Int) (*types.Transaction, error) {
	return _LegacyERC20MNT.Contract.TransferFrom(&_LegacyERC20MNT.TransactOpts, arg0, arg1, arg2)
}

// LegacyERC20MNTApprovalIterator is returned from FilterApproval and is used to iterate over the raw logs and unpacked data for Approval events raised by the LegacyERC20MNT contract.
type LegacyERC20MNTApprovalIterator struct {
	Event *LegacyERC20MNTApproval // Event containing the contract specifics and raw log

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
func (it *LegacyERC20MNTApprovalIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20MNTApproval)
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
		it.Event = new(LegacyERC20MNTApproval)
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
func (it *LegacyERC20MNTApprovalIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20MNTApprovalIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20MNTApproval represents a Approval event raised by the LegacyERC20MNT contract.
type LegacyERC20MNTApproval struct {
	Owner   common.Address
	Spender common.Address
	Value   *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterApproval is a free log retrieval operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) FilterApproval(opts *bind.FilterOpts, owner []common.Address, spender []common.Address) (*LegacyERC20MNTApprovalIterator, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.FilterLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNTApprovalIterator{contract: _LegacyERC20MNT.contract, event: "Approval", logs: logs, sub: sub}, nil
}

// WatchApproval is a free log subscription operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) WatchApproval(opts *bind.WatchOpts, sink chan<- *LegacyERC20MNTApproval, owner []common.Address, spender []common.Address) (event.Subscription, error) {

	var ownerRule []interface{}
	for _, ownerItem := range owner {
		ownerRule = append(ownerRule, ownerItem)
	}
	var spenderRule []interface{}
	for _, spenderItem := range spender {
		spenderRule = append(spenderRule, spenderItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.WatchLogs(opts, "Approval", ownerRule, spenderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20MNTApproval)
				if err := _LegacyERC20MNT.contract.UnpackLog(event, "Approval", log); err != nil {
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

// ParseApproval is a log parse operation binding the contract event 0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925.
//
// Solidity: event Approval(address indexed owner, address indexed spender, uint256 value)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) ParseApproval(log types.Log) (*LegacyERC20MNTApproval, error) {
	event := new(LegacyERC20MNTApproval)
	if err := _LegacyERC20MNT.contract.UnpackLog(event, "Approval", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LegacyERC20MNTBurnIterator is returned from FilterBurn and is used to iterate over the raw logs and unpacked data for Burn events raised by the LegacyERC20MNT contract.
type LegacyERC20MNTBurnIterator struct {
	Event *LegacyERC20MNTBurn // Event containing the contract specifics and raw log

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
func (it *LegacyERC20MNTBurnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20MNTBurn)
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
		it.Event = new(LegacyERC20MNTBurn)
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
func (it *LegacyERC20MNTBurnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20MNTBurnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20MNTBurn represents a Burn event raised by the LegacyERC20MNT contract.
type LegacyERC20MNTBurn struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterBurn is a free log retrieval operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) FilterBurn(opts *bind.FilterOpts, account []common.Address) (*LegacyERC20MNTBurnIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.FilterLogs(opts, "Burn", accountRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNTBurnIterator{contract: _LegacyERC20MNT.contract, event: "Burn", logs: logs, sub: sub}, nil
}

// WatchBurn is a free log subscription operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) WatchBurn(opts *bind.WatchOpts, sink chan<- *LegacyERC20MNTBurn, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.WatchLogs(opts, "Burn", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20MNTBurn)
				if err := _LegacyERC20MNT.contract.UnpackLog(event, "Burn", log); err != nil {
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

// ParseBurn is a log parse operation binding the contract event 0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5.
//
// Solidity: event Burn(address indexed account, uint256 amount)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) ParseBurn(log types.Log) (*LegacyERC20MNTBurn, error) {
	event := new(LegacyERC20MNTBurn)
	if err := _LegacyERC20MNT.contract.UnpackLog(event, "Burn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LegacyERC20MNTMintIterator is returned from FilterMint and is used to iterate over the raw logs and unpacked data for Mint events raised by the LegacyERC20MNT contract.
type LegacyERC20MNTMintIterator struct {
	Event *LegacyERC20MNTMint // Event containing the contract specifics and raw log

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
func (it *LegacyERC20MNTMintIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20MNTMint)
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
		it.Event = new(LegacyERC20MNTMint)
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
func (it *LegacyERC20MNTMintIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20MNTMintIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20MNTMint represents a Mint event raised by the LegacyERC20MNT contract.
type LegacyERC20MNTMint struct {
	Account common.Address
	Amount  *big.Int
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterMint is a free log retrieval operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) FilterMint(opts *bind.FilterOpts, account []common.Address) (*LegacyERC20MNTMintIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.FilterLogs(opts, "Mint", accountRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNTMintIterator{contract: _LegacyERC20MNT.contract, event: "Mint", logs: logs, sub: sub}, nil
}

// WatchMint is a free log subscription operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) WatchMint(opts *bind.WatchOpts, sink chan<- *LegacyERC20MNTMint, account []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.WatchLogs(opts, "Mint", accountRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20MNTMint)
				if err := _LegacyERC20MNT.contract.UnpackLog(event, "Mint", log); err != nil {
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

// ParseMint is a log parse operation binding the contract event 0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885.
//
// Solidity: event Mint(address indexed account, uint256 amount)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) ParseMint(log types.Log) (*LegacyERC20MNTMint, error) {
	event := new(LegacyERC20MNTMint)
	if err := _LegacyERC20MNT.contract.UnpackLog(event, "Mint", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// LegacyERC20MNTTransferIterator is returned from FilterTransfer and is used to iterate over the raw logs and unpacked data for Transfer events raised by the LegacyERC20MNT contract.
type LegacyERC20MNTTransferIterator struct {
	Event *LegacyERC20MNTTransfer // Event containing the contract specifics and raw log

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
func (it *LegacyERC20MNTTransferIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(LegacyERC20MNTTransfer)
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
		it.Event = new(LegacyERC20MNTTransfer)
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
func (it *LegacyERC20MNTTransferIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *LegacyERC20MNTTransferIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// LegacyERC20MNTTransfer represents a Transfer event raised by the LegacyERC20MNT contract.
type LegacyERC20MNTTransfer struct {
	From  common.Address
	To    common.Address
	Value *big.Int
	Raw   types.Log // Blockchain specific contextual infos
}

// FilterTransfer is a free log retrieval operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) FilterTransfer(opts *bind.FilterOpts, from []common.Address, to []common.Address) (*LegacyERC20MNTTransferIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.FilterLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return &LegacyERC20MNTTransferIterator{contract: _LegacyERC20MNT.contract, event: "Transfer", logs: logs, sub: sub}, nil
}

// WatchTransfer is a free log subscription operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) WatchTransfer(opts *bind.WatchOpts, sink chan<- *LegacyERC20MNTTransfer, from []common.Address, to []common.Address) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _LegacyERC20MNT.contract.WatchLogs(opts, "Transfer", fromRule, toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(LegacyERC20MNTTransfer)
				if err := _LegacyERC20MNT.contract.UnpackLog(event, "Transfer", log); err != nil {
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

// ParseTransfer is a log parse operation binding the contract event 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef.
//
// Solidity: event Transfer(address indexed from, address indexed to, uint256 value)
func (_LegacyERC20MNT *LegacyERC20MNTFilterer) ParseTransfer(log types.Log) (*LegacyERC20MNTTransfer, error) {
	event := new(LegacyERC20MNTTransfer)
	if err := _LegacyERC20MNT.contract.UnpackLog(event, "Transfer", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
