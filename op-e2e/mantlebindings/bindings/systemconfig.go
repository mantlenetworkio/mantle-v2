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

// ResourceMeteringResourceConfig is an auto generated low-level Go binding around an user-defined struct.
type ResourceMeteringResourceConfig struct {
	MaxResourceLimit            uint32
	ElasticityMultiplier        uint8
	BaseFeeMaxChangeDenominator uint8
	MinimumBaseFee              uint32
	SystemTxMaxGas              uint32
	MaximumBaseFee              *big.Int
}

// SystemConfigMetaData contains all meta data concerning the SystemConfig contract.
var SystemConfigMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_basefeeScalar\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_blobbasefeeScalar\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_batcherHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_gasLimit\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"_baseFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_unsafeBlockSigner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_config\",\"type\":\"tuple\",\"internalType\":\"structResourceMetering.ResourceConfig\",\"components\":[{\"name\":\"maxResourceLimit\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"elasticityMultiplier\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"baseFeeMaxChangeDenominator\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"minimumBaseFee\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"systemTxMaxGas\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"maximumBaseFee\",\"type\":\"uint128\",\"internalType\":\"uint128\"}]}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"UNSAFE_BLOCK_SIGNER_SLOT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"baseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"basefeeScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"batcherHash\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"blobbasefeeScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"daFootprintGasScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"eip1559Denominator\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"eip1559Elasticity\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"gasLimit\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initialize\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_basefeeScalar\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_blobbasefeeScalar\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_batcherHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"_gasLimit\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"_baseFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_unsafeBlockSigner\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_config\",\"type\":\"tuple\",\"internalType\":\"structResourceMetering.ResourceConfig\",\"components\":[{\"name\":\"maxResourceLimit\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"elasticityMultiplier\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"baseFeeMaxChangeDenominator\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"minimumBaseFee\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"systemTxMaxGas\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"maximumBaseFee\",\"type\":\"uint128\",\"internalType\":\"uint128\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"minBaseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"minimumGasLimit\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorFeeConstant\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorFeeScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"overhead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"renounceOwnership\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"resourceConfig\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structResourceMetering.ResourceConfig\",\"components\":[{\"name\":\"maxResourceLimit\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"elasticityMultiplier\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"baseFeeMaxChangeDenominator\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"minimumBaseFee\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"systemTxMaxGas\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"maximumBaseFee\",\"type\":\"uint128\",\"internalType\":\"uint128\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"scalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setBaseFee\",\"inputs\":[{\"name\":\"_baseFee\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setBatcherHash\",\"inputs\":[{\"name\":\"_batcherHash\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setDAFootprintGasScalar\",\"inputs\":[{\"name\":\"_daFootprintGasScalar\",\"type\":\"uint16\",\"internalType\":\"uint16\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setEIP1559Params\",\"inputs\":[{\"name\":\"_denominator\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_elasticity\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGasConfig\",\"inputs\":[{\"name\":\"_overhead\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_scalar\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGasConfigArsia\",\"inputs\":[{\"name\":\"_basefeeScalar\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_blobbasefeeScalar\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setGasLimit\",\"inputs\":[{\"name\":\"_gasLimit\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setMinBaseFee\",\"inputs\":[{\"name\":\"_minBaseFee\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperatorFeeScalars\",\"inputs\":[{\"name\":\"_operatorFeeScalar\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"_operatorFeeConstant\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setResourceConfig\",\"inputs\":[{\"name\":\"_config\",\"type\":\"tuple\",\"internalType\":\"structResourceMetering.ResourceConfig\",\"components\":[{\"name\":\"maxResourceLimit\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"elasticityMultiplier\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"baseFeeMaxChangeDenominator\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"minimumBaseFee\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"systemTxMaxGas\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"maximumBaseFee\",\"type\":\"uint128\",\"internalType\":\"uint128\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setUnsafeBlockSigner\",\"inputs\":[{\"name\":\"_unsafeBlockSigner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"newOwner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"unsafeBlockSigner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"ConfigUpdate\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"updateType\",\"type\":\"uint8\",\"indexed\":true,\"internalType\":\"enumSystemConfig.UpdateType\"},{\"name\":\"data\",\"type\":\"bytes\",\"indexed\":false,\"internalType\":\"bytes\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Initialized\",\"inputs\":[{\"name\":\"version\",\"type\":\"uint8\",\"indexed\":false,\"internalType\":\"uint8\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false}]",
	Bin: "0x60e06040523480156200001157600080fd5b5060405162002abb38038062002abb83398101604081905262000034916200087e565b6001608052600460a052600060c05262000055888888888888888862000063565b505050505050505062000a99565b600054610100900460ff1615808015620000845750600054600160ff909116105b80620000b45750620000a1306200029560201b62000f5a1760201c565b158015620000b4575060005460ff166001145b6200011d5760405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b60648201526084015b60405180910390fd5b6000805460ff19166001179055801562000141576000805461ff0019166101001790555b6200014b620002a4565b62000156896200030c565b6067869055606880546001600160401b0387166001600160401b031991821617909155606a859055606b805463ffffffff8a81166401000000000291909316928b1692909217919091179055620001cb837f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c0855565b620001d6826200038b565b620001e0620006e0565b6001600160401b0316856001600160401b03161015620002435760405162461bcd60e51b815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f7700604482015260640162000114565b80156200028a576000805461ff0019169055604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b505050505050505050565b6001600160a01b03163b151590565b600054610100900460ff16620003005760405162461bcd60e51b815260206004820152602b602482015260008051602062002a9b83398151915260448201526a6e697469616c697a696e6760a81b606482015260840162000114565b6200030a6200070d565b565b6200031662000774565b6001600160a01b0381166200037d5760405162461bcd60e51b815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201526564647265737360d01b606482015260840162000114565b6200038881620007d0565b50565b8060a001516001600160801b0316816060015163ffffffff1611156200041a5760405162461bcd60e51b815260206004820152603560248201527f53797374656d436f6e6669673a206d696e206261736520666565206d7573742060448201527f6265206c657373207468616e206d617820626173650000000000000000000000606482015260840162000114565b6001816040015160ff16116200048b5760405162461bcd60e51b815260206004820152602f60248201527f53797374656d436f6e6669673a2064656e6f6d696e61746f72206d757374206260448201526e65206c6172676572207468616e203160881b606482015260840162000114565b606854608082015182516001600160401b0390921691620004ad9190620009e8565b63ffffffff161115620005035760405162461bcd60e51b815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f7700604482015260640162000114565b6000816020015160ff1611620005745760405162461bcd60e51b815260206004820152602f60248201527f53797374656d436f6e6669673a20656c6173746963697479206d756c7469706c60448201526e06965722063616e6e6f74206265203608c1b606482015260840162000114565b8051602082015163ffffffff82169160ff909116906200059690829062000a13565b620005a2919062000a45565b63ffffffff16146200061d5760405162461bcd60e51b815260206004820152603760248201527f53797374656d436f6e6669673a20707265636973696f6e206c6f73732077697460448201527f6820746172676574207265736f75726365206c696d6974000000000000000000606482015260840162000114565b805160698054602084015160408501516060860151608087015160a09097015163ffffffff96871664ffffffffff199095169490941764010000000060ff948516021764ffffffffff60281b191665010000000000939092169290920263ffffffff60301b19161766010000000000009185169190910217600160501b600160f01b0319166a01000000000000000000009390941692909202600160701b600160f01b03191692909217600160701b6001600160801b0390921691909102179055565b606954600090620007089063ffffffff6a010000000000000000000082048116911662000a74565b905090565b600054610100900460ff16620007695760405162461bcd60e51b815260206004820152602b602482015260008051602062002a9b83398151915260448201526a6e697469616c697a696e6760a81b606482015260840162000114565b6200030a33620007d0565b6033546001600160a01b031633146200030a5760405162461bcd60e51b815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e6572604482015260640162000114565b603380546001600160a01b038381166001600160a01b0319831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b80516001600160a01b03811681146200083a57600080fd5b919050565b805163ffffffff811681146200083a57600080fd5b805160ff811681146200083a57600080fd5b80516001600160801b03811681146200083a57600080fd5b600080600080600080600080888a036101a08112156200089d57600080fd5b620008a88a62000822565b9850620008b860208b016200083f565b9750620008c860408b016200083f565b60608b015160808c015191985096506001600160401b038082168214620008ee57600080fd5b81965060a08c015195506200090660c08d0162000822565b945060c060df19840112156200091b57600080fd5b604051925060c08301915082821081831117156200094957634e487b7160e01b600052604160045260246000fd5b506040526200095b60e08b016200083f565b81526200096c6101008b0162000854565b6020820152620009806101208b0162000854565b6040820152620009946101408b016200083f565b6060820152620009a86101608b016200083f565b6080820152620009bc6101808b0162000866565b60a0820152809150509295985092959890939650565b634e487b7160e01b600052601160045260246000fd5b600063ffffffff80831681851680830382111562000a0a5762000a0a620009d2565b01949350505050565b600063ffffffff8084168062000a3957634e487b7160e01b600052601260045260246000fd5b92169190910492915050565b600063ffffffff8083168185168183048111821515161562000a6b5762000a6b620009d2565b02949350505050565b60006001600160401b0382811684821680830382111562000a0a5762000a0a620009d2565b60805160a05160c051611fd262000ac96000396000610bf701526000610bce01526000610ba50152611fd26000f3fe608060405234801561001057600080fd5b50600436106102265760003560e01c8063a62611a21161012a578063cc731b02116100bd578063f2fde38b1161008c578063f68016b711610071578063f68016b714610637578063fe3d57101461064b578063ffa1ad741461067857600080fd5b8063f2fde38b1461061b578063f45e65d81461062e57600080fd5b8063cc731b02146104aa578063d220a9e0146105de578063e81b2c6d146105fa578063ec7075171461060357600080fd5b8063c71973f6116100f9578063c71973f614610451578063c9b26f6114610464578063c9ff2d1614610477578063ca76bd6c1461049757600080fd5b8063a62611a214610407578063b40a817c1461041b578063bfb14fb71461042e578063c0fd4b411461043e57600080fd5b80634add321d116101bd5780636ef25c3a1161018c5780637616f0e8116101715780637616f0e8146103c35780638da5cb5b146103d6578063935f029e146103f457600080fd5b80636ef25c3a146103b2578063715018a6146103bb57600080fd5b80634add321d146103355780634d5d9a2a1461033d5780634f16540b1461037657806354fd4d501461039d57600080fd5b806318d13918116101f957806318d13918146102b45780631fd19ee1146102c757806320f06fdc1461030f578063468606981461032257600080fd5b8063029865ef1461022b5780630c18c16214610240578063155b6c6f1461025c57806316d3bc7f1461026f575b600080fd5b61023e610239366004611b6d565b610680565b005b61024960655481565b6040519081526020015b60405180910390f35b61023e61026a366004611bf7565b610943565b606b5461029b9074010000000000000000000000000000000000000000900467ffffffffffffffff1681565b60405167ffffffffffffffff9091168152602001610253565b61023e6102c2366004611c2a565b610a6b565b7f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c08545b60405173ffffffffffffffffffffffffffffffffffffffff9091168152602001610253565b61023e61031d366004611c4c565b610b2f565b61023e610330366004611c70565b610b43565b61029b610b73565b606b5461036190700100000000000000000000000000000000900463ffffffff1681565b60405163ffffffff9091168152602001610253565b6102497f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c0881565b6103a5610b9e565b6040516102539190611d03565b610249606a5481565b61023e610c41565b61023e6103d1366004611d16565b610c55565b60335473ffffffffffffffffffffffffffffffffffffffff166102ea565b61023e610402366004611d31565b610c66565b606c5461029b9067ffffffffffffffff1681565b61023e610429366004611d16565b610cc7565b606b546103619063ffffffff1681565b61023e61044c366004611d53565b610dad565b61023e61045f366004611d7d565b610dc3565b61023e610472366004611c70565b610dd4565b606b54610361906c01000000000000000000000000900463ffffffff1681565b61023e6104a5366004611d53565b610e04565b61056e6040805160c081018252600080825260208201819052918101829052606081018290526080810182905260a0810191909152506040805160c08101825260695463ffffffff8082168352640100000000820460ff9081166020850152650100000000008304169383019390935266010000000000008104831660608301526a0100000000000000000000810490921660808201526e0100000000000000000000000000009091046fffffffffffffffffffffffffffffffff1660a082015290565b6040516102539190600060c08201905063ffffffff80845116835260ff602085015116602084015260ff6040850151166040840152806060850151166060840152806080850151166080840152506fffffffffffffffffffffffffffffffff60a08401511660a083015292915050565b606b546103619068010000000000000000900463ffffffff1681565b61024960675481565b606b5461036190640100000000900463ffffffff1681565b61023e610629366004611c2a565b610ea6565b61024960665481565b60685461029b9067ffffffffffffffff1681565b606c546106659068010000000000000000900461ffff1681565b60405161ffff9091168152602001610253565b610249600081565b600054610100900460ff16158080156106a05750600054600160ff909116105b806106ba5750303b1580156106ba575060005460ff166001145b61074b576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201527f647920696e697469616c697a656400000000000000000000000000000000000060648201526084015b60405180910390fd5b600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0016600117905580156107a957600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff166101001790555b6107b1610f76565b6107ba89610ea6565b60678690556068805467ffffffffffffffff87167fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000091821617909155606a859055606b805463ffffffff8a81166401000000000291909316928b1692909217919091179055610847837f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c0855565b61085082611015565b610858610b73565b67ffffffffffffffff168567ffffffffffffffff1610156108d5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f77006044820152606401610742565b801561093857600080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff00ff169055604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b505050505050505050565b61094b611489565b606b80547fffffffff000000000000000000000000ffffffffffffffffffffffffffffffff1670010000000000000000000000000000000063ffffffff8516027fffffffff0000000000000000ffffffffffffffffffffffffffffffffffffffff16177401000000000000000000000000000000000000000067ffffffffffffffff8416908102919091179091556040805184821b6bffffffff000000000000000016909217602083015260009101604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0818403018152919052905060065b60007f1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be83604051610a5e9190611d03565b60405180910390a3505050565b610a73611489565b610a9b817f65a7ed542fb37fe237fdfbdd70b31598523fe5b32879e307bae27a0bd9581c0855565b6040805173ffffffffffffffffffffffffffffffffffffffff8316602082015260009101604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0818403018152919052905060035b60007f1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be83604051610b239190611d03565b60405180910390a35050565b610b37611489565b610b408161150a565b50565b610b4b611489565b606a819055604080516020808201849052825180830390910181529082019091526004610af2565b606954600090610b999063ffffffff6a0100000000000000000000820481169116611dc8565b905090565b6060610bc97f000000000000000000000000000000000000000000000000000000000000000061156b565b610bf27f000000000000000000000000000000000000000000000000000000000000000061156b565b610c1b7f000000000000000000000000000000000000000000000000000000000000000061156b565b604051602001610c2d93929190611df4565b604051602081830303815290604052905090565b610c49611489565b610c5360006116a8565b565b610c5d611489565b610b408161171f565b610c6e611489565b6065829055606681905560408051602081018490529081018290526000906060015b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe081840301815291905290506001610a2d565b610ccf611489565b610cd7610b73565b67ffffffffffffffff168167ffffffffffffffff161015610d54576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f77006044820152606401610742565b606880547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001667ffffffffffffffff83169081179091556040805160208082019390935281518082039093018352810190526002610af2565b610db5611489565b610dbf8282611778565b5050565b610dcb611489565b610b4081611015565b610ddc611489565b6067819055604080516020808201849052825180830390910181529082019091526000610af2565b610e0c611489565b606b805463ffffffff838116640100000000027fffffffffffffffffffffffffffffffffffffffffffffffff000000000000000090921690851690811791909117909155602082811b67ffffffff00000000169091177f0100000000000000000000000000000000000000000000000000000000000000176066819055606554604051600093610c90939101918252602082015260400190565b610eae611489565b73ffffffffffffffffffffffffffffffffffffffff8116610f51576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f4f776e61626c653a206e6577206f776e657220697320746865207a65726f206160448201527f64647265737300000000000000000000000000000000000000000000000000006064820152608401610742565b610b40816116a8565b73ffffffffffffffffffffffffffffffffffffffff163b151590565b600054610100900460ff1661100d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e670000000000000000000000000000000000000000006064820152608401610742565b610c5361194f565b8060a001516fffffffffffffffffffffffffffffffff16816060015163ffffffff1611156110c5576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603560248201527f53797374656d436f6e6669673a206d696e206261736520666565206d7573742060448201527f6265206c657373207468616e206d6178206261736500000000000000000000006064820152608401610742565b6001816040015160ff161161115c576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f53797374656d436f6e6669673a2064656e6f6d696e61746f72206d757374206260448201527f65206c6172676572207468616e203100000000000000000000000000000000006064820152608401610742565b6068546080820151825167ffffffffffffffff9092169161117d9190611e6a565b63ffffffff1611156111eb576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601f60248201527f53797374656d436f6e6669673a20676173206c696d697420746f6f206c6f77006044820152606401610742565b6000816020015160ff1611611282576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602f60248201527f53797374656d436f6e6669673a20656c6173746963697479206d756c7469706c60448201527f6965722063616e6e6f74206265203000000000000000000000000000000000006064820152608401610742565b8051602082015163ffffffff82169160ff909116906112a2908290611eb8565b6112ac9190611edb565b63ffffffff161461133f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f53797374656d436f6e6669673a20707265636973696f6e206c6f73732077697460448201527f6820746172676574207265736f75726365206c696d69740000000000000000006064820152608401610742565b805160698054602084015160408501516060860151608087015160a09097015163ffffffff9687167fffffffffffffffffffffffffffffffffffffffffffffffffffffff00000000009095169490941764010000000060ff94851602177fffffffffffffffffffffffffffffffffffffffffffff0000000000ffffffffff166501000000000093909216929092027fffffffffffffffffffffffffffffffffffffffffffff00000000ffffffffffff1617660100000000000091851691909102177fffff0000000000000000000000000000000000000000ffffffffffffffffffff166a010000000000000000000093909416929092027fffff00000000000000000000000000000000ffffffffffffffffffffffffffff16929092176e0100000000000000000000000000006fffffffffffffffffffffffffffffffff90921691909102179055565b60335473ffffffffffffffffffffffffffffffffffffffff163314610c53576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820181905260248201527f4f776e61626c653a2063616c6c6572206973206e6f7420746865206f776e65726044820152606401610742565b606c80547fffffffffffffffffffffffffffffffffffffffffffff0000ffffffffffffffff166801000000000000000061ffff8416908102919091179091556040805160208082019390935281518082039093018352810190526008610af2565b6060816000036115ae57505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b81156115d857806115c281611f07565b91506115d19050600a83611f3f565b91506115b2565b60008167ffffffffffffffff8111156115f3576115f3611a44565b6040519080825280601f01601f19166020018201604052801561161d576020820181803683370190505b5090505b84156116a057611632600183611f53565b915061163f600a86611f6a565b61164a906030611f7e565b60f81b81838151811061165f5761165f611f96565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350611699600a86611f3f565b9450611621565b949350505050565b6033805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b606c80547fffffffffffffffffffffffffffffffffffffffffffffffff00000000000000001667ffffffffffffffff83169081179091556040805160208082019390935281518082039093018352810190526007610af2565b60018263ffffffff16101561180f576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602660248201527f53797374656d436f6e6669673a2064656e6f6d696e61746f72206d757374206260448201527f65203e3d203100000000000000000000000000000000000000000000000000006064820152608401610742565b60018163ffffffff1610156118a6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602560248201527f53797374656d436f6e6669673a20656c6173746963697479206d75737420626560448201527f203e3d20310000000000000000000000000000000000000000000000000000006064820152608401610742565b606b80547fffffffffffffffffffffffffffffffff0000000000000000ffffffffffffffff166801000000000000000063ffffffff858116919091027fffffffffffffffffffffffffffffffff00000000ffffffffffffffffffffffff16919091176c010000000000000000000000009184169182021790915560408051602085811b67ffffffff00000000169093178382015281518082039093018352810190526005610a2d565b600054610100900460ff166119e6576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e670000000000000000000000000000000000000000006064820152608401610742565b610c53336116a8565b803573ffffffffffffffffffffffffffffffffffffffff81168114611a1357600080fd5b919050565b803563ffffffff81168114611a1357600080fd5b803567ffffffffffffffff81168114611a1357600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b803560ff81168114611a1357600080fd5b80356fffffffffffffffffffffffffffffffff81168114611a1357600080fd5b600060c08284031215611ab657600080fd5b60405160c0810181811067ffffffffffffffff82111715611b00577f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b604052905080611b0f83611a18565b8152611b1d60208401611a73565b6020820152611b2e60408401611a73565b6040820152611b3f60608401611a18565b6060820152611b5060808401611a18565b6080820152611b6160a08401611a84565b60a08201525092915050565b6000806000806000806000806101a0898b031215611b8a57600080fd5b611b93896119ef565b9750611ba160208a01611a18565b9650611baf60408a01611a18565b955060608901359450611bc460808a01611a2c565b935060a08901359250611bd960c08a016119ef565b9150611be88a60e08b01611aa4565b90509295985092959890939650565b60008060408385031215611c0a57600080fd5b611c1383611a18565b9150611c2160208401611a2c565b90509250929050565b600060208284031215611c3c57600080fd5b611c45826119ef565b9392505050565b600060208284031215611c5e57600080fd5b813561ffff81168114611c4557600080fd5b600060208284031215611c8257600080fd5b5035919050565b60005b83811015611ca4578181015183820152602001611c8c565b83811115611cb3576000848401525b50505050565b60008151808452611cd1816020860160208601611c89565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b602081526000611c456020830184611cb9565b600060208284031215611d2857600080fd5b611c4582611a2c565b60008060408385031215611d4457600080fd5b50508035926020909101359150565b60008060408385031215611d6657600080fd5b611d6f83611a18565b9150611c2160208401611a18565b600060c08284031215611d8f57600080fd5b611c458383611aa4565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600067ffffffffffffffff808316818516808303821115611deb57611deb611d99565b01949350505050565b60008451611e06818460208901611c89565b80830190507f2e000000000000000000000000000000000000000000000000000000000000008082528551611e42816001850160208a01611c89565b60019201918201528351611e5d816002840160208801611c89565b0160020195945050505050565b600063ffffffff808316818516808303821115611deb57611deb611d99565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600063ffffffff80841680611ecf57611ecf611e89565b92169190910492915050565b600063ffffffff80831681851681830481118215151615611efe57611efe611d99565b02949350505050565b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203611f3857611f38611d99565b5060010190565b600082611f4e57611f4e611e89565b500490565b600082821015611f6557611f65611d99565b500390565b600082611f7957611f79611e89565b500690565b60008219821115611f9157611f91611d99565b500190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fdfea164736f6c634300080f000a496e697469616c697a61626c653a20636f6e7472616374206973206e6f742069",
}

