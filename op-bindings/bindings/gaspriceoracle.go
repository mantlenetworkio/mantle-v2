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

// GasPriceOracleMetaData contains all meta data concerning the GasPriceOracle contract.
var GasPriceOracleMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DECIMALS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"_gap\",\"inputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"baseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decimals\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"gasPrice\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getL1Fee\",\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getL1GasUsed\",\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorFee\",\"inputs\":[{\"name\":\"_gasUsed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isLimb\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"l1BaseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operator\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorFeeConstant\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorFeeScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"overhead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"scalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setLimb\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperator\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperatorFeeConstant\",\"inputs\":[{\"name\":\"_operatorFeeConstant\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperatorFeeScalar\",\"inputs\":[{\"name\":\"_operatorFeeScalar\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTokenRatio\",\"inputs\":[{\"name\":\"_tokenRatio\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"tokenRatio\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"OperatorFeeConstantUpdated\",\"inputs\":[{\"name\":\"previousOperatorFeeConstant\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"newOperatorFeeConstant\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorFeeScalarUpdated\",\"inputs\":[{\"name\":\"previousOperatorFeeScalar\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"newOperatorFeeScalar\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OperatorUpdated\",\"inputs\":[{\"name\":\"previousOperator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOperator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TokenRatioUpdated\",\"inputs\":[{\"name\":\"previousTokenRatio\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"newTokenRatio\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false}]",
	Bin: "0x60e060405234801561001057600080fd5b506001608081905260a052600060c05260805160a05160c05161131661004f6000396000610569015260006105400152600061051701526113166000f3fe608060405234801561001057600080fd5b50600436106101985760003560e01c80635d71ff8f116100e3578063de26c4a11161008c578063f2fde38b11610066578063f2fde38b14610329578063f45e65d81461033c578063fe173b971461029257600080fd5b8063de26c4a1146102e6578063e38e91f9146102f9578063e875f02a1461030c57600080fd5b8063b3ab15fb116100bd578063b3ab15fb146102b8578063d76a0548146102cb578063dab3b3da146102d357600080fd5b80635d71ff8f1461027f5780636ef25c3a146102925780638da5cb5b1461029857600080fd5b806332e70fea11610145578063519b4bd31161011f578063519b4bd31461021d57806354fd4d5014610225578063570ca7351461023a57600080fd5b806332e70fea146101ec57806349948e0e146102015780634d5d9a2a1461021457600080fd5b8063275aedd211610176578063275aedd2146101ca5780632e0f2625146101dd578063313ce567146101e557600080fd5b806306f837d31461019d5780630c18c162146101b957806316d3bc7f146101c1575b600080fd5b6101a660005481565b6040519081526020015b60405180910390f35b6101a6610344565b6101a660035481565b6101a66101d8366004610e20565b610353565b6101a6600681565b60066101a6565b6101ff6101fa366004610e20565b6103c1565b005b6101a661020f366004610e68565b610480565b6101a660045481565b6101a661048b565b61022d610510565b6040516101b09190610f67565b60025461025a9073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101b0565b6101ff61028d366004610e20565b6105b3565b486101a6565b60015461025a9073ffffffffffffffffffffffffffffffffffffffff1681565b6101ff6102c6366004610fb8565b61066d565b6101ff610765565b6101a66102e1366004610e20565b6108c8565b6101a66102f4366004610e68565b6108df565b6101ff610307366004610e20565b6108fc565b600f546103199060ff1681565b60405190151581526020016101b0565b6101ff610337366004610fb8565b6109b4565b6101a6610b29565b600061034e610b33565b905090565b600f5460009060ff1661036857506000919050565b6004546103bb906103ac90620f4240907fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8615878302888104909314170117611053565b60035481019081106000031790565b92915050565b60025473ffffffffffffffffffffffffffffffffffffffff163314610447576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f43616c6c6572206973206e6f7420746865206f70657261746f7200000000000060448201526064015b60405180910390fd5b6003805490829055604051829082907f08a9bc8992a7c4fa053bafee70f234ebf754c491d16759a28adf47e3cd9375b990600090a35050565b60006103bb82610b94565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16635cf249696040518163ffffffff1660e01b8152600401602060405180830381865afa1580156104ec573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061034e9190611067565b606061053b7f0000000000000000000000000000000000000000000000000000000000000000610bfa565b6105647f0000000000000000000000000000000000000000000000000000000000000000610bfa565b61058d7f0000000000000000000000000000000000000000000000000000000000000000610bfa565b60405160200161059f93929190611080565b604051602081830303815290604052905090565b60025473ffffffffffffffffffffffffffffffffffffffff163314610634576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f43616c6c6572206973206e6f7420746865206f70657261746f72000000000000604482015260640161043e565b6004805490829055604051829082907f977ba0b597123a7c26f0d57b10b1ab88e14d4e8676e6629640df681ccf5ffcf290600090a35050565b60015473ffffffffffffffffffffffffffffffffffffffff1633146106ee576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f43616c6c6572206973206e6f7420746865206f776e6572000000000000000000604482015260640161043e565b6002805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907ffbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad0390600090a35050565b3373deaddeaddeaddeaddeaddeaddeaddeaddead000114610808576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603e60248201527f47617350726963654f7261636c653a206f6e6c7920746865206465706f73697460448201527f6f72206163636f756e742063616e207365742069734c696d6220666c61670000606482015260840161043e565b600f5460ff161561089b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602260248201527f47617350726963654f7261636c653a2049734c696d6220616c7265616479207360448201527f6574000000000000000000000000000000000000000000000000000000000000606482015260840161043e565b600f80547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00166001179055565b600581600a81106108d857600080fd5b0154905081565b60006108e9610b33565b6108f283610d2f565b6103bb91906110f6565b60025473ffffffffffffffffffffffffffffffffffffffff16331461097d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f43616c6c6572206973206e6f7420746865206f70657261746f72000000000000604482015260640161043e565b600080548282556040519091839183917f5d6ae9db2d6725497bed0302a8212c0db5fdb3bd7d14f188a83b5589089caafd91a35050565b60015473ffffffffffffffffffffffffffffffffffffffff163314610a35576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f43616c6c6572206973206e6f7420746865206f776e6572000000000000000000604482015260640161043e565b73ffffffffffffffffffffffffffffffffffffffff8116610ab2576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f6e6577206f776e657220697320746865207a65726f2061646472657373000000604482015260640161043e565b6001805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b600061034e610dbf565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16638b239f736040518163ffffffff1660e01b8152600401602060405180830381865afa1580156104ec573d6000803e3d6000fd5b600080610ba083610d2f565b90506000610bac610dbf565b610bb461048b565b610bbc610b33565b610bc690856110f6565b610bd0919061110e565b610bda919061110e565b9050610be86006600a61126b565b610bf29082611053565b949350505050565b606081600003610c3d57505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b8115610c675780610c5181611277565b9150610c609050600a83611053565b9150610c41565b60008167ffffffffffffffff811115610c8257610c82610e39565b6040519080825280601f01601f191660200182016040528015610cac576020820181803683370190505b5090505b8415610bf257610cc16001836112af565b9150610cce600a866112c6565b610cd99060306110f6565b60f81b818381518110610cee57610cee6112da565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350610d28600a86611053565b9450610cb0565b80516000908190815b81811015610db257848181518110610d5257610d526112da565b01602001517fff0000000000000000000000000000000000000000000000000000000000000016600003610d9257610d8b6004846110f6565b9250610da0565b610d9d6010846110f6565b92505b80610daa81611277565b915050610d38565b50610bf2826104406110f6565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16639e8c49666040518163ffffffff1660e01b8152600401602060405180830381865afa1580156104ec573d6000803e3d6000fd5b600060208284031215610e3257600080fd5b5035919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600060208284031215610e7a57600080fd5b813567ffffffffffffffff80821115610e9257600080fd5b818401915084601f830112610ea657600080fd5b813581811115610eb857610eb8610e39565b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0908116603f01168101908382118183101715610efe57610efe610e39565b81604052828152876020848701011115610f1757600080fd5b826020860160208301376000928101602001929092525095945050505050565b60005b83811015610f52578181015183820152602001610f3a565b83811115610f61576000848401525b50505050565b6020815260008251806020840152610f86816040850160208701610f37565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169190910160400192915050565b600060208284031215610fca57600080fd5b813573ffffffffffffffffffffffffffffffffffffffff81168114610fee57600080fd5b9392505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b60008261106257611062610ff5565b500490565b60006020828403121561107957600080fd5b5051919050565b60008451611092818460208901610f37565b80830190507f2e0000000000000000000000000000000000000000000000000000000000000080825285516110ce816001850160208a01610f37565b600192019182015283516110e9816002840160208801610f37565b0160020195945050505050565b6000821982111561110957611109611024565b500190565b6000817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff048311821515161561114657611146611024565b500290565b600181815b808511156111a457817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0482111561118a5761118a611024565b8085161561119757918102915b93841c9390800290611150565b509250929050565b6000826111bb575060016103bb565b816111c8575060006103bb565b81600181146111de57600281146111e857611204565b60019150506103bb565b60ff8411156111f9576111f9611024565b50506001821b6103bb565b5060208310610133831016604e8410600b8410161715611227575081810a6103bb565b611231838361114b565b807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0482111561126357611263611024565b029392505050565b6000610fee83836111ac565b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036112a8576112a8611024565b5060010190565b6000828210156112c1576112c1611024565b500390565b6000826112d5576112d5610ff5565b500690565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fdfea164736f6c634300080f000a",
}

