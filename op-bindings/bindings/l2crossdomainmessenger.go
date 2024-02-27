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

// L2CrossDomainMessengerMetaData contains all meta data concerning the L2CrossDomainMessenger contract.
var L2CrossDomainMessengerMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_l1CrossDomainMessenger\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"}],\"name\":\"FailedRelayedMessage\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint8\",\"name\":\"version\",\"type\":\"uint8\"}],\"name\":\"Initialized\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"bytes32\",\"name\":\"msgHash\",\"type\":\"bytes32\"}],\"name\":\"RelayedMessage\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"target\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"message\",\"type\":\"bytes\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"messageNonce\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"gasLimit\",\"type\":\"uint256\"}],\"name\":\"SentMessage\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"sender\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"mntValue\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"ethValue\",\"type\":\"uint256\"}],\"name\":\"SentMessageExtension1\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"MESSAGE_VERSION\",\"outputs\":[{\"internalType\":\"uint16\",\"name\":\"\",\"type\":\"uint16\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_GAS_CALLDATA_OVERHEAD\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_GAS_DYNAMIC_OVERHEAD_DENOMINATOR\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"MIN_GAS_DYNAMIC_OVERHEAD_NUMERATOR\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"OTHER_MESSENGER\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RELAY_CALL_OVERHEAD\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RELAY_CONSTANT_OVERHEAD\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RELAY_GAS_CHECK_BUFFER\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"RELAY_RESERVED_GAS\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes\",\"name\":\"_message\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"}],\"name\":\"baseGas\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"pure\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"failedMessages\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"initialize\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"l1CrossDomainMessenger\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"messageNonce\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_nonce\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_sender\",\"type\":\"address\"},{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"_mntValue\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_ethValue\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"_minGasLimit\",\"type\":\"uint256\"},{\"internalType\":\"bytes\",\"name\":\"_message\",\"type\":\"bytes\"}],\"name\":\"relayMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"_message\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"}],\"name\":\"sendMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"_ethAmount\",\"type\":\"uint256\"},{\"internalType\":\"address\",\"name\":\"_target\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"_message\",\"type\":\"bytes\"},{\"internalType\":\"uint32\",\"name\":\"_minGasLimit\",\"type\":\"uint32\"}],\"name\":\"sendMessage\",\"outputs\":[],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"name\":\"successfulMessages\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"version\",\"outputs\":[{\"internalType\":\"string\",\"name\":\"\",\"type\":\"string\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"xDomainMessageSender\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"}]",
	Bin: "0x6101006040523480156200001257600080fd5b5060405162002afc38038062002afc833981016040819052620000359162000243565b6001600160a01b038116608052600160a052600460c052600060e0526200005b62000062565b5062000275565b600054600160a81b900460ff16158080156200008b57506000546001600160a01b90910460ff16105b80620000c25750620000a830620001af60201b6200186d1760201c565b158015620000c25750600054600160a01b900460ff166001145b6200012b5760405162461bcd60e51b815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201526d191e481a5b9a5d1a585b1a5e995960921b60648201526084015b60405180910390fd5b6000805460ff60a01b1916600160a01b179055801562000159576000805460ff60a81b1916600160a81b1790555b62000163620001be565b8015620001ac576000805460ff60a81b19169055604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b50565b6001600160a01b03163b151590565b600054600160a81b900460ff166200022d5760405162461bcd60e51b815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201526a6e697469616c697a696e6760a81b606482015260840162000122565b60cc80546001600160a01b03191661dead179055565b6000602082840312156200025657600080fd5b81516001600160a01b03811681146200026e57600080fd5b9392505050565b60805160a05160c05160e051612831620002cb60003960006107af015260006107860152600061075d0152600081816102c401528181610335015281816104ef01528181610c28015261104f01526128316000f3fe60806040526004361061015f5760003560e01c806383a74074116100c0578063b1b1b20911610074578063b575e30011610059578063b575e300146103a9578063ecc70428146103bc578063ff8daf151461042157600080fd5b8063b1b1b20914610359578063b28ade251461038957600080fd5b80639fce812c116100a55780639fce812c146102b2578063a4e7f8bd146102e6578063a71198691461032657600080fd5b806383a740741461029b5780638cbeeef2146101fe57600080fd5b80634c1d6a69116101175780635644cfdf116100fc5780635644cfdf146102365780636e296e451461024c5780638129fc1c1461028657600080fd5b80634c1d6a69146101fe57806354fd4d501461021457600080fd5b80632828d7e8116101485780632828d7e8146101ac5780633dbb202b146101c15780633f827a5a146101d657600080fd5b8063028f85f7146101645780630c56849814610197575b600080fd5b34801561017057600080fd5b50610179601081565b60405167ffffffffffffffff90911681526020015b60405180910390f35b3480156101a357600080fd5b50610179603f81565b3480156101b857600080fd5b50610179604081565b6101d46101cf3660046120cf565b610434565b005b3480156101e257600080fd5b506101eb600181565b60405161ffff909116815260200161018e565b34801561020a57600080fd5b50610179619c4081565b34801561022057600080fd5b50610229610756565b60405161018e91906121aa565b34801561024257600080fd5b5061017961138881565b34801561025857600080fd5b506102616107f9565b60405173ffffffffffffffffffffffffffffffffffffffff909116815260200161018e565b34801561029257600080fd5b506101d46108e0565b3480156102a757600080fd5b5061017962030d4081565b3480156102be57600080fd5b506102617f000000000000000000000000000000000000000000000000000000000000000081565b3480156102f257600080fd5b506103166103013660046121bd565b60cf6020526000908152604090205460ff1681565b604051901515815260200161018e565b34801561033257600080fd5b507f0000000000000000000000000000000000000000000000000000000000000000610261565b34801561036557600080fd5b506103166103743660046121bd565b60cb6020526000908152604090205460ff1681565b34801561039557600080fd5b506101796103a43660046121d6565b610add565b6101d46103b736600461222a565b610b4d565b3480156103c857600080fd5b5061041360ce547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b60405190815260200161018e565b6101d461042f366004612299565b610e0d565b73ffffffffffffffffffffffffffffffffffffffff841632141580610457575034155b6104e8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f6f6e63652074617267657420697320616e20454f412c206d73672e76616c756560448201527f206d757374206265207a65726f0000000000000000000000000000000000000060648201526084015b60405180910390fd5b61062560007f0000000000000000000000000000000000000000000000000000000000000000610519868686610add565b7fff8daf150000000000000000000000000000000000000000000000000000000061058460ce547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b338a3460008a8d8d6040516024016105a398979695949392919061236d565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fffffffff0000000000000000000000000000000000000000000000000000000090931692909217909152611889565b8373ffffffffffffffffffffffffffffffffffffffff167fcb0f7ffd78f9aee47a248fae8db181db6eee833039123e026dcbff529522e52a3385856106aa60ce547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b866040516106bc9594939291906123d6565b60405180910390a2604080513481526000602082015233917fcf00802ba1f8c659140235227979ca08afaba336a9f9fdc4a5107ed9e8013d08910160405180910390a2505060ce80547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff808216600101167fffff0000000000000000000000000000000000000000000000000000000000009091161790555050565b60606107817f0000000000000000000000000000000000000000000000000000000000000000611919565b6107aa7f0000000000000000000000000000000000000000000000000000000000000000611919565b6107d37f0000000000000000000000000000000000000000000000000000000000000000611919565b6040516020016107e593929190612424565b604051602081830303815290604052905090565b60cc5460009073ffffffffffffffffffffffffffffffffffffffff167fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff2153016108c3576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603560248201527f43726f7373446f6d61696e4d657373656e6765723a2078446f6d61696e4d657360448201527f7361676553656e646572206973206e6f7420736574000000000000000000000060648201526084016104df565b5060cc5473ffffffffffffffffffffffffffffffffffffffff1690565b6000547501000000000000000000000000000000000000000000900460ff161580801561092b575060005460017401000000000000000000000000000000000000000090910460ff16105b8061095d5750303b15801561095d575060005474010000000000000000000000000000000000000000900460ff166001145b6109e9576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602e60248201527f496e697469616c697a61626c653a20636f6e747261637420697320616c72656160448201527f647920696e697469616c697a656400000000000000000000000000000000000060648201526084016104df565b600080547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff16740100000000000000000000000000000000000000001790558015610a6f57600080547fffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffffff1675010000000000000000000000000000000000000000001790555b610a77611a56565b8015610ada57600080547fffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffffff169055604051600181527f7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb38474024989060200160405180910390a15b50565b6000611388619c4080603f610af9604063ffffffff88166124c9565b610b039190612528565b610b0e6010886124c9565b610b1b9062030d4061254f565b610b25919061254f565b610b2f919061254f565b610b39919061254f565b610b43919061254f565b90505b9392505050565b73ffffffffffffffffffffffffffffffffffffffff841632141580610b70575034155b610bfc576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f6f6e63652074617267657420697320616e20454f412c206d73672e76616c756560448201527f206d757374206265207a65726f0000000000000000000000000000000000000060648201526084016104df565b8415610c2257610c2273deaddeaddeaddeaddeaddeaddeaddeaddead1111333088611b2f565b610cdb857f0000000000000000000000000000000000000000000000000000000000000000610c52868686610add565b7fff8daf1500000000000000000000000000000000000000000000000000000000610cbd60ce547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b338a348d8a8d8d6040516024016105a398979695949392919061257b565b8373ffffffffffffffffffffffffffffffffffffffff167fcb0f7ffd78f9aee47a248fae8db181db6eee833039123e026dcbff529522e52a338585610d6060ce547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff167e010000000000000000000000000000000000000000000000000000000000001790565b86604051610d729594939291906123d6565b60405180910390a2604080513481526020810187905233917fcf00802ba1f8c659140235227979ca08afaba336a9f9fdc4a5107ed9e8013d08910160405180910390a2505060ce80547dffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff808216600101167fffff000000000000000000000000000000000000000000000000000000000000909116179055505050565b60f088901c60028110610ec8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604d60248201527f43726f7373446f6d61696e4d657373656e6765723a206f6e6c7920766572736960448201527f6f6e2030206f722031206d657373616765732061726520737570706f7274656460648201527f20617420746869732074696d6500000000000000000000000000000000000000608482015260a4016104df565b8061ffff16600003610fca576000610f26888a86868080601f016020809104026020016040519081016040528093929190818152602001838380828437600081840152601f19601f820116905080830192505050505050508d611bca565b600081815260cb602052604090205490915060ff1615610fc8576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603760248201527f43726f7373446f6d61696e4d657373656e6765723a206c65676163792077697460448201527f6864726177616c20616c72656164792072656c6179656400000000000000000060648201526084016104df565b505b60006110118a8a8a8a8a8a8a8a8080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250611be992505050565b905073ffffffffffffffffffffffffffffffffffffffff7fffffffffffffffffffffffffeeeeffffffffffffffffffffffffffffffffeeef330181167f0000000000000000000000000000000000000000000000000000000000000000909116036110a957863414611085576110856125d3565b600081815260cf602052604090205460ff16156110a4576110a46125d3565b6111fb565b341561115d576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152605060248201527f43726f7373446f6d61696e4d657373656e6765723a2076616c7565206d75737460448201527f206265207a65726f20756e6c657373206d6573736167652069732066726f6d2060648201527f612073797374656d206164647265737300000000000000000000000000000000608482015260a4016104df565b600081815260cf602052604090205460ff166111fb576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603060248201527f43726f7373446f6d61696e4d657373656e6765723a206d65737361676520636160448201527f6e6e6f74206265207265706c617965640000000000000000000000000000000060648201526084016104df565b61120488611c0e565b156112b7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152604360248201527f43726f7373446f6d61696e4d657373656e6765723a2063616e6e6f742073656e60448201527f64206d65737361676520746f20626c6f636b65642073797374656d206164647260648201527f6573730000000000000000000000000000000000000000000000000000000000608482015260a4016104df565b600081815260cb602052604090205460ff1615611356576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603660248201527f43726f7373446f6d61696e4d657373656e6765723a206d65737361676520686160448201527f7320616c7265616479206265656e2072656c617965640000000000000000000060648201526084016104df565b61137785611368611388619c4061254f565b67ffffffffffffffff16611c63565b158061139d575060cc5473ffffffffffffffffffffffffffffffffffffffff1661dead14155b156114b657600081815260cf602052604080822080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660011790555182917f99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f91a27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff32016114af576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f43726f7373446f6d61696e4d657373656e6765723a206661696c656420746f2060448201527f72656c6179206d6573736167650000000000000000000000000000000000000060648201526084016104df565b5050611863565b6001861561156c576040517f095ea7b300000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8a1660048201526024810188905273deaddeaddeaddeaddeaddeaddeaddeaddead11119063095ea7b3906044016020604051808303816000875af1158015611545573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906115699190612602565b90505b60cc80547fffffffffffffffffffffffff00000000000000000000000000000000000000001673ffffffffffffffffffffffffffffffffffffffff8c1617905560006115fd8a619c405a6115c09190612624565b8b89898080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250611c8192505050565b60cc80547fffffffffffffffffffffffff00000000000000000000000000000000000000001661dead179055905087156116df576040517f095ea7b300000000000000000000000000000000000000000000000000000000815273ffffffffffffffffffffffffffffffffffffffff8b1660048201526000602482015273deaddeaddeaddeaddeaddeaddeaddeaddead11119063095ea7b3906044016020604051808303816000875af11580156116b8573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906116dc9190612602565b91505b8080156116e95750815b1561175157600083815260cb602052604080822080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660011790555184917f4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c91a261185e565b600083815260cf602052604080822080547fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff001660011790555184917f99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f91a27fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff320161185e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602d60248201527f43726f7373446f6d61696e4d657373656e6765723a206661696c656420746f2060448201527f72656c6179206d6573736167650000000000000000000000000000000000000060648201526084016104df565b505050505b5050505050505050565b73ffffffffffffffffffffffffffffffffffffffff163b151590565b6040517f39fd0090000000000000000000000000000000000000000000000000000000008152734200000000000000000000000000000000000016906339fd00909034906118e190889088908890889060040161263b565b6000604051808303818588803b1580156118fa57600080fd5b505af115801561190e573d6000803e3d6000fd5b505050505050505050565b60608160000361195c57505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b811561198657806119708161268a565b915061197f9050600a836126c2565b9150611960565b60008167ffffffffffffffff8111156119a1576119a16126d6565b6040519080825280601f01601f1916602001820160405280156119cb576020820181803683370190505b5090505b8415611a4e576119e0600183612624565b91506119ed600a86612705565b6119f8906030612719565b60f81b818381518110611a0d57611a0d612731565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a905350611a47600a866126c2565b94506119cf565b949350505050565b6000547501000000000000000000000000000000000000000000900460ff16611b01576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602b60248201527f496e697469616c697a61626c653a20636f6e7472616374206973206e6f74206960448201527f6e697469616c697a696e6700000000000000000000000000000000000000000060648201526084016104df565b60cc80547fffffffffffffffffffffffff00000000000000000000000000000000000000001661dead179055565b6040805173ffffffffffffffffffffffffffffffffffffffff85811660248301528416604482015260648082018490528251808303909101815260849091019091526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167f23b872dd00000000000000000000000000000000000000000000000000000000179052611bc4908590611c9b565b50505050565b6000611bd885858585611dac565b805190602001209050949350505050565b6000611bfa88888888888888611e45565b805190602001209050979650505050505050565b600073ffffffffffffffffffffffffffffffffffffffff8216301480611c5d575073ffffffffffffffffffffffffffffffffffffffff8216734200000000000000000000000000000000000016145b92915050565b600080603f83619c4001026040850201603f5a021015949350505050565b600080600080845160208601878a8af19695505050505050565b6000611cfd826040518060400160405280602081526020017f5361666545524332303a206c6f772d6c6576656c2063616c6c206661696c65648152508573ffffffffffffffffffffffffffffffffffffffff16611ee79092919063ffffffff16565b805190915015611da75780806020019051810190611d1b9190612602565b611da7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602a60248201527f5361666545524332303a204552433230206f7065726174696f6e20646964206e60448201527f6f7420737563636565640000000000000000000000000000000000000000000060648201526084016104df565b505050565b606084848484604051602401611dc59493929190612760565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fcbd4ece9000000000000000000000000000000000000000000000000000000001790529050949350505050565b606087878787878787604051602401611e6497969594939291906127aa565b604080517fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe08184030181529190526020810180517bffffffffffffffffffffffffffffffffffffffffffffffffffffffff167fff8daf15000000000000000000000000000000000000000000000000000000001790529050979650505050505050565b6060610b4384846000858573ffffffffffffffffffffffffffffffffffffffff85163b611f70576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f416464726573733a2063616c6c20746f206e6f6e2d636f6e747261637400000060448201526064016104df565b6000808673ffffffffffffffffffffffffffffffffffffffff168587604051611f999190612808565b60006040518083038185875af1925050503d8060008114611fd6576040519150601f19603f3d011682016040523d82523d6000602084013e611fdb565b606091505b5091509150611feb828286611ff6565b979650505050505050565b60608315612005575081610b46565b8251156120155782518084602001fd5b816040517f08c379a00000000000000000000000000000000000000000000000000000000081526004016104df91906121aa565b803573ffffffffffffffffffffffffffffffffffffffff8116811461206d57600080fd5b919050565b60008083601f84011261208457600080fd5b50813567ffffffffffffffff81111561209c57600080fd5b6020830191508360208285010111156120b457600080fd5b9250929050565b803563ffffffff8116811461206d57600080fd5b600080600080606085870312156120e557600080fd5b6120ee85612049565b9350602085013567ffffffffffffffff81111561210a57600080fd5b61211687828801612072565b90945092506121299050604086016120bb565b905092959194509250565b60005b8381101561214f578181015183820152602001612137565b83811115611bc45750506000910152565b60008151808452612178816020860160208601612134565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169290920160200192915050565b602081526000610b466020830184612160565b6000602082840312156121cf57600080fd5b5035919050565b6000806000604084860312156121eb57600080fd5b833567ffffffffffffffff81111561220257600080fd5b61220e86828701612072565b90945092506122219050602085016120bb565b90509250925092565b60008060008060006080868803121561224257600080fd5b8535945061225260208701612049565b9350604086013567ffffffffffffffff81111561226e57600080fd5b61227a88828901612072565b909450925061228d9050606087016120bb565b90509295509295909350565b60008060008060008060008060e0898b0312156122b557600080fd5b883597506122c560208a01612049565b96506122d360408a01612049565b9550606089013594506080890135935060a0890135925060c089013567ffffffffffffffff81111561230457600080fd5b6123108b828c01612072565b999c989b5096995094979396929594505050565b8183528181602085013750600060208284010152600060207fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0601f840116840101905092915050565b888152600073ffffffffffffffffffffffffffffffffffffffff808a16602084015280891660408401525086606083015260ff8616608083015263ffffffff851660a083015260e060c08301526123c860e083018486612324565b9a9950505050505050505050565b73ffffffffffffffffffffffffffffffffffffffff86168152608060208201526000612406608083018688612324565b905083604083015263ffffffff831660608301529695505050505050565b60008451612436818460208901612134565b80830190507f2e000000000000000000000000000000000000000000000000000000000000008082528551612472816001850160208a01612134565b6001920191820152835161248d816002840160208801612134565b0160020195945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b600067ffffffffffffffff808316818516818304811182151516156124f0576124f061249a565b02949350505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600067ffffffffffffffff80841680612543576125436124f9565b92169190910492915050565b600067ffffffffffffffff8083168185168083038211156125725761257261249a565b01949350505050565b888152600073ffffffffffffffffffffffffffffffffffffffff808a16602084015280891660408401525086606083015285608083015263ffffffff851660a083015260e060c08301526123c860e083018486612324565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052600160045260246000fd5b60006020828403121561261457600080fd5b81518015158114610b4657600080fd5b6000828210156126365761263661249a565b500390565b84815273ffffffffffffffffffffffffffffffffffffffff8416602082015267ffffffffffffffff831660408201526080606082015260006126806080830184612160565b9695505050505050565b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff82036126bb576126bb61249a565b5060010190565b6000826126d1576126d16124f9565b500490565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600082612714576127146124f9565b500690565b6000821982111561272c5761272c61249a565b500190565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b600073ffffffffffffffffffffffffffffffffffffffff8087168352808616602084015250608060408301526127996080830185612160565b905082606083015295945050505050565b878152600073ffffffffffffffffffffffffffffffffffffffff80891660208401528088166040840152508560608301528460808301528360a083015260e060c08301526127fb60e0830184612160565b9998505050505050505050565b6000825161281a818460208701612134565b919091019291505056fea164736f6c634300080f000a",
}