// SystemConfigABI is the input ABI used to generate the binding from.
// Deprecated: Use SystemConfigMetaData.ABI instead.
var SystemConfigABI = SystemConfigMetaData.ABI

// SystemConfigBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use SystemConfigMetaData.Bin instead.
var SystemConfigBin = SystemConfigMetaData.Bin

// DeploySystemConfig deploys a new Ethereum contract, binding an instance of SystemConfig to it.
func DeploySystemConfig(auth *bind.TransactOpts, backend bind.ContractBackend, _owner common.Address, _basefeeScalar uint32, _blobbasefeeScalar uint32, _batcherHash [32]byte, _gasLimit uint64, _baseFee *big.Int, _unsafeBlockSigner common.Address, _config ResourceMeteringResourceConfig) (common.Address, *types.Transaction, *SystemConfig, error) {
	parsed, err := SystemConfigMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(SystemConfigBin), backend, _owner, _basefeeScalar, _blobbasefeeScalar, _batcherHash, _gasLimit, _baseFee, _unsafeBlockSigner, _config)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SystemConfig{SystemConfigCaller: SystemConfigCaller{contract: contract}, SystemConfigTransactor: SystemConfigTransactor{contract: contract}, SystemConfigFilterer: SystemConfigFilterer{contract: contract}}, nil
}