// GasPriceOracleABI is the input ABI used to generate the binding from.
// Deprecated: Use GasPriceOracleMetaData.ABI instead.
var GasPriceOracleABI = GasPriceOracleMetaData.ABI

// GasPriceOracleBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use GasPriceOracleMetaData.Bin instead.
var GasPriceOracleBin = GasPriceOracleMetaData.Bin

// DeployGasPriceOracle deploys a new Ethereum contract, binding an instance of GasPriceOracle to it.
func DeployGasPriceOracle(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *GasPriceOracle, error) {
	parsed, err := GasPriceOracleMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(GasPriceOracleBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &GasPriceOracle{GasPriceOracleCaller: GasPriceOracleCaller{contract: contract}, GasPriceOracleTransactor: GasPriceOracleTransactor{contract: contract}, GasPriceOracleFilterer: GasPriceOracleFilterer{contract: contract}}, nil
}

// GasPriceOracle is an auto generated Go binding around an Ethereum contract.
type GasPriceOracle struct {
	GasPriceOracleCaller     // Read-only binding to the contract
	GasPriceOracleTransactor // Write-only binding to the contract
	GasPriceOracleFilterer   // Log filterer for contract events
}

// GasPriceOracleCaller is an auto generated read-only Go binding around an Ethereum contract.
type GasPriceOracleCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GasPriceOracleTransactor is an auto generated write-only Go binding around an Ethereum contract.
type GasPriceOracleTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GasPriceOracleFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type GasPriceOracleFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// GasPriceOracleSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type GasPriceOracleSession struct {
	Contract     *GasPriceOracle   // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// GasPriceOracleCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type GasPriceOracleCallerSession struct {
	Contract *GasPriceOracleCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts         // Call options to use throughout this session
}

// GasPriceOracleTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type GasPriceOracleTransactorSession struct {
	Contract     *GasPriceOracleTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts         // Transaction auth options to use throughout this session
}

// GasPriceOracleRaw is an auto generated low-level Go binding around an Ethereum contract.
type GasPriceOracleRaw struct {
	Contract *GasPriceOracle // Generic contract binding to access the raw methods on
}

// GasPriceOracleCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type GasPriceOracleCallerRaw struct {
	Contract *GasPriceOracleCaller // Generic read-only contract binding to access the raw methods on
}

// GasPriceOracleTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type GasPriceOracleTransactorRaw struct {
	Contract *GasPriceOracleTransactor // Generic write-only contract binding to access the raw methods on
}

// NewGasPriceOracle creates a new instance of GasPriceOracle, bound to a specific deployed contract.
func NewGasPriceOracle(address common.Address, backend bind.ContractBackend) (*GasPriceOracle, error) {
	contract, err := bindGasPriceOracle(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracle{GasPriceOracleCaller: GasPriceOracleCaller{contract: contract}, GasPriceOracleTransactor: GasPriceOracleTransactor{contract: contract}, GasPriceOracleFilterer: GasPriceOracleFilterer{contract: contract}}, nil
}

// NewGasPriceOracleCaller creates a new read-only instance of GasPriceOracle, bound to a specific deployed contract.
func NewGasPriceOracleCaller(address common.Address, caller bind.ContractCaller) (*GasPriceOracleCaller, error) {
	contract, err := bindGasPriceOracle(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleCaller{contract: contract}, nil
}

// NewGasPriceOracleTransactor creates a new write-only instance of GasPriceOracle, bound to a specific deployed contract.
func NewGasPriceOracleTransactor(address common.Address, transactor bind.ContractTransactor) (*GasPriceOracleTransactor, error) {
	contract, err := bindGasPriceOracle(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleTransactor{contract: contract}, nil
}

// NewGasPriceOracleFilterer creates a new log filterer instance of GasPriceOracle, bound to a specific deployed contract.
func NewGasPriceOracleFilterer(address common.Address, filterer bind.ContractFilterer) (*GasPriceOracleFilterer, error) {
	contract, err := bindGasPriceOracle(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleFilterer{contract: contract}, nil
}

// bindGasPriceOracle binds a generic wrapper to an already deployed contract.
func bindGasPriceOracle(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := GasPriceOracleMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GasPriceOracle *GasPriceOracleRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GasPriceOracle.Contract.GasPriceOracleCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GasPriceOracle *GasPriceOracleRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.GasPriceOracleTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GasPriceOracle *GasPriceOracleRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.GasPriceOracleTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_GasPriceOracle *GasPriceOracleCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _GasPriceOracle.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_GasPriceOracle *GasPriceOracleTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_GasPriceOracle *GasPriceOracleTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.contract.Transact(opts, method, params...)
}

// DECIMALS is a free data retrieval call binding the contract method 0x2e0f2625.
//
// Solidity: function DECIMALS() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) DECIMALS(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "DECIMALS")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// DECIMALS is a free data retrieval call binding the contract method 0x2e0f2625.
//
// Solidity: function DECIMALS() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) DECIMALS() (*big.Int, error) {
	return _GasPriceOracle.Contract.DECIMALS(&_GasPriceOracle.CallOpts)
}

// DECIMALS is a free data retrieval call binding the contract method 0x2e0f2625.
//
// Solidity: function DECIMALS() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) DECIMALS() (*big.Int, error) {
	return _GasPriceOracle.Contract.DECIMALS(&_GasPriceOracle.CallOpts)
}

// Gap is a free data retrieval call binding the contract method 0xdab3b3da.
//
// Solidity: function _gap(uint256 ) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) Gap(opts *bind.CallOpts, arg0 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "_gap", arg0)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Gap is a free data retrieval call binding the contract method 0xdab3b3da.
//
// Solidity: function _gap(uint256 ) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) Gap(arg0 *big.Int) (*big.Int, error) {
	return _GasPriceOracle.Contract.Gap(&_GasPriceOracle.CallOpts, arg0)
}

// Gap is a free data retrieval call binding the contract method 0xdab3b3da.
//
// Solidity: function _gap(uint256 ) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) Gap(arg0 *big.Int) (*big.Int, error) {
	return _GasPriceOracle.Contract.Gap(&_GasPriceOracle.CallOpts, arg0)
}

// BaseFee is a free data retrieval call binding the contract method 0x6ef25c3a.
//
// Solidity: function baseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) BaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "baseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BaseFee is a free data retrieval call binding the contract method 0x6ef25c3a.
//
// Solidity: function baseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) BaseFee() (*big.Int, error) {
	return _GasPriceOracle.Contract.BaseFee(&_GasPriceOracle.CallOpts)
}

// BaseFee is a free data retrieval call binding the contract method 0x6ef25c3a.
//
// Solidity: function baseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) BaseFee() (*big.Int, error) {
	return _GasPriceOracle.Contract.BaseFee(&_GasPriceOracle.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() pure returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) Decimals(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "decimals")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() pure returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) Decimals() (*big.Int, error) {
	return _GasPriceOracle.Contract.Decimals(&_GasPriceOracle.CallOpts)
}

// Decimals is a free data retrieval call binding the contract method 0x313ce567.
//
// Solidity: function decimals() pure returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) Decimals() (*big.Int, error) {
	return _GasPriceOracle.Contract.Decimals(&_GasPriceOracle.CallOpts)
}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) GasPrice(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "gasPrice")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) GasPrice() (*big.Int, error) {
	return _GasPriceOracle.Contract.GasPrice(&_GasPriceOracle.CallOpts)
}

// GasPrice is a free data retrieval call binding the contract method 0xfe173b97.
//
// Solidity: function gasPrice() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) GasPrice() (*big.Int, error) {
	return _GasPriceOracle.Contract.GasPrice(&_GasPriceOracle.CallOpts)
}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) GetL1Fee(opts *bind.CallOpts, _data []byte) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "getL1Fee", _data)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) GetL1Fee(_data []byte) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetL1Fee(&_GasPriceOracle.CallOpts, _data)
}

// GetL1Fee is a free data retrieval call binding the contract method 0x49948e0e.
//
// Solidity: function getL1Fee(bytes _data) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) GetL1Fee(_data []byte) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetL1Fee(&_GasPriceOracle.CallOpts, _data)
}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) GetL1GasUsed(opts *bind.CallOpts, _data []byte) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "getL1GasUsed", _data)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) GetL1GasUsed(_data []byte) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetL1GasUsed(&_GasPriceOracle.CallOpts, _data)
}