// L2CrossDomainMessengerABI is the input ABI used to generate the binding from.
// Deprecated: Use L2CrossDomainMessengerMetaData.ABI instead.
var L2CrossDomainMessengerABI = L2CrossDomainMessengerMetaData.ABI

// L2CrossDomainMessengerBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use L2CrossDomainMessengerMetaData.Bin instead.
var L2CrossDomainMessengerBin = L2CrossDomainMessengerMetaData.Bin

// DeployL2CrossDomainMessenger deploys a new Ethereum contract, binding an instance of L2CrossDomainMessenger to it.
func DeployL2CrossDomainMessenger(auth *bind.TransactOpts, backend bind.ContractBackend, _l1CrossDomainMessenger common.Address) (common.Address, *types.Transaction, *L2CrossDomainMessenger, error) {
	parsed, err := L2CrossDomainMessengerMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(L2CrossDomainMessengerBin), backend, _l1CrossDomainMessenger)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &L2CrossDomainMessenger{L2CrossDomainMessengerCaller: L2CrossDomainMessengerCaller{contract: contract}, L2CrossDomainMessengerTransactor: L2CrossDomainMessengerTransactor{contract: contract}, L2CrossDomainMessengerFilterer: L2CrossDomainMessengerFilterer{contract: contract}}, nil
}