// SystemConfig is an auto generated Go binding around an Ethereum contract.
type SystemConfig struct {
	SystemConfigCaller     // Read-only binding to the contract
	SystemConfigTransactor // Write-only binding to the contract
	SystemConfigFilterer   // Log filterer for contract events
}

// SystemConfigCaller is an auto generated read-only Go binding around an Ethereum contract.
type SystemConfigCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SystemConfigTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SystemConfigTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SystemConfigFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SystemConfigFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SystemConfigSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SystemConfigSession struct {
	Contract     *SystemConfig     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SystemConfigCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SystemConfigCallerSession struct {
	Contract *SystemConfigCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// SystemConfigTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SystemConfigTransactorSession struct {
	Contract     *SystemConfigTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// SystemConfigRaw is an auto generated low-level Go binding around an Ethereum contract.
type SystemConfigRaw struct {
	Contract *SystemConfig // Generic contract binding to access the raw methods on
}

// SystemConfigCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SystemConfigCallerRaw struct {
	Contract *SystemConfigCaller // Generic read-only contract binding to access the raw methods on
}

// SystemConfigTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SystemConfigTransactorRaw struct {
	Contract *SystemConfigTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSystemConfig creates a new instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfig(address common.Address, backend bind.ContractBackend) (*SystemConfig, error) {
	contract, err := bindSystemConfig(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SystemConfig{SystemConfigCaller: SystemConfigCaller{contract: contract}, SystemConfigTransactor: SystemConfigTransactor{contract: contract}, SystemConfigFilterer: SystemConfigFilterer{contract: contract}}, nil
}

// NewSystemConfigCaller creates a new read-only instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfigCaller(address common.Address, caller bind.ContractCaller) (*SystemConfigCaller, error) {
	contract, err := bindSystemConfig(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SystemConfigCaller{contract: contract}, nil
}

// NewSystemConfigTransactor creates a new write-only instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfigTransactor(address common.Address, transactor bind.ContractTransactor) (*SystemConfigTransactor, error) {
	contract, err := bindSystemConfig(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SystemConfigTransactor{contract: contract}, nil
}

// NewSystemConfigFilterer creates a new log filterer instance of SystemConfig, bound to a specific deployed contract.
func NewSystemConfigFilterer(address common.Address, filterer bind.ContractFilterer) (*SystemConfigFilterer, error) {
	contract, err := bindSystemConfig(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SystemConfigFilterer{contract: contract}, nil
}

// bindSystemConfig binds a generic wrapper to an already deployed contract.
func bindSystemConfig(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := SystemConfigMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SystemConfig *SystemConfigRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SystemConfig.Contract.SystemConfigCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SystemConfig *SystemConfigRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SystemConfig.Contract.SystemConfigTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SystemConfig *SystemConfigRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SystemConfig.Contract.SystemConfigTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SystemConfig *SystemConfigCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SystemConfig.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SystemConfig *SystemConfigTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SystemConfig.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SystemConfig *SystemConfigTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SystemConfig.Contract.contract.Transact(opts, method, params...)
}

// UNSAFEBLOCKSIGNERSLOT is a free data retrieval call binding the contract method 0x4f16540b.
//
// Solidity: function UNSAFE_BLOCK_SIGNER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) UNSAFEBLOCKSIGNERSLOT(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "UNSAFE_BLOCK_SIGNER_SLOT")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// UNSAFEBLOCKSIGNERSLOT is a free data retrieval call binding the contract method 0x4f16540b.
//
// Solidity: function UNSAFE_BLOCK_SIGNER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) UNSAFEBLOCKSIGNERSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.UNSAFEBLOCKSIGNERSLOT(&_SystemConfig.CallOpts)
}

// UNSAFEBLOCKSIGNERSLOT is a free data retrieval call binding the contract method 0x4f16540b.
//
// Solidity: function UNSAFE_BLOCK_SIGNER_SLOT() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) UNSAFEBLOCKSIGNERSLOT() ([32]byte, error) {
	return _SystemConfig.Contract.UNSAFEBLOCKSIGNERSLOT(&_SystemConfig.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) VERSION(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_SystemConfig *SystemConfigSession) VERSION() (*big.Int, error) {
	return _SystemConfig.Contract.VERSION(&_SystemConfig.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) VERSION() (*big.Int, error) {
	return _SystemConfig.Contract.VERSION(&_SystemConfig.CallOpts)
}

// BaseFee is a free data retrieval call binding the contract method 0x6ef25c3a.
//
// Solidity: function baseFee() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) BaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "baseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BaseFee is a free data retrieval call binding the contract method 0x6ef25c3a.
//
// Solidity: function baseFee() view returns(uint256)
func (_SystemConfig *SystemConfigSession) BaseFee() (*big.Int, error) {
	return _SystemConfig.Contract.BaseFee(&_SystemConfig.CallOpts)
}

// BaseFee is a free data retrieval call binding the contract method 0x6ef25c3a.
//
// Solidity: function baseFee() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) BaseFee() (*big.Int, error) {
	return _SystemConfig.Contract.BaseFee(&_SystemConfig.CallOpts)
}

// BasefeeScalar is a free data retrieval call binding the contract method 0xbfb14fb7.
//
// Solidity: function basefeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigCaller) BasefeeScalar(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "basefeeScalar")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// BasefeeScalar is a free data retrieval call binding the contract method 0xbfb14fb7.
//
// Solidity: function basefeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigSession) BasefeeScalar() (uint32, error) {
	return _SystemConfig.Contract.BasefeeScalar(&_SystemConfig.CallOpts)
}

// BasefeeScalar is a free data retrieval call binding the contract method 0xbfb14fb7.
//
// Solidity: function basefeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigCallerSession) BasefeeScalar() (uint32, error) {
	return _SystemConfig.Contract.BasefeeScalar(&_SystemConfig.CallOpts)
}