// GetL1GasUsed is a free data retrieval call binding the contract method 0xde26c4a1.
//
// Solidity: function getL1GasUsed(bytes _data) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) GetL1GasUsed(_data []byte) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetL1GasUsed(&_GasPriceOracle.CallOpts, _data)
}

// GetOperatorFee is a free data retrieval call binding the contract method 0x275aedd2.
//
// Solidity: function getOperatorFee(uint256 _gasUsed) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) GetOperatorFee(opts *bind.CallOpts, _gasUsed *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "getOperatorFee", _gasUsed)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetOperatorFee is a free data retrieval call binding the contract method 0x275aedd2.
//
// Solidity: function getOperatorFee(uint256 _gasUsed) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) GetOperatorFee(_gasUsed *big.Int) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetOperatorFee(&_GasPriceOracle.CallOpts, _gasUsed)
}

// GetOperatorFee is a free data retrieval call binding the contract method 0x275aedd2.
//
// Solidity: function getOperatorFee(uint256 _gasUsed) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) GetOperatorFee(_gasUsed *big.Int) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetOperatorFee(&_GasPriceOracle.CallOpts, _gasUsed)
}

// IsLimb is a free data retrieval call binding the contract method 0xe875f02a.
//
// Solidity: function isLimb() view returns(bool)
func (_GasPriceOracle *GasPriceOracleCaller) IsLimb(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "isLimb")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsLimb is a free data retrieval call binding the contract method 0xe875f02a.
//
// Solidity: function isLimb() view returns(bool)
func (_GasPriceOracle *GasPriceOracleSession) IsLimb() (bool, error) {
	return _GasPriceOracle.Contract.IsLimb(&_GasPriceOracle.CallOpts)
}

// IsLimb is a free data retrieval call binding the contract method 0xe875f02a.
//
// Solidity: function isLimb() view returns(bool)
func (_GasPriceOracle *GasPriceOracleCallerSession) IsLimb() (bool, error) {
	return _GasPriceOracle.Contract.IsLimb(&_GasPriceOracle.CallOpts)
}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) L1BaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "l1BaseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) L1BaseFee() (*big.Int, error) {
	return _GasPriceOracle.Contract.L1BaseFee(&_GasPriceOracle.CallOpts)
}

// L1BaseFee is a free data retrieval call binding the contract method 0x519b4bd3.
//
// Solidity: function l1BaseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) L1BaseFee() (*big.Int, error) {
	return _GasPriceOracle.Contract.L1BaseFee(&_GasPriceOracle.CallOpts)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_GasPriceOracle *GasPriceOracleCaller) Operator(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "operator")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_GasPriceOracle *GasPriceOracleSession) Operator() (common.Address, error) {
	return _GasPriceOracle.Contract.Operator(&_GasPriceOracle.CallOpts)
}