// L2CrossDomainMessenger is an auto generated Go binding around an Ethereum contract.
type L2CrossDomainMessenger struct {
	L2CrossDomainMessengerCaller     // Read-only binding to the contract
	L2CrossDomainMessengerTransactor // Write-only binding to the contract
	L2CrossDomainMessengerFilterer   // Log filterer for contract events
}

// L2CrossDomainMessengerCaller is an auto generated read-only Go binding around an Ethereum contract.
type L2CrossDomainMessengerCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L2CrossDomainMessengerTransactor is an auto generated write-only Go binding around an Ethereum contract.
type L2CrossDomainMessengerTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L2CrossDomainMessengerFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type L2CrossDomainMessengerFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// L2CrossDomainMessengerSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type L2CrossDomainMessengerSession struct {
	Contract     *L2CrossDomainMessenger // Generic contract binding to set the session for
	CallOpts     bind.CallOpts           // Call options to use throughout this session
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// L2CrossDomainMessengerCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type L2CrossDomainMessengerCallerSession struct {
	Contract *L2CrossDomainMessengerCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                 // Call options to use throughout this session
}

// L2CrossDomainMessengerTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type L2CrossDomainMessengerTransactorSession struct {
	Contract     *L2CrossDomainMessengerTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                 // Transaction auth options to use throughout this session
}