// BatcherHash is a free data retrieval call binding the contract method 0xe81b2c6d.
//
// Solidity: function batcherHash() view returns(bytes32)
func (_SystemConfig *SystemConfigCaller) BatcherHash(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "batcherHash")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// BatcherHash is a free data retrieval call binding the contract method 0xe81b2c6d.
//
// Solidity: function batcherHash() view returns(bytes32)
func (_SystemConfig *SystemConfigSession) BatcherHash() ([32]byte, error) {
	return _SystemConfig.Contract.BatcherHash(&_SystemConfig.CallOpts)
}

// BatcherHash is a free data retrieval call binding the contract method 0xe81b2c6d.
//
// Solidity: function batcherHash() view returns(bytes32)
func (_SystemConfig *SystemConfigCallerSession) BatcherHash() ([32]byte, error) {
	return _SystemConfig.Contract.BatcherHash(&_SystemConfig.CallOpts)
}

// BlobbasefeeScalar is a free data retrieval call binding the contract method 0xec707517.
//
// Solidity: function blobbasefeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigCaller) BlobbasefeeScalar(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "blobbasefeeScalar")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// BlobbasefeeScalar is a free data retrieval call binding the contract method 0xec707517.
//
// Solidity: function blobbasefeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigSession) BlobbasefeeScalar() (uint32, error) {
	return _SystemConfig.Contract.BlobbasefeeScalar(&_SystemConfig.CallOpts)
}