// Operator is a free data retrieval call binding the contract method 0x570ca735.
//
// Solidity: function operator() view returns(address)
func (_GasPriceOracle *GasPriceOracleCallerSession) Operator() (common.Address, error) {
	return _GasPriceOracle.Contract.Operator(&_GasPriceOracle.CallOpts)
}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) OperatorFeeConstant(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "operatorFeeConstant")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) OperatorFeeConstant() (*big.Int, error) {
	return _GasPriceOracle.Contract.OperatorFeeConstant(&_GasPriceOracle.CallOpts)
}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) OperatorFeeConstant() (*big.Int, error) {
	return _GasPriceOracle.Contract.OperatorFeeConstant(&_GasPriceOracle.CallOpts)
}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) OperatorFeeScalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "operatorFeeScalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) OperatorFeeScalar() (*big.Int, error) {
	return _GasPriceOracle.Contract.OperatorFeeScalar(&_GasPriceOracle.CallOpts)
}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) OperatorFeeScalar() (*big.Int, error) {
	return _GasPriceOracle.Contract.OperatorFeeScalar(&_GasPriceOracle.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) Overhead(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "overhead")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) Overhead() (*big.Int, error) {
	return _GasPriceOracle.Contract.Overhead(&_GasPriceOracle.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) Overhead() (*big.Int, error) {
	return _GasPriceOracle.Contract.Overhead(&_GasPriceOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GasPriceOracle *GasPriceOracleCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GasPriceOracle *GasPriceOracleSession) Owner() (common.Address, error) {
	return _GasPriceOracle.Contract.Owner(&_GasPriceOracle.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_GasPriceOracle *GasPriceOracleCallerSession) Owner() (common.Address, error) {
	return _GasPriceOracle.Contract.Owner(&_GasPriceOracle.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) Scalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "scalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) Scalar() (*big.Int, error) {
	return _GasPriceOracle.Contract.Scalar(&_GasPriceOracle.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) Scalar() (*big.Int, error) {
	return _GasPriceOracle.Contract.Scalar(&_GasPriceOracle.CallOpts)
}

// TokenRatio is a free data retrieval call binding the contract method 0x06f837d3.
//
// Solidity: function tokenRatio() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) TokenRatio(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "tokenRatio")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TokenRatio is a free data retrieval call binding the contract method 0x06f837d3.
//
// Solidity: function tokenRatio() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) TokenRatio() (*big.Int, error) {
	return _GasPriceOracle.Contract.TokenRatio(&_GasPriceOracle.CallOpts)
}

// TokenRatio is a free data retrieval call binding the contract method 0x06f837d3.
//
// Solidity: function tokenRatio() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) TokenRatio() (*big.Int, error) {
	return _GasPriceOracle.Contract.TokenRatio(&_GasPriceOracle.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_GasPriceOracle *GasPriceOracleCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_GasPriceOracle *GasPriceOracleSession) Version() (string, error) {
	return _GasPriceOracle.Contract.Version(&_GasPriceOracle.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_GasPriceOracle *GasPriceOracleCallerSession) Version() (string, error) {
	return _GasPriceOracle.Contract.Version(&_GasPriceOracle.CallOpts)
}

// SetLimb is a paid mutator transaction binding the contract method 0xd76a0548.
//
// Solidity: function setLimb() returns()
func (_GasPriceOracle *GasPriceOracleTransactor) SetLimb(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracle.contract.Transact(opts, "setLimb")
}

// SetLimb is a paid mutator transaction binding the contract method 0xd76a0548.
//
// Solidity: function setLimb() returns()
func (_GasPriceOracle *GasPriceOracleSession) SetLimb() (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetLimb(&_GasPriceOracle.TransactOpts)
}

// SetLimb is a paid mutator transaction binding the contract method 0xd76a0548.
//
// Solidity: function setLimb() returns()
func (_GasPriceOracle *GasPriceOracleTransactorSession) SetLimb() (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetLimb(&_GasPriceOracle.TransactOpts)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address _operator) returns()
func (_GasPriceOracle *GasPriceOracleTransactor) SetOperator(opts *bind.TransactOpts, _operator common.Address) (*types.Transaction, error) {
	return _GasPriceOracle.contract.Transact(opts, "setOperator", _operator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address _operator) returns()
func (_GasPriceOracle *GasPriceOracleSession) SetOperator(_operator common.Address) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetOperator(&_GasPriceOracle.TransactOpts, _operator)
}

// SetOperator is a paid mutator transaction binding the contract method 0xb3ab15fb.
//
// Solidity: function setOperator(address _operator) returns()
func (_GasPriceOracle *GasPriceOracleTransactorSession) SetOperator(_operator common.Address) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetOperator(&_GasPriceOracle.TransactOpts, _operator)
}

// SetOperatorFeeConstant is a paid mutator transaction binding the contract method 0x32e70fea.
//
// Solidity: function setOperatorFeeConstant(uint256 _operatorFeeConstant) returns()
func (_GasPriceOracle *GasPriceOracleTransactor) SetOperatorFeeConstant(opts *bind.TransactOpts, _operatorFeeConstant *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.contract.Transact(opts, "setOperatorFeeConstant", _operatorFeeConstant)
}

// SetOperatorFeeConstant is a paid mutator transaction binding the contract method 0x32e70fea.
//
// Solidity: function setOperatorFeeConstant(uint256 _operatorFeeConstant) returns()
func (_GasPriceOracle *GasPriceOracleSession) SetOperatorFeeConstant(_operatorFeeConstant *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetOperatorFeeConstant(&_GasPriceOracle.TransactOpts, _operatorFeeConstant)
}

// SetOperatorFeeConstant is a paid mutator transaction binding the contract method 0x32e70fea.
//
// Solidity: function setOperatorFeeConstant(uint256 _operatorFeeConstant) returns()
func (_GasPriceOracle *GasPriceOracleTransactorSession) SetOperatorFeeConstant(_operatorFeeConstant *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetOperatorFeeConstant(&_GasPriceOracle.TransactOpts, _operatorFeeConstant)
}

// SetOperatorFeeScalar is a paid mutator transaction binding the contract method 0x5d71ff8f.
//
// Solidity: function setOperatorFeeScalar(uint256 _operatorFeeScalar) returns()
func (_GasPriceOracle *GasPriceOracleTransactor) SetOperatorFeeScalar(opts *bind.TransactOpts, _operatorFeeScalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.contract.Transact(opts, "setOperatorFeeScalar", _operatorFeeScalar)
}

// SetOperatorFeeScalar is a paid mutator transaction binding the contract method 0x5d71ff8f.
//
// Solidity: function setOperatorFeeScalar(uint256 _operatorFeeScalar) returns()
func (_GasPriceOracle *GasPriceOracleSession) SetOperatorFeeScalar(_operatorFeeScalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetOperatorFeeScalar(&_GasPriceOracle.TransactOpts, _operatorFeeScalar)
}

// SetOperatorFeeScalar is a paid mutator transaction binding the contract method 0x5d71ff8f.
//
// Solidity: function setOperatorFeeScalar(uint256 _operatorFeeScalar) returns()
func (_GasPriceOracle *GasPriceOracleTransactorSession) SetOperatorFeeScalar(_operatorFeeScalar *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetOperatorFeeScalar(&_GasPriceOracle.TransactOpts, _operatorFeeScalar)
}

// SetTokenRatio is a paid mutator transaction binding the contract method 0xe38e91f9.
//
// Solidity: function setTokenRatio(uint256 _tokenRatio) returns()
func (_GasPriceOracle *GasPriceOracleTransactor) SetTokenRatio(opts *bind.TransactOpts, _tokenRatio *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.contract.Transact(opts, "setTokenRatio", _tokenRatio)
}

// SetTokenRatio is a paid mutator transaction binding the contract method 0xe38e91f9.
//
// Solidity: function setTokenRatio(uint256 _tokenRatio) returns()
func (_GasPriceOracle *GasPriceOracleSession) SetTokenRatio(_tokenRatio *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetTokenRatio(&_GasPriceOracle.TransactOpts, _tokenRatio)
}

// SetTokenRatio is a paid mutator transaction binding the contract method 0xe38e91f9.
//
// Solidity: function setTokenRatio(uint256 _tokenRatio) returns()
func (_GasPriceOracle *GasPriceOracleTransactorSession) SetTokenRatio(_tokenRatio *big.Int) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetTokenRatio(&_GasPriceOracle.TransactOpts, _tokenRatio)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _owner) returns()
func (_GasPriceOracle *GasPriceOracleTransactor) TransferOwnership(opts *bind.TransactOpts, _owner common.Address) (*types.Transaction, error) {
	return _GasPriceOracle.contract.Transact(opts, "transferOwnership", _owner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _owner) returns()
func (_GasPriceOracle *GasPriceOracleSession) TransferOwnership(_owner common.Address) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.TransferOwnership(&_GasPriceOracle.TransactOpts, _owner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address _owner) returns()
func (_GasPriceOracle *GasPriceOracleTransactorSession) TransferOwnership(_owner common.Address) (*types.Transaction, error) {
	return _GasPriceOracle.Contract.TransferOwnership(&_GasPriceOracle.TransactOpts, _owner)
}

// GasPriceOracleOperatorFeeConstantUpdatedIterator is returned from FilterOperatorFeeConstantUpdated and is used to iterate over the raw logs and unpacked data for OperatorFeeConstantUpdated events raised by the GasPriceOracle contract.
type GasPriceOracleOperatorFeeConstantUpdatedIterator struct {
	Event *GasPriceOracleOperatorFeeConstantUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleOperatorFeeConstantUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleOperatorFeeConstantUpdated)
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
		it.Event = new(GasPriceOracleOperatorFeeConstantUpdated)
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
func (it *GasPriceOracleOperatorFeeConstantUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleOperatorFeeConstantUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleOperatorFeeConstantUpdated represents a OperatorFeeConstantUpdated event raised by the GasPriceOracle contract.
type GasPriceOracleOperatorFeeConstantUpdated struct {
	PreviousOperatorFeeConstant *big.Int
	NewOperatorFeeConstant      *big.Int
	Raw                         types.Log // Blockchain specific contextual infos
}

// FilterOperatorFeeConstantUpdated is a free log retrieval operation binding the contract event 0x08a9bc8992a7c4fa053bafee70f234ebf754c491d16759a28adf47e3cd9375b9.
//
// Solidity: event OperatorFeeConstantUpdated(uint256 indexed previousOperatorFeeConstant, uint256 indexed newOperatorFeeConstant)
func (_GasPriceOracle *GasPriceOracleFilterer) FilterOperatorFeeConstantUpdated(opts *bind.FilterOpts, previousOperatorFeeConstant []*big.Int, newOperatorFeeConstant []*big.Int) (*GasPriceOracleOperatorFeeConstantUpdatedIterator, error) {

	var previousOperatorFeeConstantRule []interface{}
	for _, previousOperatorFeeConstantItem := range previousOperatorFeeConstant {
		previousOperatorFeeConstantRule = append(previousOperatorFeeConstantRule, previousOperatorFeeConstantItem)
	}
	var newOperatorFeeConstantRule []interface{}
	for _, newOperatorFeeConstantItem := range newOperatorFeeConstant {
		newOperatorFeeConstantRule = append(newOperatorFeeConstantRule, newOperatorFeeConstantItem)
	}

	logs, sub, err := _GasPriceOracle.contract.FilterLogs(opts, "OperatorFeeConstantUpdated", previousOperatorFeeConstantRule, newOperatorFeeConstantRule)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleOperatorFeeConstantUpdatedIterator{contract: _GasPriceOracle.contract, event: "OperatorFeeConstantUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorFeeConstantUpdated is a free log subscription operation binding the contract event 0x08a9bc8992a7c4fa053bafee70f234ebf754c491d16759a28adf47e3cd9375b9.
//
// Solidity: event OperatorFeeConstantUpdated(uint256 indexed previousOperatorFeeConstant, uint256 indexed newOperatorFeeConstant)
func (_GasPriceOracle *GasPriceOracleFilterer) WatchOperatorFeeConstantUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleOperatorFeeConstantUpdated, previousOperatorFeeConstant []*big.Int, newOperatorFeeConstant []*big.Int) (event.Subscription, error) {

	var previousOperatorFeeConstantRule []interface{}
	for _, previousOperatorFeeConstantItem := range previousOperatorFeeConstant {
		previousOperatorFeeConstantRule = append(previousOperatorFeeConstantRule, previousOperatorFeeConstantItem)
	}
	var newOperatorFeeConstantRule []interface{}
	for _, newOperatorFeeConstantItem := range newOperatorFeeConstant {
		newOperatorFeeConstantRule = append(newOperatorFeeConstantRule, newOperatorFeeConstantItem)
	}

	logs, sub, err := _GasPriceOracle.contract.WatchLogs(opts, "OperatorFeeConstantUpdated", previousOperatorFeeConstantRule, newOperatorFeeConstantRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleOperatorFeeConstantUpdated)
				if err := _GasPriceOracle.contract.UnpackLog(event, "OperatorFeeConstantUpdated", log); err != nil {
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

// ParseOperatorFeeConstantUpdated is a log parse operation binding the contract event 0x08a9bc8992a7c4fa053bafee70f234ebf754c491d16759a28adf47e3cd9375b9.
//
// Solidity: event OperatorFeeConstantUpdated(uint256 indexed previousOperatorFeeConstant, uint256 indexed newOperatorFeeConstant)
func (_GasPriceOracle *GasPriceOracleFilterer) ParseOperatorFeeConstantUpdated(log types.Log) (*GasPriceOracleOperatorFeeConstantUpdated, error) {
	event := new(GasPriceOracleOperatorFeeConstantUpdated)
	if err := _GasPriceOracle.contract.UnpackLog(event, "OperatorFeeConstantUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleOperatorFeeScalarUpdatedIterator is returned from FilterOperatorFeeScalarUpdated and is used to iterate over the raw logs and unpacked data for OperatorFeeScalarUpdated events raised by the GasPriceOracle contract.
type GasPriceOracleOperatorFeeScalarUpdatedIterator struct {
	Event *GasPriceOracleOperatorFeeScalarUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleOperatorFeeScalarUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleOperatorFeeScalarUpdated)
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
		it.Event = new(GasPriceOracleOperatorFeeScalarUpdated)
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
func (it *GasPriceOracleOperatorFeeScalarUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleOperatorFeeScalarUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleOperatorFeeScalarUpdated represents a OperatorFeeScalarUpdated event raised by the GasPriceOracle contract.
type GasPriceOracleOperatorFeeScalarUpdated struct {
	PreviousOperatorFeeScalar *big.Int
	NewOperatorFeeScalar      *big.Int
	Raw                       types.Log // Blockchain specific contextual infos
}

// FilterOperatorFeeScalarUpdated is a free log retrieval operation binding the contract event 0x977ba0b597123a7c26f0d57b10b1ab88e14d4e8676e6629640df681ccf5ffcf2.
//
// Solidity: event OperatorFeeScalarUpdated(uint256 indexed previousOperatorFeeScalar, uint256 indexed newOperatorFeeScalar)
func (_GasPriceOracle *GasPriceOracleFilterer) FilterOperatorFeeScalarUpdated(opts *bind.FilterOpts, previousOperatorFeeScalar []*big.Int, newOperatorFeeScalar []*big.Int) (*GasPriceOracleOperatorFeeScalarUpdatedIterator, error) {

	var previousOperatorFeeScalarRule []interface{}
	for _, previousOperatorFeeScalarItem := range previousOperatorFeeScalar {
		previousOperatorFeeScalarRule = append(previousOperatorFeeScalarRule, previousOperatorFeeScalarItem)
	}
	var newOperatorFeeScalarRule []interface{}
	for _, newOperatorFeeScalarItem := range newOperatorFeeScalar {
		newOperatorFeeScalarRule = append(newOperatorFeeScalarRule, newOperatorFeeScalarItem)
	}

	logs, sub, err := _GasPriceOracle.contract.FilterLogs(opts, "OperatorFeeScalarUpdated", previousOperatorFeeScalarRule, newOperatorFeeScalarRule)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleOperatorFeeScalarUpdatedIterator{contract: _GasPriceOracle.contract, event: "OperatorFeeScalarUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorFeeScalarUpdated is a free log subscription operation binding the contract event 0x977ba0b597123a7c26f0d57b10b1ab88e14d4e8676e6629640df681ccf5ffcf2.
//
// Solidity: event OperatorFeeScalarUpdated(uint256 indexed previousOperatorFeeScalar, uint256 indexed newOperatorFeeScalar)
func (_GasPriceOracle *GasPriceOracleFilterer) WatchOperatorFeeScalarUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleOperatorFeeScalarUpdated, previousOperatorFeeScalar []*big.Int, newOperatorFeeScalar []*big.Int) (event.Subscription, error) {

	var previousOperatorFeeScalarRule []interface{}
	for _, previousOperatorFeeScalarItem := range previousOperatorFeeScalar {
		previousOperatorFeeScalarRule = append(previousOperatorFeeScalarRule, previousOperatorFeeScalarItem)
	}
	var newOperatorFeeScalarRule []interface{}
	for _, newOperatorFeeScalarItem := range newOperatorFeeScalar {
		newOperatorFeeScalarRule = append(newOperatorFeeScalarRule, newOperatorFeeScalarItem)
	}

	logs, sub, err := _GasPriceOracle.contract.WatchLogs(opts, "OperatorFeeScalarUpdated", previousOperatorFeeScalarRule, newOperatorFeeScalarRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleOperatorFeeScalarUpdated)
				if err := _GasPriceOracle.contract.UnpackLog(event, "OperatorFeeScalarUpdated", log); err != nil {
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

// ParseOperatorFeeScalarUpdated is a log parse operation binding the contract event 0x977ba0b597123a7c26f0d57b10b1ab88e14d4e8676e6629640df681ccf5ffcf2.
//
// Solidity: event OperatorFeeScalarUpdated(uint256 indexed previousOperatorFeeScalar, uint256 indexed newOperatorFeeScalar)
func (_GasPriceOracle *GasPriceOracleFilterer) ParseOperatorFeeScalarUpdated(log types.Log) (*GasPriceOracleOperatorFeeScalarUpdated, error) {
	event := new(GasPriceOracleOperatorFeeScalarUpdated)
	if err := _GasPriceOracle.contract.UnpackLog(event, "OperatorFeeScalarUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleOperatorUpdatedIterator is returned from FilterOperatorUpdated and is used to iterate over the raw logs and unpacked data for OperatorUpdated events raised by the GasPriceOracle contract.
type GasPriceOracleOperatorUpdatedIterator struct {
	Event *GasPriceOracleOperatorUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleOperatorUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleOperatorUpdated)
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
		it.Event = new(GasPriceOracleOperatorUpdated)
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
func (it *GasPriceOracleOperatorUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleOperatorUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleOperatorUpdated represents a OperatorUpdated event raised by the GasPriceOracle contract.
type GasPriceOracleOperatorUpdated struct {
	PreviousOperator common.Address
	NewOperator      common.Address
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterOperatorUpdated is a free log retrieval operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address indexed previousOperator, address indexed newOperator)
func (_GasPriceOracle *GasPriceOracleFilterer) FilterOperatorUpdated(opts *bind.FilterOpts, previousOperator []common.Address, newOperator []common.Address) (*GasPriceOracleOperatorUpdatedIterator, error) {

	var previousOperatorRule []interface{}
	for _, previousOperatorItem := range previousOperator {
		previousOperatorRule = append(previousOperatorRule, previousOperatorItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}

	logs, sub, err := _GasPriceOracle.contract.FilterLogs(opts, "OperatorUpdated", previousOperatorRule, newOperatorRule)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleOperatorUpdatedIterator{contract: _GasPriceOracle.contract, event: "OperatorUpdated", logs: logs, sub: sub}, nil
}

// WatchOperatorUpdated is a free log subscription operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address indexed previousOperator, address indexed newOperator)
func (_GasPriceOracle *GasPriceOracleFilterer) WatchOperatorUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleOperatorUpdated, previousOperator []common.Address, newOperator []common.Address) (event.Subscription, error) {

	var previousOperatorRule []interface{}
	for _, previousOperatorItem := range previousOperator {
		previousOperatorRule = append(previousOperatorRule, previousOperatorItem)
	}
	var newOperatorRule []interface{}
	for _, newOperatorItem := range newOperator {
		newOperatorRule = append(newOperatorRule, newOperatorItem)
	}

	logs, sub, err := _GasPriceOracle.contract.WatchLogs(opts, "OperatorUpdated", previousOperatorRule, newOperatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleOperatorUpdated)
				if err := _GasPriceOracle.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
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

// ParseOperatorUpdated is a log parse operation binding the contract event 0xfbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad03.
//
// Solidity: event OperatorUpdated(address indexed previousOperator, address indexed newOperator)
func (_GasPriceOracle *GasPriceOracleFilterer) ParseOperatorUpdated(log types.Log) (*GasPriceOracleOperatorUpdated, error) {
	event := new(GasPriceOracleOperatorUpdated)
	if err := _GasPriceOracle.contract.UnpackLog(event, "OperatorUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the GasPriceOracle contract.
type GasPriceOracleOwnershipTransferredIterator struct {
	Event *GasPriceOracleOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleOwnershipTransferred)
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
		it.Event = new(GasPriceOracleOwnershipTransferred)
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
func (it *GasPriceOracleOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleOwnershipTransferred represents a OwnershipTransferred event raised by the GasPriceOracle contract.
type GasPriceOracleOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_GasPriceOracle *GasPriceOracleFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*GasPriceOracleOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _GasPriceOracle.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleOwnershipTransferredIterator{contract: _GasPriceOracle.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_GasPriceOracle *GasPriceOracleFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *GasPriceOracleOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _GasPriceOracle.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleOwnershipTransferred)
				if err := _GasPriceOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_GasPriceOracle *GasPriceOracleFilterer) ParseOwnershipTransferred(log types.Log) (*GasPriceOracleOwnershipTransferred, error) {
	event := new(GasPriceOracleOwnershipTransferred)
	if err := _GasPriceOracle.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// GasPriceOracleTokenRatioUpdatedIterator is returned from FilterTokenRatioUpdated and is used to iterate over the raw logs and unpacked data for TokenRatioUpdated events raised by the GasPriceOracle contract.
type GasPriceOracleTokenRatioUpdatedIterator struct {
	Event *GasPriceOracleTokenRatioUpdated // Event containing the contract specifics and raw log

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
func (it *GasPriceOracleTokenRatioUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(GasPriceOracleTokenRatioUpdated)
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
		it.Event = new(GasPriceOracleTokenRatioUpdated)
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
func (it *GasPriceOracleTokenRatioUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *GasPriceOracleTokenRatioUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// GasPriceOracleTokenRatioUpdated represents a TokenRatioUpdated event raised by the GasPriceOracle contract.
type GasPriceOracleTokenRatioUpdated struct {
	PreviousTokenRatio *big.Int
	NewTokenRatio      *big.Int
	Raw                types.Log // Blockchain specific contextual infos
}

// FilterTokenRatioUpdated is a free log retrieval operation binding the contract event 0x5d6ae9db2d6725497bed0302a8212c0db5fdb3bd7d14f188a83b5589089caafd.
//
// Solidity: event TokenRatioUpdated(uint256 indexed previousTokenRatio, uint256 indexed newTokenRatio)
func (_GasPriceOracle *GasPriceOracleFilterer) FilterTokenRatioUpdated(opts *bind.FilterOpts, previousTokenRatio []*big.Int, newTokenRatio []*big.Int) (*GasPriceOracleTokenRatioUpdatedIterator, error) {

	var previousTokenRatioRule []interface{}
	for _, previousTokenRatioItem := range previousTokenRatio {
		previousTokenRatioRule = append(previousTokenRatioRule, previousTokenRatioItem)
	}
	var newTokenRatioRule []interface{}
	for _, newTokenRatioItem := range newTokenRatio {
		newTokenRatioRule = append(newTokenRatioRule, newTokenRatioItem)
	}

	logs, sub, err := _GasPriceOracle.contract.FilterLogs(opts, "TokenRatioUpdated", previousTokenRatioRule, newTokenRatioRule)
	if err != nil {
		return nil, err
	}
	return &GasPriceOracleTokenRatioUpdatedIterator{contract: _GasPriceOracle.contract, event: "TokenRatioUpdated", logs: logs, sub: sub}, nil
}

// WatchTokenRatioUpdated is a free log subscription operation binding the contract event 0x5d6ae9db2d6725497bed0302a8212c0db5fdb3bd7d14f188a83b5589089caafd.
//
// Solidity: event TokenRatioUpdated(uint256 indexed previousTokenRatio, uint256 indexed newTokenRatio)
func (_GasPriceOracle *GasPriceOracleFilterer) WatchTokenRatioUpdated(opts *bind.WatchOpts, sink chan<- *GasPriceOracleTokenRatioUpdated, previousTokenRatio []*big.Int, newTokenRatio []*big.Int) (event.Subscription, error) {

	var previousTokenRatioRule []interface{}
	for _, previousTokenRatioItem := range previousTokenRatio {
		previousTokenRatioRule = append(previousTokenRatioRule, previousTokenRatioItem)
	}
	var newTokenRatioRule []interface{}
	for _, newTokenRatioItem := range newTokenRatio {
		newTokenRatioRule = append(newTokenRatioRule, newTokenRatioItem)
	}

	logs, sub, err := _GasPriceOracle.contract.WatchLogs(opts, "TokenRatioUpdated", previousTokenRatioRule, newTokenRatioRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(GasPriceOracleTokenRatioUpdated)
				if err := _GasPriceOracle.contract.UnpackLog(event, "TokenRatioUpdated", log); err != nil {
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

// ParseTokenRatioUpdated is a log parse operation binding the contract event 0x5d6ae9db2d6725497bed0302a8212c0db5fdb3bd7d14f188a83b5589089caafd.
//
// Solidity: event TokenRatioUpdated(uint256 indexed previousTokenRatio, uint256 indexed newTokenRatio)
func (_GasPriceOracle *GasPriceOracleFilterer) ParseTokenRatioUpdated(log types.Log) (*GasPriceOracleTokenRatioUpdated, error) {
	event := new(GasPriceOracleTokenRatioUpdated)
	if err := _GasPriceOracle.contract.UnpackLog(event, "TokenRatioUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