// L2CrossDomainMessengerRaw is an auto generated low-level Go binding around an Ethereum contract.
type L2CrossDomainMessengerRaw struct {
	Contract *L2CrossDomainMessenger // Generic contract binding to access the raw methods on
}

// L2CrossDomainMessengerCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type L2CrossDomainMessengerCallerRaw struct {
	Contract *L2CrossDomainMessengerCaller // Generic read-only contract binding to access the raw methods on
}

// L2CrossDomainMessengerTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type L2CrossDomainMessengerTransactorRaw struct {
	Contract *L2CrossDomainMessengerTransactor // Generic write-only contract binding to access the raw methods on
}

// NewL2CrossDomainMessenger creates a new instance of L2CrossDomainMessenger, bound to a specific deployed contract.
func NewL2CrossDomainMessenger(address common.Address, backend bind.ContractBackend) (*L2CrossDomainMessenger, error) {
	contract, err := bindL2CrossDomainMessenger(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessenger{L2CrossDomainMessengerCaller: L2CrossDomainMessengerCaller{contract: contract}, L2CrossDomainMessengerTransactor: L2CrossDomainMessengerTransactor{contract: contract}, L2CrossDomainMessengerFilterer: L2CrossDomainMessengerFilterer{contract: contract}}, nil
}

// NewL2CrossDomainMessengerCaller creates a new read-only instance of L2CrossDomainMessenger, bound to a specific deployed contract.
func NewL2CrossDomainMessengerCaller(address common.Address, caller bind.ContractCaller) (*L2CrossDomainMessengerCaller, error) {
	contract, err := bindL2CrossDomainMessenger(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerCaller{contract: contract}, nil
}

// NewL2CrossDomainMessengerTransactor creates a new write-only instance of L2CrossDomainMessenger, bound to a specific deployed contract.
func NewL2CrossDomainMessengerTransactor(address common.Address, transactor bind.ContractTransactor) (*L2CrossDomainMessengerTransactor, error) {
	contract, err := bindL2CrossDomainMessenger(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerTransactor{contract: contract}, nil
}

// NewL2CrossDomainMessengerFilterer creates a new log filterer instance of L2CrossDomainMessenger, bound to a specific deployed contract.
func NewL2CrossDomainMessengerFilterer(address common.Address, filterer bind.ContractFilterer) (*L2CrossDomainMessengerFilterer, error) {
	contract, err := bindL2CrossDomainMessenger(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerFilterer{contract: contract}, nil
}

// bindL2CrossDomainMessenger binds a generic wrapper to an already deployed contract.
func bindL2CrossDomainMessenger(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := L2CrossDomainMessengerMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L2CrossDomainMessenger *L2CrossDomainMessengerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L2CrossDomainMessenger.Contract.L2CrossDomainMessengerCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L2CrossDomainMessenger *L2CrossDomainMessengerRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.L2CrossDomainMessengerTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L2CrossDomainMessenger *L2CrossDomainMessengerRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.L2CrossDomainMessengerTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _L2CrossDomainMessenger.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.contract.Transact(opts, method, params...)
}

// MESSAGEVERSION is a free data retrieval call binding the contract method 0x3f827a5a.
//
// Solidity: function MESSAGE_VERSION() view returns(uint16)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) MESSAGEVERSION(opts *bind.CallOpts) (uint16, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "MESSAGE_VERSION")

	if err != nil {
		return *new(uint16), err
	}

	out0 := *abi.ConvertType(out[0], new(uint16)).(*uint16)

	return out0, err

}

// MESSAGEVERSION is a free data retrieval call binding the contract method 0x3f827a5a.
//
// Solidity: function MESSAGE_VERSION() view returns(uint16)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) MESSAGEVERSION() (uint16, error) {
	return _L2CrossDomainMessenger.Contract.MESSAGEVERSION(&_L2CrossDomainMessenger.CallOpts)
}

// MESSAGEVERSION is a free data retrieval call binding the contract method 0x3f827a5a.
//
// Solidity: function MESSAGE_VERSION() view returns(uint16)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) MESSAGEVERSION() (uint16, error) {
	return _L2CrossDomainMessenger.Contract.MESSAGEVERSION(&_L2CrossDomainMessenger.CallOpts)
}

// MINGASCALLDATAOVERHEAD is a free data retrieval call binding the contract method 0x028f85f7.
//
// Solidity: function MIN_GAS_CALLDATA_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) MINGASCALLDATAOVERHEAD(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "MIN_GAS_CALLDATA_OVERHEAD")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MINGASCALLDATAOVERHEAD is a free data retrieval call binding the contract method 0x028f85f7.
//
// Solidity: function MIN_GAS_CALLDATA_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) MINGASCALLDATAOVERHEAD() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.MINGASCALLDATAOVERHEAD(&_L2CrossDomainMessenger.CallOpts)
}

// MINGASCALLDATAOVERHEAD is a free data retrieval call binding the contract method 0x028f85f7.
//
// Solidity: function MIN_GAS_CALLDATA_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) MINGASCALLDATAOVERHEAD() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.MINGASCALLDATAOVERHEAD(&_L2CrossDomainMessenger.CallOpts)
}

// MINGASDYNAMICOVERHEADDENOMINATOR is a free data retrieval call binding the contract method 0x0c568498.
//
// Solidity: function MIN_GAS_DYNAMIC_OVERHEAD_DENOMINATOR() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) MINGASDYNAMICOVERHEADDENOMINATOR(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "MIN_GAS_DYNAMIC_OVERHEAD_DENOMINATOR")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MINGASDYNAMICOVERHEADDENOMINATOR is a free data retrieval call binding the contract method 0x0c568498.
//
// Solidity: function MIN_GAS_DYNAMIC_OVERHEAD_DENOMINATOR() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) MINGASDYNAMICOVERHEADDENOMINATOR() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.MINGASDYNAMICOVERHEADDENOMINATOR(&_L2CrossDomainMessenger.CallOpts)
}