// BlobbasefeeScalar is a free data retrieval call binding the contract method 0xec707517.
//
// Solidity: function blobbasefeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigCallerSession) BlobbasefeeScalar() (uint32, error) {
	return _SystemConfig.Contract.BlobbasefeeScalar(&_SystemConfig.CallOpts)
}

// DaFootprintGasScalar is a free data retrieval call binding the contract method 0xfe3d5710.
//
// Solidity: function daFootprintGasScalar() view returns(uint16)
func (_SystemConfig *SystemConfigCaller) DaFootprintGasScalar(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "daFootprintGasScalar")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// DaFootprintGasScalar is a free data retrieval call binding the contract method 0xfe3d5710.
//
// Solidity: function daFootprintGasScalar() view returns(uint16)
func (_SystemConfig *SystemConfigSession) DaFootprintGasScalar() (uint16, error) {
	return _SystemConfig.Contract.DaFootprintGasScalar(&_SystemConfig.CallOpts)
}

// DaFootprintGasScalar is a free data retrieval call binding the contract method 0xfe3d5710.
//
// Solidity: function daFootprintGasScalar() view returns(uint16)
func (_SystemConfig *SystemConfigCallerSession) DaFootprintGasScalar() (uint16, error) {
	return _SystemConfig.Contract.DaFootprintGasScalar(&_SystemConfig.CallOpts)
}

// Eip1559Denominator is a free data retrieval call binding the contract method 0xd220a9e0.
//
// Solidity: function eip1559Denominator() view returns(uint32)
func (_SystemConfig *SystemConfigCaller) Eip1559Denominator(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "eip1559Denominator")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// Eip1559Denominator is a free data retrieval call binding the contract method 0xd220a9e0.
//
// Solidity: function eip1559Denominator() view returns(uint32)
func (_SystemConfig *SystemConfigSession) Eip1559Denominator() (uint32, error) {
	return _SystemConfig.Contract.Eip1559Denominator(&_SystemConfig.CallOpts)
}

// Eip1559Denominator is a free data retrieval call binding the contract method 0xd220a9e0.
//
// Solidity: function eip1559Denominator() view returns(uint32)
func (_SystemConfig *SystemConfigCallerSession) Eip1559Denominator() (uint32, error) {
	return _SystemConfig.Contract.Eip1559Denominator(&_SystemConfig.CallOpts)
}

// Eip1559Elasticity is a free data retrieval call binding the contract method 0xc9ff2d16.
//
// Solidity: function eip1559Elasticity() view returns(uint32)
func (_SystemConfig *SystemConfigCaller) Eip1559Elasticity(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "eip1559Elasticity")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// Eip1559Elasticity is a free data retrieval call binding the contract method 0xc9ff2d16.
//
// Solidity: function eip1559Elasticity() view returns(uint32)
func (_SystemConfig *SystemConfigSession) Eip1559Elasticity() (uint32, error) {
	return _SystemConfig.Contract.Eip1559Elasticity(&_SystemConfig.CallOpts)
}

// Eip1559Elasticity is a free data retrieval call binding the contract method 0xc9ff2d16.
//
// Solidity: function eip1559Elasticity() view returns(uint32)
func (_SystemConfig *SystemConfigCallerSession) Eip1559Elasticity() (uint32, error) {
	return _SystemConfig.Contract.Eip1559Elasticity(&_SystemConfig.CallOpts)
}

// GasLimit is a free data retrieval call binding the contract method 0xf68016b7.
//
// Solidity: function gasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCaller) GasLimit(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "gasLimit")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// GasLimit is a free data retrieval call binding the contract method 0xf68016b7.
//
// Solidity: function gasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigSession) GasLimit() (uint64, error) {
	return _SystemConfig.Contract.GasLimit(&_SystemConfig.CallOpts)
}

// GasLimit is a free data retrieval call binding the contract method 0xf68016b7.
//
// Solidity: function gasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCallerSession) GasLimit() (uint64, error) {
	return _SystemConfig.Contract.GasLimit(&_SystemConfig.CallOpts)
}

// MinBaseFee is a free data retrieval call binding the contract method 0xa62611a2.
//
// Solidity: function minBaseFee() view returns(uint64)
func (_SystemConfig *SystemConfigCaller) MinBaseFee(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "minBaseFee")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MinBaseFee is a free data retrieval call binding the contract method 0xa62611a2.
//
// Solidity: function minBaseFee() view returns(uint64)
func (_SystemConfig *SystemConfigSession) MinBaseFee() (uint64, error) {
	return _SystemConfig.Contract.MinBaseFee(&_SystemConfig.CallOpts)
}

// MinBaseFee is a free data retrieval call binding the contract method 0xa62611a2.
//
// Solidity: function minBaseFee() view returns(uint64)
func (_SystemConfig *SystemConfigCallerSession) MinBaseFee() (uint64, error) {
	return _SystemConfig.Contract.MinBaseFee(&_SystemConfig.CallOpts)
}

// MinimumGasLimit is a free data retrieval call binding the contract method 0x4add321d.
//
// Solidity: function minimumGasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCaller) MinimumGasLimit(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "minimumGasLimit")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MinimumGasLimit is a free data retrieval call binding the contract method 0x4add321d.
//
// Solidity: function minimumGasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigSession) MinimumGasLimit() (uint64, error) {
	return _SystemConfig.Contract.MinimumGasLimit(&_SystemConfig.CallOpts)
}

// MinimumGasLimit is a free data retrieval call binding the contract method 0x4add321d.
//
// Solidity: function minimumGasLimit() view returns(uint64)
func (_SystemConfig *SystemConfigCallerSession) MinimumGasLimit() (uint64, error) {
	return _SystemConfig.Contract.MinimumGasLimit(&_SystemConfig.CallOpts)
}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint64)
func (_SystemConfig *SystemConfigCaller) OperatorFeeConstant(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "operatorFeeConstant")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint64)
func (_SystemConfig *SystemConfigSession) OperatorFeeConstant() (uint64, error) {
	return _SystemConfig.Contract.OperatorFeeConstant(&_SystemConfig.CallOpts)
}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint64)
func (_SystemConfig *SystemConfigCallerSession) OperatorFeeConstant() (uint64, error) {
	return _SystemConfig.Contract.OperatorFeeConstant(&_SystemConfig.CallOpts)
}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigCaller) OperatorFeeScalar(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "operatorFeeScalar")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigSession) OperatorFeeScalar() (uint32, error) {
	return _SystemConfig.Contract.OperatorFeeScalar(&_SystemConfig.CallOpts)
}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint32)
func (_SystemConfig *SystemConfigCallerSession) OperatorFeeScalar() (uint32, error) {
	return _SystemConfig.Contract.OperatorFeeScalar(&_SystemConfig.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) Overhead(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "overhead")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_SystemConfig *SystemConfigSession) Overhead() (*big.Int, error) {
	return _SystemConfig.Contract.Overhead(&_SystemConfig.CallOpts)
}

// Overhead is a free data retrieval call binding the contract method 0x0c18c162.
//
// Solidity: function overhead() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) Overhead() (*big.Int, error) {
	return _SystemConfig.Contract.Overhead(&_SystemConfig.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SystemConfig *SystemConfigCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SystemConfig *SystemConfigSession) Owner() (common.Address, error) {
	return _SystemConfig.Contract.Owner(&_SystemConfig.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_SystemConfig *SystemConfigCallerSession) Owner() (common.Address, error) {
	return _SystemConfig.Contract.Owner(&_SystemConfig.CallOpts)
}

// ResourceConfig is a free data retrieval call binding the contract method 0xcc731b02.
//
// Solidity: function resourceConfig() view returns((uint32,uint8,uint8,uint32,uint32,uint128))
func (_SystemConfig *SystemConfigCaller) ResourceConfig(opts *bind.CallOpts) (ResourceMeteringResourceConfig, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "resourceConfig")

	if err != nil {
		return *new(ResourceMeteringResourceConfig), err
	}

	out0 := *abi.ConvertType(out[0], new(ResourceMeteringResourceConfig)).(*ResourceMeteringResourceConfig)

	return out0, err

}

// ResourceConfig is a free data retrieval call binding the contract method 0xcc731b02.
//
// Solidity: function resourceConfig() view returns((uint32,uint8,uint8,uint32,uint32,uint128))
func (_SystemConfig *SystemConfigSession) ResourceConfig() (ResourceMeteringResourceConfig, error) {
	return _SystemConfig.Contract.ResourceConfig(&_SystemConfig.CallOpts)
}

// ResourceConfig is a free data retrieval call binding the contract method 0xcc731b02.
//
// Solidity: function resourceConfig() view returns((uint32,uint8,uint8,uint32,uint32,uint128))
func (_SystemConfig *SystemConfigCallerSession) ResourceConfig() (ResourceMeteringResourceConfig, error) {
	return _SystemConfig.Contract.ResourceConfig(&_SystemConfig.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_SystemConfig *SystemConfigCaller) Scalar(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "scalar")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_SystemConfig *SystemConfigSession) Scalar() (*big.Int, error) {
	return _SystemConfig.Contract.Scalar(&_SystemConfig.CallOpts)
}

// Scalar is a free data retrieval call binding the contract method 0xf45e65d8.
//
// Solidity: function scalar() view returns(uint256)
func (_SystemConfig *SystemConfigCallerSession) Scalar() (*big.Int, error) {
	return _SystemConfig.Contract.Scalar(&_SystemConfig.CallOpts)
}

// UnsafeBlockSigner is a free data retrieval call binding the contract method 0x1fd19ee1.
//
// Solidity: function unsafeBlockSigner() view returns(address)
func (_SystemConfig *SystemConfigCaller) UnsafeBlockSigner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "unsafeBlockSigner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// UnsafeBlockSigner is a free data retrieval call binding the contract method 0x1fd19ee1.
//
// Solidity: function unsafeBlockSigner() view returns(address)
func (_SystemConfig *SystemConfigSession) UnsafeBlockSigner() (common.Address, error) {
	return _SystemConfig.Contract.UnsafeBlockSigner(&_SystemConfig.CallOpts)
}

// UnsafeBlockSigner is a free data retrieval call binding the contract method 0x1fd19ee1.
//
// Solidity: function unsafeBlockSigner() view returns(address)
func (_SystemConfig *SystemConfigCallerSession) UnsafeBlockSigner() (common.Address, error) {
	return _SystemConfig.Contract.UnsafeBlockSigner(&_SystemConfig.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SystemConfig *SystemConfigCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _SystemConfig.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SystemConfig *SystemConfigSession) Version() (string, error) {
	return _SystemConfig.Contract.Version(&_SystemConfig.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_SystemConfig *SystemConfigCallerSession) Version() (string, error) {
	return _SystemConfig.Contract.Version(&_SystemConfig.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x029865ef.
//
// Solidity: function initialize(address _owner, uint32 _basefeeScalar, uint32 _blobbasefeeScalar, bytes32 _batcherHash, uint64 _gasLimit, uint256 _baseFee, address _unsafeBlockSigner, (uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigTransactor) Initialize(opts *bind.TransactOpts, _owner common.Address, _basefeeScalar uint32, _blobbasefeeScalar uint32, _batcherHash [32]byte, _gasLimit uint64, _baseFee *big.Int, _unsafeBlockSigner common.Address, _config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "initialize", _owner, _basefeeScalar, _blobbasefeeScalar, _batcherHash, _gasLimit, _baseFee, _unsafeBlockSigner, _config)
}

// Initialize is a paid mutator transaction binding the contract method 0x029865ef.
//
// Solidity: function initialize(address _owner, uint32 _basefeeScalar, uint32 _blobbasefeeScalar, bytes32 _batcherHash, uint64 _gasLimit, uint256 _baseFee, address _unsafeBlockSigner, (uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigSession) Initialize(_owner common.Address, _basefeeScalar uint32, _blobbasefeeScalar uint32, _batcherHash [32]byte, _gasLimit uint64, _baseFee *big.Int, _unsafeBlockSigner common.Address, _config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.Contract.Initialize(&_SystemConfig.TransactOpts, _owner, _basefeeScalar, _blobbasefeeScalar, _batcherHash, _gasLimit, _baseFee, _unsafeBlockSigner, _config)
}

// Initialize is a paid mutator transaction binding the contract method 0x029865ef.
//
// Solidity: function initialize(address _owner, uint32 _basefeeScalar, uint32 _blobbasefeeScalar, bytes32 _batcherHash, uint64 _gasLimit, uint256 _baseFee, address _unsafeBlockSigner, (uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigTransactorSession) Initialize(_owner common.Address, _basefeeScalar uint32, _blobbasefeeScalar uint32, _batcherHash [32]byte, _gasLimit uint64, _baseFee *big.Int, _unsafeBlockSigner common.Address, _config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.Contract.Initialize(&_SystemConfig.TransactOpts, _owner, _basefeeScalar, _blobbasefeeScalar, _batcherHash, _gasLimit, _baseFee, _unsafeBlockSigner, _config)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SystemConfig *SystemConfigTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SystemConfig *SystemConfigSession) RenounceOwnership() (*types.Transaction, error) {
	return _SystemConfig.Contract.RenounceOwnership(&_SystemConfig.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_SystemConfig *SystemConfigTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _SystemConfig.Contract.RenounceOwnership(&_SystemConfig.TransactOpts)
}

// SetBaseFee is a paid mutator transaction binding the contract method 0x46860698.
//
// Solidity: function setBaseFee(uint256 _baseFee) returns()
func (_SystemConfig *SystemConfigTransactor) SetBaseFee(opts *bind.TransactOpts, _baseFee *big.Int) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setBaseFee", _baseFee)
}

// SetBaseFee is a paid mutator transaction binding the contract method 0x46860698.
//
// Solidity: function setBaseFee(uint256 _baseFee) returns()
func (_SystemConfig *SystemConfigSession) SetBaseFee(_baseFee *big.Int) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetBaseFee(&_SystemConfig.TransactOpts, _baseFee)
}

// SetBaseFee is a paid mutator transaction binding the contract method 0x46860698.
//
// Solidity: function setBaseFee(uint256 _baseFee) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetBaseFee(_baseFee *big.Int) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetBaseFee(&_SystemConfig.TransactOpts, _baseFee)
}

// SetBatcherHash is a paid mutator transaction binding the contract method 0xc9b26f61.
//
// Solidity: function setBatcherHash(bytes32 _batcherHash) returns()
func (_SystemConfig *SystemConfigTransactor) SetBatcherHash(opts *bind.TransactOpts, _batcherHash [32]byte) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setBatcherHash", _batcherHash)
}

// SetBatcherHash is a paid mutator transaction binding the contract method 0xc9b26f61.
//
// Solidity: function setBatcherHash(bytes32 _batcherHash) returns()
func (_SystemConfig *SystemConfigSession) SetBatcherHash(_batcherHash [32]byte) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetBatcherHash(&_SystemConfig.TransactOpts, _batcherHash)
}

// SetBatcherHash is a paid mutator transaction binding the contract method 0xc9b26f61.
//
// Solidity: function setBatcherHash(bytes32 _batcherHash) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetBatcherHash(_batcherHash [32]byte) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetBatcherHash(&_SystemConfig.TransactOpts, _batcherHash)
}

// SetDAFootprintGasScalar is a paid mutator transaction binding the contract method 0x20f06fdc.
//
// Solidity: function setDAFootprintGasScalar(uint16 _daFootprintGasScalar) returns()
func (_SystemConfig *SystemConfigTransactor) SetDAFootprintGasScalar(opts *bind.TransactOpts, _daFootprintGasScalar uint16) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setDAFootprintGasScalar", _daFootprintGasScalar)
}