// MINGASDYNAMICOVERHEADDENOMINATOR is a free data retrieval call binding the contract method 0x0c568498.
//
// Solidity: function MIN_GAS_DYNAMIC_OVERHEAD_DENOMINATOR() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) MINGASDYNAMICOVERHEADDENOMINATOR() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.MINGASDYNAMICOVERHEADDENOMINATOR(&_L2CrossDomainMessenger.CallOpts)
}

// MINGASDYNAMICOVERHEADNUMERATOR is a free data retrieval call binding the contract method 0x2828d7e8.
//
// Solidity: function MIN_GAS_DYNAMIC_OVERHEAD_NUMERATOR() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) MINGASDYNAMICOVERHEADNUMERATOR(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "MIN_GAS_DYNAMIC_OVERHEAD_NUMERATOR")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// MINGASDYNAMICOVERHEADNUMERATOR is a free data retrieval call binding the contract method 0x2828d7e8.
//
// Solidity: function MIN_GAS_DYNAMIC_OVERHEAD_NUMERATOR() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) MINGASDYNAMICOVERHEADNUMERATOR() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.MINGASDYNAMICOVERHEADNUMERATOR(&_L2CrossDomainMessenger.CallOpts)
}

// MINGASDYNAMICOVERHEADNUMERATOR is a free data retrieval call binding the contract method 0x2828d7e8.
//
// Solidity: function MIN_GAS_DYNAMIC_OVERHEAD_NUMERATOR() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) MINGASDYNAMICOVERHEADNUMERATOR() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.MINGASDYNAMICOVERHEADNUMERATOR(&_L2CrossDomainMessenger.CallOpts)
}

// OTHERMESSENGER is a free data retrieval call binding the contract method 0x9fce812c.
//
// Solidity: function OTHER_MESSENGER() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) OTHERMESSENGER(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "OTHER_MESSENGER")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// OTHERMESSENGER is a free data retrieval call binding the contract method 0x9fce812c.
//
// Solidity: function OTHER_MESSENGER() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) OTHERMESSENGER() (common.Address, error) {
	return _L2CrossDomainMessenger.Contract.OTHERMESSENGER(&_L2CrossDomainMessenger.CallOpts)
}

// OTHERMESSENGER is a free data retrieval call binding the contract method 0x9fce812c.
//
// Solidity: function OTHER_MESSENGER() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) OTHERMESSENGER() (common.Address, error) {
	return _L2CrossDomainMessenger.Contract.OTHERMESSENGER(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYCALLOVERHEAD is a free data retrieval call binding the contract method 0x4c1d6a69.
//
// Solidity: function RELAY_CALL_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) RELAYCALLOVERHEAD(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "RELAY_CALL_OVERHEAD")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// RELAYCALLOVERHEAD is a free data retrieval call binding the contract method 0x4c1d6a69.
//
// Solidity: function RELAY_CALL_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) RELAYCALLOVERHEAD() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYCALLOVERHEAD(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYCALLOVERHEAD is a free data retrieval call binding the contract method 0x4c1d6a69.
//
// Solidity: function RELAY_CALL_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) RELAYCALLOVERHEAD() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYCALLOVERHEAD(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYCONSTANTOVERHEAD is a free data retrieval call binding the contract method 0x83a74074.
//
// Solidity: function RELAY_CONSTANT_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) RELAYCONSTANTOVERHEAD(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "RELAY_CONSTANT_OVERHEAD")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// RELAYCONSTANTOVERHEAD is a free data retrieval call binding the contract method 0x83a74074.
//
// Solidity: function RELAY_CONSTANT_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) RELAYCONSTANTOVERHEAD() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYCONSTANTOVERHEAD(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYCONSTANTOVERHEAD is a free data retrieval call binding the contract method 0x83a74074.
//
// Solidity: function RELAY_CONSTANT_OVERHEAD() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) RELAYCONSTANTOVERHEAD() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYCONSTANTOVERHEAD(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYGASCHECKBUFFER is a free data retrieval call binding the contract method 0x5644cfdf.
//
// Solidity: function RELAY_GAS_CHECK_BUFFER() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) RELAYGASCHECKBUFFER(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "RELAY_GAS_CHECK_BUFFER")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// RELAYGASCHECKBUFFER is a free data retrieval call binding the contract method 0x5644cfdf.
//
// Solidity: function RELAY_GAS_CHECK_BUFFER() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) RELAYGASCHECKBUFFER() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYGASCHECKBUFFER(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYGASCHECKBUFFER is a free data retrieval call binding the contract method 0x5644cfdf.
//
// Solidity: function RELAY_GAS_CHECK_BUFFER() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) RELAYGASCHECKBUFFER() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYGASCHECKBUFFER(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYRESERVEDGAS is a free data retrieval call binding the contract method 0x8cbeeef2.
//
// Solidity: function RELAY_RESERVED_GAS() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) RELAYRESERVEDGAS(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "RELAY_RESERVED_GAS")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// RELAYRESERVEDGAS is a free data retrieval call binding the contract method 0x8cbeeef2.
//
// Solidity: function RELAY_RESERVED_GAS() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) RELAYRESERVEDGAS() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYRESERVEDGAS(&_L2CrossDomainMessenger.CallOpts)
}

// RELAYRESERVEDGAS is a free data retrieval call binding the contract method 0x8cbeeef2.
//
// Solidity: function RELAY_RESERVED_GAS() view returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) RELAYRESERVEDGAS() (uint64, error) {
	return _L2CrossDomainMessenger.Contract.RELAYRESERVEDGAS(&_L2CrossDomainMessenger.CallOpts)
}

// BaseGas is a free data retrieval call binding the contract method 0xb28ade25.
//
// Solidity: function baseGas(bytes _message, uint32 _minGasLimit) pure returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) BaseGas(opts *bind.CallOpts, _message []byte, _minGasLimit uint32) (uint64, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "baseGas", _message, _minGasLimit)

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// BaseGas is a free data retrieval call binding the contract method 0xb28ade25.
//
// Solidity: function baseGas(bytes _message, uint32 _minGasLimit) pure returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) BaseGas(_message []byte, _minGasLimit uint32) (uint64, error) {
	return _L2CrossDomainMessenger.Contract.BaseGas(&_L2CrossDomainMessenger.CallOpts, _message, _minGasLimit)
}

// BaseGas is a free data retrieval call binding the contract method 0xb28ade25.
//
// Solidity: function baseGas(bytes _message, uint32 _minGasLimit) pure returns(uint64)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) BaseGas(_message []byte, _minGasLimit uint32) (uint64, error) {
	return _L2CrossDomainMessenger.Contract.BaseGas(&_L2CrossDomainMessenger.CallOpts, _message, _minGasLimit)
}

// FailedMessages is a free data retrieval call binding the contract method 0xa4e7f8bd.
//
// Solidity: function failedMessages(bytes32 ) view returns(bool)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) FailedMessages(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "failedMessages", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// FailedMessages is a free data retrieval call binding the contract method 0xa4e7f8bd.
//
// Solidity: function failedMessages(bytes32 ) view returns(bool)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) FailedMessages(arg0 [32]byte) (bool, error) {
	return _L2CrossDomainMessenger.Contract.FailedMessages(&_L2CrossDomainMessenger.CallOpts, arg0)
}