// SetDAFootprintGasScalar is a paid mutator transaction binding the contract method 0x20f06fdc.
//
// Solidity: function setDAFootprintGasScalar(uint16 _daFootprintGasScalar) returns()
func (_SystemConfig *SystemConfigSession) SetDAFootprintGasScalar(_daFootprintGasScalar uint16) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetDAFootprintGasScalar(&_SystemConfig.TransactOpts, _daFootprintGasScalar)
}

// SetDAFootprintGasScalar is a paid mutator transaction binding the contract method 0x20f06fdc.
//
// Solidity: function setDAFootprintGasScalar(uint16 _daFootprintGasScalar) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetDAFootprintGasScalar(_daFootprintGasScalar uint16) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetDAFootprintGasScalar(&_SystemConfig.TransactOpts, _daFootprintGasScalar)
}

// SetEIP1559Params is a paid mutator transaction binding the contract method 0xc0fd4b41.
//
// Solidity: function setEIP1559Params(uint32 _denominator, uint32 _elasticity) returns()
func (_SystemConfig *SystemConfigTransactor) SetEIP1559Params(opts *bind.TransactOpts, _denominator uint32, _elasticity uint32) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setEIP1559Params", _denominator, _elasticity)
}

// SetEIP1559Params is a paid mutator transaction binding the contract method 0xc0fd4b41.
//
// Solidity: function setEIP1559Params(uint32 _denominator, uint32 _elasticity) returns()
func (_SystemConfig *SystemConfigSession) SetEIP1559Params(_denominator uint32, _elasticity uint32) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetEIP1559Params(&_SystemConfig.TransactOpts, _denominator, _elasticity)
}

// SetEIP1559Params is a paid mutator transaction binding the contract method 0xc0fd4b41.
//
// Solidity: function setEIP1559Params(uint32 _denominator, uint32 _elasticity) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetEIP1559Params(_denominator uint32, _elasticity uint32) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetEIP1559Params(&_SystemConfig.TransactOpts, _denominator, _elasticity)
}

// SetGasConfig is a paid mutator transaction binding the contract method 0x935f029e.
//
// Solidity: function setGasConfig(uint256 _overhead, uint256 _scalar) returns()
func (_SystemConfig *SystemConfigTransactor) SetGasConfig(opts *bind.TransactOpts, _overhead *big.Int, _scalar *big.Int) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setGasConfig", _overhead, _scalar)
}

// SetGasConfig is a paid mutator transaction binding the contract method 0x935f029e.
//
// Solidity: function setGasConfig(uint256 _overhead, uint256 _scalar) returns()
func (_SystemConfig *SystemConfigSession) SetGasConfig(_overhead *big.Int, _scalar *big.Int) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasConfig(&_SystemConfig.TransactOpts, _overhead, _scalar)
}

// SetGasConfig is a paid mutator transaction binding the contract method 0x935f029e.
//
// Solidity: function setGasConfig(uint256 _overhead, uint256 _scalar) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetGasConfig(_overhead *big.Int, _scalar *big.Int) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasConfig(&_SystemConfig.TransactOpts, _overhead, _scalar)
}

// SetGasConfigArsia is a paid mutator transaction binding the contract method 0xca76bd6c.
//
// Solidity: function setGasConfigArsia(uint32 _basefeeScalar, uint32 _blobbasefeeScalar) returns()
func (_SystemConfig *SystemConfigTransactor) SetGasConfigArsia(opts *bind.TransactOpts, _basefeeScalar uint32, _blobbasefeeScalar uint32) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setGasConfigArsia", _basefeeScalar, _blobbasefeeScalar)
}

// SetGasConfigArsia is a paid mutator transaction binding the contract method 0xca76bd6c.
//
// Solidity: function setGasConfigArsia(uint32 _basefeeScalar, uint32 _blobbasefeeScalar) returns()
func (_SystemConfig *SystemConfigSession) SetGasConfigArsia(_basefeeScalar uint32, _blobbasefeeScalar uint32) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasConfigArsia(&_SystemConfig.TransactOpts, _basefeeScalar, _blobbasefeeScalar)
}

// SetGasConfigArsia is a paid mutator transaction binding the contract method 0xca76bd6c.
//
// Solidity: function setGasConfigArsia(uint32 _basefeeScalar, uint32 _blobbasefeeScalar) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetGasConfigArsia(_basefeeScalar uint32, _blobbasefeeScalar uint32) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasConfigArsia(&_SystemConfig.TransactOpts, _basefeeScalar, _blobbasefeeScalar)
}

// SetGasLimit is a paid mutator transaction binding the contract method 0xb40a817c.
//
// Solidity: function setGasLimit(uint64 _gasLimit) returns()
func (_SystemConfig *SystemConfigTransactor) SetGasLimit(opts *bind.TransactOpts, _gasLimit uint64) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setGasLimit", _gasLimit)
}

// SetGasLimit is a paid mutator transaction binding the contract method 0xb40a817c.
//
// Solidity: function setGasLimit(uint64 _gasLimit) returns()
func (_SystemConfig *SystemConfigSession) SetGasLimit(_gasLimit uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasLimit(&_SystemConfig.TransactOpts, _gasLimit)
}

// SetGasLimit is a paid mutator transaction binding the contract method 0xb40a817c.
//
// Solidity: function setGasLimit(uint64 _gasLimit) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetGasLimit(_gasLimit uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetGasLimit(&_SystemConfig.TransactOpts, _gasLimit)
}

// SetMinBaseFee is a paid mutator transaction binding the contract method 0x7616f0e8.
//
// Solidity: function setMinBaseFee(uint64 _minBaseFee) returns()
func (_SystemConfig *SystemConfigTransactor) SetMinBaseFee(opts *bind.TransactOpts, _minBaseFee uint64) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setMinBaseFee", _minBaseFee)
}

// SetMinBaseFee is a paid mutator transaction binding the contract method 0x7616f0e8.
//
// Solidity: function setMinBaseFee(uint64 _minBaseFee) returns()
func (_SystemConfig *SystemConfigSession) SetMinBaseFee(_minBaseFee uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetMinBaseFee(&_SystemConfig.TransactOpts, _minBaseFee)
}

// SetMinBaseFee is a paid mutator transaction binding the contract method 0x7616f0e8.
//
// Solidity: function setMinBaseFee(uint64 _minBaseFee) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetMinBaseFee(_minBaseFee uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetMinBaseFee(&_SystemConfig.TransactOpts, _minBaseFee)
}

// SetOperatorFeeScalars is a paid mutator transaction binding the contract method 0x155b6c6f.
//
// Solidity: function setOperatorFeeScalars(uint32 _operatorFeeScalar, uint64 _operatorFeeConstant) returns()
func (_SystemConfig *SystemConfigTransactor) SetOperatorFeeScalars(opts *bind.TransactOpts, _operatorFeeScalar uint32, _operatorFeeConstant uint64) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setOperatorFeeScalars", _operatorFeeScalar, _operatorFeeConstant)
}

// SetOperatorFeeScalars is a paid mutator transaction binding the contract method 0x155b6c6f.
//
// Solidity: function setOperatorFeeScalars(uint32 _operatorFeeScalar, uint64 _operatorFeeConstant) returns()
func (_SystemConfig *SystemConfigSession) SetOperatorFeeScalars(_operatorFeeScalar uint32, _operatorFeeConstant uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetOperatorFeeScalars(&_SystemConfig.TransactOpts, _operatorFeeScalar, _operatorFeeConstant)
}

// SetOperatorFeeScalars is a paid mutator transaction binding the contract method 0x155b6c6f.
//
// Solidity: function setOperatorFeeScalars(uint32 _operatorFeeScalar, uint64 _operatorFeeConstant) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetOperatorFeeScalars(_operatorFeeScalar uint32, _operatorFeeConstant uint64) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetOperatorFeeScalars(&_SystemConfig.TransactOpts, _operatorFeeScalar, _operatorFeeConstant)
}