// FailedMessages is a free data retrieval call binding the contract method 0xa4e7f8bd.
//
// Solidity: function failedMessages(bytes32 ) view returns(bool)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) FailedMessages(arg0 [32]byte) (bool, error) {
	return _L2CrossDomainMessenger.Contract.FailedMessages(&_L2CrossDomainMessenger.CallOpts, arg0)
}

// L1CrossDomainMessenger is a free data retrieval call binding the contract method 0xa7119869.
//
// Solidity: function l1CrossDomainMessenger() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) L1CrossDomainMessenger(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "l1CrossDomainMessenger")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// L1CrossDomainMessenger is a free data retrieval call binding the contract method 0xa7119869.
//
// Solidity: function l1CrossDomainMessenger() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) L1CrossDomainMessenger() (common.Address, error) {
	return _L2CrossDomainMessenger.Contract.L1CrossDomainMessenger(&_L2CrossDomainMessenger.CallOpts)
}

// L1CrossDomainMessenger is a free data retrieval call binding the contract method 0xa7119869.
//
// Solidity: function l1CrossDomainMessenger() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) L1CrossDomainMessenger() (common.Address, error) {
	return _L2CrossDomainMessenger.Contract.L1CrossDomainMessenger(&_L2CrossDomainMessenger.CallOpts)
}

// MessageNonce is a free data retrieval call binding the contract method 0xecc70428.
//
// Solidity: function messageNonce() view returns(uint256)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) MessageNonce(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "messageNonce")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// MessageNonce is a free data retrieval call binding the contract method 0xecc70428.
//
// Solidity: function messageNonce() view returns(uint256)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) MessageNonce() (*big.Int, error) {
	return _L2CrossDomainMessenger.Contract.MessageNonce(&_L2CrossDomainMessenger.CallOpts)
}

// MessageNonce is a free data retrieval call binding the contract method 0xecc70428.
//
// Solidity: function messageNonce() view returns(uint256)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) MessageNonce() (*big.Int, error) {
	return _L2CrossDomainMessenger.Contract.MessageNonce(&_L2CrossDomainMessenger.CallOpts)
}

// SuccessfulMessages is a free data retrieval call binding the contract method 0xb1b1b209.
//
// Solidity: function successfulMessages(bytes32 ) view returns(bool)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) SuccessfulMessages(opts *bind.CallOpts, arg0 [32]byte) (bool, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "successfulMessages", arg0)

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// SuccessfulMessages is a free data retrieval call binding the contract method 0xb1b1b209.
//
// Solidity: function successfulMessages(bytes32 ) view returns(bool)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) SuccessfulMessages(arg0 [32]byte) (bool, error) {
	return _L2CrossDomainMessenger.Contract.SuccessfulMessages(&_L2CrossDomainMessenger.CallOpts, arg0)
}

// SuccessfulMessages is a free data retrieval call binding the contract method 0xb1b1b209.
//
// Solidity: function successfulMessages(bytes32 ) view returns(bool)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) SuccessfulMessages(arg0 [32]byte) (bool, error) {
	return _L2CrossDomainMessenger.Contract.SuccessfulMessages(&_L2CrossDomainMessenger.CallOpts, arg0)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) Version(opts *bind.CallOpts) (string, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "version")

	if err != nil {
		return *new(string), err
	}

	out0 := *abi.ConvertType(out[0], new(string)).(*string)

	return out0, err

}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) Version() (string, error) {
	return _L2CrossDomainMessenger.Contract.Version(&_L2CrossDomainMessenger.CallOpts)
}

// Version is a free data retrieval call binding the contract method 0x54fd4d50.
//
// Solidity: function version() view returns(string)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) Version() (string, error) {
	return _L2CrossDomainMessenger.Contract.Version(&_L2CrossDomainMessenger.CallOpts)
}

// XDomainMessageSender is a free data retrieval call binding the contract method 0x6e296e45.
//
// Solidity: function xDomainMessageSender() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCaller) XDomainMessageSender(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _L2CrossDomainMessenger.contract.Call(opts, &out, "xDomainMessageSender")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// XDomainMessageSender is a free data retrieval call binding the contract method 0x6e296e45.
//
// Solidity: function xDomainMessageSender() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) XDomainMessageSender() (common.Address, error) {
	return _L2CrossDomainMessenger.Contract.XDomainMessageSender(&_L2CrossDomainMessenger.CallOpts)
}

// XDomainMessageSender is a free data retrieval call binding the contract method 0x6e296e45.
//
// Solidity: function xDomainMessageSender() view returns(address)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerCallerSession) XDomainMessageSender() (common.Address, error) {
	return _L2CrossDomainMessenger.Contract.XDomainMessageSender(&_L2CrossDomainMessenger.CallOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactor) Initialize(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.contract.Transact(opts, "initialize")
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) Initialize() (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.Initialize(&_L2CrossDomainMessenger.TransactOpts)
}

// Initialize is a paid mutator transaction binding the contract method 0x8129fc1c.
//
// Solidity: function initialize() returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactorSession) Initialize() (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.Initialize(&_L2CrossDomainMessenger.TransactOpts)
}

// RelayMessage is a paid mutator transaction binding the contract method 0xff8daf15.
//
// Solidity: function relayMessage(uint256 _nonce, address _sender, address _target, uint256 _mntValue, uint256 _ethValue, uint256 _minGasLimit, bytes _message) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactor) RelayMessage(opts *bind.TransactOpts, _nonce *big.Int, _sender common.Address, _target common.Address, _mntValue *big.Int, _ethValue *big.Int, _minGasLimit *big.Int, _message []byte) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.contract.Transact(opts, "relayMessage", _nonce, _sender, _target, _mntValue, _ethValue, _minGasLimit, _message)
}

// RelayMessage is a paid mutator transaction binding the contract method 0xff8daf15.
//
// Solidity: function relayMessage(uint256 _nonce, address _sender, address _target, uint256 _mntValue, uint256 _ethValue, uint256 _minGasLimit, bytes _message) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) RelayMessage(_nonce *big.Int, _sender common.Address, _target common.Address, _mntValue *big.Int, _ethValue *big.Int, _minGasLimit *big.Int, _message []byte) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.RelayMessage(&_L2CrossDomainMessenger.TransactOpts, _nonce, _sender, _target, _mntValue, _ethValue, _minGasLimit, _message)
}

// RelayMessage is a paid mutator transaction binding the contract method 0xff8daf15.
//
// Solidity: function relayMessage(uint256 _nonce, address _sender, address _target, uint256 _mntValue, uint256 _ethValue, uint256 _minGasLimit, bytes _message) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactorSession) RelayMessage(_nonce *big.Int, _sender common.Address, _target common.Address, _mntValue *big.Int, _ethValue *big.Int, _minGasLimit *big.Int, _message []byte) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.RelayMessage(&_L2CrossDomainMessenger.TransactOpts, _nonce, _sender, _target, _mntValue, _ethValue, _minGasLimit, _message)
}

// SendMessage is a paid mutator transaction binding the contract method 0x3dbb202b.
//
// Solidity: function sendMessage(address _target, bytes _message, uint32 _minGasLimit) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactor) SendMessage(opts *bind.TransactOpts, _target common.Address, _message []byte, _minGasLimit uint32) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.contract.Transact(opts, "sendMessage", _target, _message, _minGasLimit)
}

// SendMessage is a paid mutator transaction binding the contract method 0x3dbb202b.
//
// Solidity: function sendMessage(address _target, bytes _message, uint32 _minGasLimit) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) SendMessage(_target common.Address, _message []byte, _minGasLimit uint32) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.SendMessage(&_L2CrossDomainMessenger.TransactOpts, _target, _message, _minGasLimit)
}

// SendMessage is a paid mutator transaction binding the contract method 0x3dbb202b.
//
// Solidity: function sendMessage(address _target, bytes _message, uint32 _minGasLimit) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactorSession) SendMessage(_target common.Address, _message []byte, _minGasLimit uint32) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.SendMessage(&_L2CrossDomainMessenger.TransactOpts, _target, _message, _minGasLimit)
}

// SendMessage0 is a paid mutator transaction binding the contract method 0xb575e300.
//
// Solidity: function sendMessage(uint256 _ethAmount, address _target, bytes _message, uint32 _minGasLimit) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactor) SendMessage0(opts *bind.TransactOpts, _ethAmount *big.Int, _target common.Address, _message []byte, _minGasLimit uint32) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.contract.Transact(opts, "sendMessage0", _ethAmount, _target, _message, _minGasLimit)
}

// SendMessage0 is a paid mutator transaction binding the contract method 0xb575e300.
//
// Solidity: function sendMessage(uint256 _ethAmount, address _target, bytes _message, uint32 _minGasLimit) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerSession) SendMessage0(_ethAmount *big.Int, _target common.Address, _message []byte, _minGasLimit uint32) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.SendMessage0(&_L2CrossDomainMessenger.TransactOpts, _ethAmount, _target, _message, _minGasLimit)
}

// SendMessage0 is a paid mutator transaction binding the contract method 0xb575e300.
//
// Solidity: function sendMessage(uint256 _ethAmount, address _target, bytes _message, uint32 _minGasLimit) payable returns()
func (_L2CrossDomainMessenger *L2CrossDomainMessengerTransactorSession) SendMessage0(_ethAmount *big.Int, _target common.Address, _message []byte, _minGasLimit uint32) (*types.Transaction, error) {
	return _L2CrossDomainMessenger.Contract.SendMessage0(&_L2CrossDomainMessenger.TransactOpts, _ethAmount, _target, _message, _minGasLimit)
}