// SetResourceConfig is a paid mutator transaction binding the contract method 0xc71973f6.
//
// Solidity: function setResourceConfig((uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigTransactor) SetResourceConfig(opts *bind.TransactOpts, _config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setResourceConfig", _config)
}

// SetResourceConfig is a paid mutator transaction binding the contract method 0xc71973f6.
//
// Solidity: function setResourceConfig((uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigSession) SetResourceConfig(_config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetResourceConfig(&_SystemConfig.TransactOpts, _config)
}

// SetResourceConfig is a paid mutator transaction binding the contract method 0xc71973f6.
//
// Solidity: function setResourceConfig((uint32,uint8,uint8,uint32,uint32,uint128) _config) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetResourceConfig(_config ResourceMeteringResourceConfig) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetResourceConfig(&_SystemConfig.TransactOpts, _config)
}

// SetUnsafeBlockSigner is a paid mutator transaction binding the contract method 0x18d13918.
//
// Solidity: function setUnsafeBlockSigner(address _unsafeBlockSigner) returns()
func (_SystemConfig *SystemConfigTransactor) SetUnsafeBlockSigner(opts *bind.TransactOpts, _unsafeBlockSigner common.Address) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "setUnsafeBlockSigner", _unsafeBlockSigner)
}

// SetUnsafeBlockSigner is a paid mutator transaction binding the contract method 0x18d13918.
//
// Solidity: function setUnsafeBlockSigner(address _unsafeBlockSigner) returns()
func (_SystemConfig *SystemConfigSession) SetUnsafeBlockSigner(_unsafeBlockSigner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetUnsafeBlockSigner(&_SystemConfig.TransactOpts, _unsafeBlockSigner)
}

// SetUnsafeBlockSigner is a paid mutator transaction binding the contract method 0x18d13918.
//
// Solidity: function setUnsafeBlockSigner(address _unsafeBlockSigner) returns()
func (_SystemConfig *SystemConfigTransactorSession) SetUnsafeBlockSigner(_unsafeBlockSigner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.SetUnsafeBlockSigner(&_SystemConfig.TransactOpts, _unsafeBlockSigner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SystemConfig *SystemConfigTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _SystemConfig.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SystemConfig *SystemConfigSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.TransferOwnership(&_SystemConfig.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_SystemConfig *SystemConfigTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _SystemConfig.Contract.TransferOwnership(&_SystemConfig.TransactOpts, newOwner)
}

// SystemConfigConfigUpdateIterator is returned from FilterConfigUpdate and is used to iterate over the raw logs and unpacked data for ConfigUpdate events raised by the SystemConfig contract.
type SystemConfigConfigUpdateIterator struct {
	Event *SystemConfigConfigUpdate // Event containing the contract specifics and raw log

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
func (it *SystemConfigConfigUpdateIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SystemConfigConfigUpdate)
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
		it.Event = new(SystemConfigConfigUpdate)
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
func (it *SystemConfigConfigUpdateIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SystemConfigConfigUpdateIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SystemConfigConfigUpdate represents a ConfigUpdate event raised by the SystemConfig contract.
type SystemConfigConfigUpdate struct {
	Version    *big.Int
	UpdateType uint8
	Data       []byte
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterConfigUpdate is a free log retrieval operation binding the contract event 0x1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be.
//
// Solidity: event ConfigUpdate(uint256 indexed version, uint8 indexed updateType, bytes data)
func (_SystemConfig *SystemConfigFilterer) FilterConfigUpdate(opts *bind.FilterOpts, version []*big.Int, updateType []uint8) (*SystemConfigConfigUpdateIterator, error) {

	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}
	var updateTypeRule []interface{}
	for _, updateTypeItem := range updateType {
		updateTypeRule = append(updateTypeRule, updateTypeItem)
	}

	logs, sub, err := _SystemConfig.contract.FilterLogs(opts, "ConfigUpdate", versionRule, updateTypeRule)
	if err != nil {
		return nil, err
	}
	return &SystemConfigConfigUpdateIterator{contract: _SystemConfig.contract, event: "ConfigUpdate", logs: logs, sub: sub}, nil
}

// WatchConfigUpdate is a free log subscription operation binding the contract event 0x1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be.
//
// Solidity: event ConfigUpdate(uint256 indexed version, uint8 indexed updateType, bytes data)
func (_SystemConfig *SystemConfigFilterer) WatchConfigUpdate(opts *bind.WatchOpts, sink chan<- *SystemConfigConfigUpdate, version []*big.Int, updateType []uint8) (event.Subscription, error) {

	var versionRule []interface{}
	for _, versionItem := range version {
		versionRule = append(versionRule, versionItem)
	}
	var updateTypeRule []interface{}
	for _, updateTypeItem := range updateType {
		updateTypeRule = append(updateTypeRule, updateTypeItem)
	}

	logs, sub, err := _SystemConfig.contract.WatchLogs(opts, "ConfigUpdate", versionRule, updateTypeRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SystemConfigConfigUpdate)
				if err := _SystemConfig.contract.UnpackLog(event, "ConfigUpdate", log); err != nil {
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

// ParseConfigUpdate is a log parse operation binding the contract event 0x1d2b0bda21d56b8bd12d4f94ebacffdfb35f5e226f84b461103bb8beab6353be.
//
// Solidity: event ConfigUpdate(uint256 indexed version, uint8 indexed updateType, bytes data)
func (_SystemConfig *SystemConfigFilterer) ParseConfigUpdate(log types.Log) (*SystemConfigConfigUpdate, error) {
	event := new(SystemConfigConfigUpdate)
	if err := _SystemConfig.contract.UnpackLog(event, "ConfigUpdate", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SystemConfigInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the SystemConfig contract.
type SystemConfigInitializedIterator struct {
	Event *SystemConfigInitialized // Event containing the contract specifics and raw log

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
func (it *SystemConfigInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SystemConfigInitialized)
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
		it.Event = new(SystemConfigInitialized)
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
func (it *SystemConfigInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SystemConfigInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SystemConfigInitialized represents a Initialized event raised by the SystemConfig contract.
type SystemConfigInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_SystemConfig *SystemConfigFilterer) FilterInitialized(opts *bind.FilterOpts) (*SystemConfigInitializedIterator, error) {

	logs, sub, err := _SystemConfig.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &SystemConfigInitializedIterator{contract: _SystemConfig.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_SystemConfig *SystemConfigFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *SystemConfigInitialized) (event.Subscription, error) {

	logs, sub, err := _SystemConfig.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SystemConfigInitialized)
				if err := _SystemConfig.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_SystemConfig *SystemConfigFilterer) ParseInitialized(log types.Log) (*SystemConfigInitialized, error) {
	event := new(SystemConfigInitialized)
	if err := _SystemConfig.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// SystemConfigOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the SystemConfig contract.
type SystemConfigOwnershipTransferredIterator struct {
	Event *SystemConfigOwnershipTransferred // Event containing the contract specifics and raw log

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
func (it *SystemConfigOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(SystemConfigOwnershipTransferred)
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
		it.Event = new(SystemConfigOwnershipTransferred)
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
func (it *SystemConfigOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *SystemConfigOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// SystemConfigOwnershipTransferred represents a OwnershipTransferred event raised by the SystemConfig contract.
type SystemConfigOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SystemConfig *SystemConfigFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*SystemConfigOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SystemConfig.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &SystemConfigOwnershipTransferredIterator{contract: _SystemConfig.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_SystemConfig *SystemConfigFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *SystemConfigOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _SystemConfig.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(SystemConfigOwnershipTransferred)
				if err := _SystemConfig.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
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
func (_SystemConfig *SystemConfigFilterer) ParseOwnershipTransferred(log types.Log) (*SystemConfigOwnershipTransferred, error) {
	event := new(SystemConfigOwnershipTransferred)
	if err := _SystemConfig.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