// L2CrossDomainMessengerFailedRelayedMessageIterator is returned from FilterFailedRelayedMessage and is used to iterate over the raw logs and unpacked data for FailedRelayedMessage events raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerFailedRelayedMessageIterator struct {
	Event *L2CrossDomainMessengerFailedRelayedMessage // Event containing the contract specifics and raw log

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
func (it *L2CrossDomainMessengerFailedRelayedMessageIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L2CrossDomainMessengerFailedRelayedMessage)
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
		it.Event = new(L2CrossDomainMessengerFailedRelayedMessage)
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
func (it *L2CrossDomainMessengerFailedRelayedMessageIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L2CrossDomainMessengerFailedRelayedMessageIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L2CrossDomainMessengerFailedRelayedMessage represents a FailedRelayedMessage event raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerFailedRelayedMessage struct {
	MsgHash [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterFailedRelayedMessage is a free log retrieval operation binding the contract event 0x99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f.
//
// Solidity: event FailedRelayedMessage(bytes32 indexed msgHash)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) FilterFailedRelayedMessage(opts *bind.FilterOpts, msgHash [][32]byte) (*L2CrossDomainMessengerFailedRelayedMessageIterator, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.FilterLogs(opts, "FailedRelayedMessage", msgHashRule)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerFailedRelayedMessageIterator{contract: _L2CrossDomainMessenger.contract, event: "FailedRelayedMessage", logs: logs, sub: sub}, nil
}

// WatchFailedRelayedMessage is a free log subscription operation binding the contract event 0x99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f.
//
// Solidity: event FailedRelayedMessage(bytes32 indexed msgHash)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) WatchFailedRelayedMessage(opts *bind.WatchOpts, sink chan<- *L2CrossDomainMessengerFailedRelayedMessage, msgHash [][32]byte) (event.Subscription, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.WatchLogs(opts, "FailedRelayedMessage", msgHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L2CrossDomainMessengerFailedRelayedMessage)
				if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "FailedRelayedMessage", log); err != nil {
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

// ParseFailedRelayedMessage is a log parse operation binding the contract event 0x99d0e048484baa1b1540b1367cb128acd7ab2946d1ed91ec10e3c85e4bf51b8f.
//
// Solidity: event FailedRelayedMessage(bytes32 indexed msgHash)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) ParseFailedRelayedMessage(log types.Log) (*L2CrossDomainMessengerFailedRelayedMessage, error) {
	event := new(L2CrossDomainMessengerFailedRelayedMessage)
	if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "FailedRelayedMessage", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L2CrossDomainMessengerInitializedIterator is returned from FilterInitialized and is used to iterate over the raw logs and unpacked data for Initialized events raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerInitializedIterator struct {
	Event *L2CrossDomainMessengerInitialized // Event containing the contract specifics and raw log

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
func (it *L2CrossDomainMessengerInitializedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L2CrossDomainMessengerInitialized)
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
		it.Event = new(L2CrossDomainMessengerInitialized)
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
func (it *L2CrossDomainMessengerInitializedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L2CrossDomainMessengerInitializedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L2CrossDomainMessengerInitialized represents a Initialized event raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerInitialized struct {
	Version uint8
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterInitialized is a free log retrieval operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) FilterInitialized(opts *bind.FilterOpts) (*L2CrossDomainMessengerInitializedIterator, error) {

	logs, sub, err := _L2CrossDomainMessenger.contract.FilterLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerInitializedIterator{contract: _L2CrossDomainMessenger.contract, event: "Initialized", logs: logs, sub: sub}, nil
}

// WatchInitialized is a free log subscription operation binding the contract event 0x7f26b83ff96e1f2b6a682f133852f6798a09c465da95921460cefb3847402498.
//
// Solidity: event Initialized(uint8 version)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) WatchInitialized(opts *bind.WatchOpts, sink chan<- *L2CrossDomainMessengerInitialized) (event.Subscription, error) {

	logs, sub, err := _L2CrossDomainMessenger.contract.WatchLogs(opts, "Initialized")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L2CrossDomainMessengerInitialized)
				if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "Initialized", log); err != nil {
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
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) ParseInitialized(log types.Log) (*L2CrossDomainMessengerInitialized, error) {
	event := new(L2CrossDomainMessengerInitialized)
	if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "Initialized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L2CrossDomainMessengerRelayedMessageIterator is returned from FilterRelayedMessage and is used to iterate over the raw logs and unpacked data for RelayedMessage events raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerRelayedMessageIterator struct {
	Event *L2CrossDomainMessengerRelayedMessage // Event containing the contract specifics and raw log

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
func (it *L2CrossDomainMessengerRelayedMessageIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L2CrossDomainMessengerRelayedMessage)
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
		it.Event = new(L2CrossDomainMessengerRelayedMessage)
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
func (it *L2CrossDomainMessengerRelayedMessageIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L2CrossDomainMessengerRelayedMessageIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L2CrossDomainMessengerRelayedMessage represents a RelayedMessage event raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerRelayedMessage struct {
	MsgHash [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterRelayedMessage is a free log retrieval operation binding the contract event 0x4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c.
//
// Solidity: event RelayedMessage(bytes32 indexed msgHash)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) FilterRelayedMessage(opts *bind.FilterOpts, msgHash [][32]byte) (*L2CrossDomainMessengerRelayedMessageIterator, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.FilterLogs(opts, "RelayedMessage", msgHashRule)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerRelayedMessageIterator{contract: _L2CrossDomainMessenger.contract, event: "RelayedMessage", logs: logs, sub: sub}, nil
}

// WatchRelayedMessage is a free log subscription operation binding the contract event 0x4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c.
//
// Solidity: event RelayedMessage(bytes32 indexed msgHash)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) WatchRelayedMessage(opts *bind.WatchOpts, sink chan<- *L2CrossDomainMessengerRelayedMessage, msgHash [][32]byte) (event.Subscription, error) {

	var msgHashRule []interface{}
	for _, msgHashItem := range msgHash {
		msgHashRule = append(msgHashRule, msgHashItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.WatchLogs(opts, "RelayedMessage", msgHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L2CrossDomainMessengerRelayedMessage)
				if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "RelayedMessage", log); err != nil {
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

// ParseRelayedMessage is a log parse operation binding the contract event 0x4641df4a962071e12719d8c8c8e5ac7fc4d97b927346a3d7a335b1f7517e133c.
//
// Solidity: event RelayedMessage(bytes32 indexed msgHash)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) ParseRelayedMessage(log types.Log) (*L2CrossDomainMessengerRelayedMessage, error) {
	event := new(L2CrossDomainMessengerRelayedMessage)
	if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "RelayedMessage", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L2CrossDomainMessengerSentMessageIterator is returned from FilterSentMessage and is used to iterate over the raw logs and unpacked data for SentMessage events raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerSentMessageIterator struct {
	Event *L2CrossDomainMessengerSentMessage // Event containing the contract specifics and raw log

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
func (it *L2CrossDomainMessengerSentMessageIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L2CrossDomainMessengerSentMessage)
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
		it.Event = new(L2CrossDomainMessengerSentMessage)
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
func (it *L2CrossDomainMessengerSentMessageIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L2CrossDomainMessengerSentMessageIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L2CrossDomainMessengerSentMessage represents a SentMessage event raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerSentMessage struct {
	Target       common.Address
	Sender       common.Address
	Message      []byte
	MessageNonce *big.Int
	GasLimit     *big.Int
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterSentMessage is a free log retrieval operation binding the contract event 0xcb0f7ffd78f9aee47a248fae8db181db6eee833039123e026dcbff529522e52a.
//
// Solidity: event SentMessage(address indexed target, address sender, bytes message, uint256 messageNonce, uint256 gasLimit)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) FilterSentMessage(opts *bind.FilterOpts, target []common.Address) (*L2CrossDomainMessengerSentMessageIterator, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.FilterLogs(opts, "SentMessage", targetRule)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerSentMessageIterator{contract: _L2CrossDomainMessenger.contract, event: "SentMessage", logs: logs, sub: sub}, nil
}

// WatchSentMessage is a free log subscription operation binding the contract event 0xcb0f7ffd78f9aee47a248fae8db181db6eee833039123e026dcbff529522e52a.
//
// Solidity: event SentMessage(address indexed target, address sender, bytes message, uint256 messageNonce, uint256 gasLimit)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) WatchSentMessage(opts *bind.WatchOpts, sink chan<- *L2CrossDomainMessengerSentMessage, target []common.Address) (event.Subscription, error) {

	var targetRule []interface{}
	for _, targetItem := range target {
		targetRule = append(targetRule, targetItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.WatchLogs(opts, "SentMessage", targetRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L2CrossDomainMessengerSentMessage)
				if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "SentMessage", log); err != nil {
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

// ParseSentMessage is a log parse operation binding the contract event 0xcb0f7ffd78f9aee47a248fae8db181db6eee833039123e026dcbff529522e52a.
//
// Solidity: event SentMessage(address indexed target, address sender, bytes message, uint256 messageNonce, uint256 gasLimit)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) ParseSentMessage(log types.Log) (*L2CrossDomainMessengerSentMessage, error) {
	event := new(L2CrossDomainMessengerSentMessage)
	if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "SentMessage", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// L2CrossDomainMessengerSentMessageExtension1Iterator is returned from FilterSentMessageExtension1 and is used to iterate over the raw logs and unpacked data for SentMessageExtension1 events raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerSentMessageExtension1Iterator struct {
	Event *L2CrossDomainMessengerSentMessageExtension1 // Event containing the contract specifics and raw log

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
func (it *L2CrossDomainMessengerSentMessageExtension1Iterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(L2CrossDomainMessengerSentMessageExtension1)
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
		it.Event = new(L2CrossDomainMessengerSentMessageExtension1)
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
func (it *L2CrossDomainMessengerSentMessageExtension1Iterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *L2CrossDomainMessengerSentMessageExtension1Iterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// L2CrossDomainMessengerSentMessageExtension1 represents a SentMessageExtension1 event raised by the L2CrossDomainMessenger contract.
type L2CrossDomainMessengerSentMessageExtension1 struct {
	Sender   common.Address
	MntValue *big.Int
	EthValue *big.Int
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterSentMessageExtension1 is a free log retrieval operation binding the contract event 0xcf00802ba1f8c659140235227979ca08afaba336a9f9fdc4a5107ed9e8013d08.
//
// Solidity: event SentMessageExtension1(address indexed sender, uint256 mntValue, uint256 ethValue)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) FilterSentMessageExtension1(opts *bind.FilterOpts, sender []common.Address) (*L2CrossDomainMessengerSentMessageExtension1Iterator, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.FilterLogs(opts, "SentMessageExtension1", senderRule)
	if err != nil {
		return nil, err
	}
	return &L2CrossDomainMessengerSentMessageExtension1Iterator{contract: _L2CrossDomainMessenger.contract, event: "SentMessageExtension1", logs: logs, sub: sub}, nil
}

// WatchSentMessageExtension1 is a free log subscription operation binding the contract event 0xcf00802ba1f8c659140235227979ca08afaba336a9f9fdc4a5107ed9e8013d08.
//
// Solidity: event SentMessageExtension1(address indexed sender, uint256 mntValue, uint256 ethValue)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) WatchSentMessageExtension1(opts *bind.WatchOpts, sink chan<- *L2CrossDomainMessengerSentMessageExtension1, sender []common.Address) (event.Subscription, error) {

	var senderRule []interface{}
	for _, senderItem := range sender {
		senderRule = append(senderRule, senderItem)
	}

	logs, sub, err := _L2CrossDomainMessenger.contract.WatchLogs(opts, "SentMessageExtension1", senderRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(L2CrossDomainMessengerSentMessageExtension1)
				if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "SentMessageExtension1", log); err != nil {
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

// ParseSentMessageExtension1 is a log parse operation binding the contract event 0xcf00802ba1f8c659140235227979ca08afaba336a9f9fdc4a5107ed9e8013d08.
//
// Solidity: event SentMessageExtension1(address indexed sender, uint256 mntValue, uint256 ethValue)
func (_L2CrossDomainMessenger *L2CrossDomainMessengerFilterer) ParseSentMessageExtension1(log types.Log) (*L2CrossDomainMessengerSentMessageExtension1, error) {
	event := new(L2CrossDomainMessengerSentMessageExtension1)
	if err := _L2CrossDomainMessenger.contract.UnpackLog(event, "SentMessageExtension1", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
