// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package mantlebindings

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
	ABI: "[{\"type\":\"constructor\",\"inputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DECIMALS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"baseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"baseFeeScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"blobBaseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"blobBaseFeeScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"decimals\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"gasPrice\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getL1Fee\",\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getL1FeeUpperBound\",\"inputs\":[{\"name\":\"_unsignedTxSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getL1GasUsed\",\"inputs\":[{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOperatorFee\",\"inputs\":[{\"name\":\"_gasUsed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"isArsia\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"l1BaseFee\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operator\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorFeeConstant\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"operatorFeeScalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"overhead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"owner\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"scalar\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"setArsia\",\"inputs\":[],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setOperator\",\"inputs\":[{\"name\":\"_operator\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"setTokenRatio\",\"inputs\":[{\"name\":\"_tokenRatio\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"tokenRatio\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"transferOwnership\",\"inputs\":[{\"name\":\"_owner\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"version\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"string\",\"internalType\":\"string\"}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"OperatorUpdated\",\"inputs\":[{\"name\":\"previousOperator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOperator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"OwnershipTransferred\",\"inputs\":[{\"name\":\"previousOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"newOwner\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TokenRatioUpdated\",\"inputs\":[{\"name\":\"previousTokenRatio\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"},{\"name\":\"newTokenRatio\",\"type\":\"uint256\",\"indexed\":true,\"internalType\":\"uint256\"}],\"anonymous\":false}]",
	Bin: "0x60e060405234801561001057600080fd5b506001608081905260a052600060c05260805160a05160c051611d8b61004f60003960006107840152600061075b015260006107320152611d8b6000f3fe608060405234801561001057600080fd5b50600436106101a35760003560e01c806368d5dca6116100ee578063de26c4a111610097578063f2fde38b11610071578063f2fde38b14610362578063f45e65d814610375578063f82061401461037d578063fe173b97146102de57600080fd5b8063de26c4a114610329578063e38e91f91461033c578063f1c7a58b1461034f57600080fd5b80638f018a7b116100c85780638f018a7b14610304578063b3ab15fb1461030e578063c59859181461032157600080fd5b806368d5dca6146102d65780636ef25c3a146102de5780638da5cb5b146102e457600080fd5b8063313ce56711610150578063519b4bd31161012a578063519b4bd31461027457806354fd4d501461027c578063570ca7351461029157600080fd5b8063313ce5671461023d57806349948e0e146102445780634d5d9a2a1461025757600080fd5b80631e2b6e7b116101815780631e2b6e7b146101ed578063275aedd2146102225780632e0f26251461023557600080fd5b806306f837d3146101a85780630c18c162146101c457806316d3bc7f146101cc575b600080fd5b6101b160005481565b6040519081526020015b60405180910390f35b6101b1610385565b6101d461040f565b60405167ffffffffffffffff90911681526020016101bb565b6002546102129074010000000000000000000000000000000000000000900460ff1681565b60405190151581526020016101bb565b6101b16102303660046116f0565b610494565b6101b1600681565b60066101b1565b6101b1610252366004611738565b610607565b61025f610645565b60405163ffffffff90911681526020016101bb565b6101b16106ca565b61028461072b565b6040516101bb9190611837565b6002546102b19073ffffffffffffffffffffffffffffffffffffffff1681565b60405173ffffffffffffffffffffffffffffffffffffffff90911681526020016101bb565b61025f6107ce565b486101b1565b6001546102b19073ffffffffffffffffffffffffffffffffffffffff1681565b61030c61082f565b005b61030c61031c366004611888565b6109c2565b61025f610aba565b6101b1610337366004611738565b610b1b565b61030c61034a3660046116f0565b610b9d565b6101b161035d3660046116f0565b610d7d565b61030c610370366004611888565b610e66565b6101b1610fdb565b6101b161103c565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16638b239f736040518163ffffffff1660e01b8152600401602060405180830381865afa1580156103e6573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061040a91906118be565b905090565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff166316d3bc7f6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610470573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061040a91906118d7565b60025460009074010000000000000000000000000000000000000000900460ff166104c157506000919050565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16634d5d9a2a6040518163ffffffff1660e01b8152600401602060405180830381865afa158015610522573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105469190611901565b63ffffffff169050600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff166316d3bc7f6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156105af573d6000803e3d6000fd5b505050506040513d601f19601f820116820180604052508101906105d391906118d7565b67ffffffffffffffff169050806105ea8386611956565b6105f5906064611956565b6105ff9190611993565b949350505050565b60025460009074010000000000000000000000000000000000000000900460ff161561063c576106368261109d565b92915050565b610636826110bc565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16634d5d9a2a6040518163ffffffff1660e01b8152600401602060405180830381865afa1580156106a6573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061040a9190611901565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16635cf249696040518163ffffffff1660e01b8152600401602060405180830381865afa1580156103e6573d6000803e3d6000fd5b60606107567f000000000000000000000000000000000000000000000000000000000000000061111d565b61077f7f000000000000000000000000000000000000000000000000000000000000000061111d565b6107a87f000000000000000000000000000000000000000000000000000000000000000061111d565b6040516020016107ba939291906119ab565b604051602081830303815290604052905090565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff166368d5dca66040518163ffffffff1660e01b8152600401602060405180830381865afa1580156106a6573d6000803e3d6000fd5b3373deaddeaddeaddeaddeaddeaddeaddeaddead0001146108d7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603f60248201527f47617350726963654f7261636c653a206f6e6c7920746865206465706f73697460448201527f6f72206163636f756e742063616e20736574206973417273696120666c61670060648201526084015b60405180910390fd5b60025474010000000000000000000000000000000000000000900460ff1615610981576040517f08c379a0000000000000000000000000000000000000000000000000000000008152602060048201526024808201527f47617350726963654f7261636c653a20417273696120616c726561647920616360448201527f746976650000000000000000000000000000000000000000000000000000000060648201526084016108ce565b600280547fffffffffffffffffffffff00ffffffffffffffffffffffffffffffffffffffff1674010000000000000000000000000000000000000000179055565b60015473ffffffffffffffffffffffffffffffffffffffff163314610a43576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f43616c6c6572206973206e6f7420746865206f776e657200000000000000000060448201526064016108ce565b6002805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907ffbe5b6cbafb274f445d7fed869dc77a838d8243a22c460de156560e8857cad0390600090a35050565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff1663c59859186040518163ffffffff1660e01b8152600401602060405180830381865afa1580156106a6573d6000803e3d6000fd5b60025460009074010000000000000000000000000000000000000000900460ff1615610b7757620f4240610b62610b5184611252565b51610b5d906044611993565b61156f565b610b6d906010611956565b6106369190611a50565b6000610b82836115ce565b9050610b8c610385565b610b969082611993565b9392505050565b60025473ffffffffffffffffffffffffffffffffffffffff163314610c1e576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601a60248201527f43616c6c6572206973206e6f7420746865206f70657261746f7200000000000060448201526064016108ce565b60008111610cae576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152602c60248201527f47617350726963654f7261636c653a20746f6b656e20726174696f206d75737460448201527f206265206e6f6e2d7a65726f000000000000000000000000000000000000000060648201526084016108ce565b67ffffffffffffffff811115610d46576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603260248201527f47617350726963654f7261636c653a20746f6b656e20726174696f206d75737460448201527f206265206c657373207468616e20325e3634000000000000000000000000000060648201526084016108ce565b600080548282556040519091839183917f5d6ae9db2d6725497bed0302a8212c0db5fdb3bd7d14f188a83b5589089caafd91a35050565b60025460009074010000000000000000000000000000000000000000900460ff16610e2a576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152603660248201527f47617350726963654f7261636c653a206765744c314665655570706572426f7560448201527f6e64206f6e6c7920737570706f7274732041727369610000000000000000000060648201526084016108ce565b6000610e37836044611993565b90506000610e4660ff83611a50565b610e509083611993565b610e5b906010611993565b90506105ff8161165e565b60015473ffffffffffffffffffffffffffffffffffffffff163314610ee7576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601760248201527f43616c6c6572206973206e6f7420746865206f776e657200000000000000000060448201526064016108ce565b73ffffffffffffffffffffffffffffffffffffffff8116610f64576040517f08c379a000000000000000000000000000000000000000000000000000000000815260206004820152601d60248201527f6e6577206f776e657220697320746865207a65726f206164647265737300000060448201526064016108ce565b6001805473ffffffffffffffffffffffffffffffffffffffff8381167fffffffffffffffffffffffff0000000000000000000000000000000000000000831681179093556040519116919082907f8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e090600090a35050565b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff16639e8c49666040518163ffffffff1660e01b8152600401602060405180830381865afa1580156103e6573d6000803e3d6000fd5b600073420000000000000000000000000000000000001573ffffffffffffffffffffffffffffffffffffffff1663f82061406040518163ffffffff1660e01b8152600401602060405180830381865afa1580156103e6573d6000803e3d6000fd5b60006106366110ab83611252565b516110b7906044611993565b61165e565b6000806110c883610b1b565b905060006110d46106ca565b6110de9083611956565b905060006110ee6006600a611b84565b905060006110fa610fdb565b6111049084611956565b905060006111128383611a50565b979650505050505050565b60608160000361116057505060408051808201909152600181527f3000000000000000000000000000000000000000000000000000000000000000602082015290565b8160005b811561118a578061117481611b90565b91506111839050600a83611a50565b9150611164565b60008167ffffffffffffffff8111156111a5576111a5611709565b6040519080825280601f01601f1916602001820160405280156111cf576020820181803683370190505b5090505b84156105ff576111e4600183611bc8565b91506111f1600a86611bdf565b6111fc906030611993565b60f81b81838151811061121157611211611bf3565b60200101907effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff1916908160001a90535061124b600a86611a50565b94506111d3565b60606113e1565b818153600101919050565b600082840393505b83811015610b965782810151828201511860001a159093029260010161126c565b825b602082106112d95782516112a4601f83611259565b52602092909201917fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe09091019060210161128f565b8115610b965782516112ee6001840383611259565b520160010192915050565b60006001830392505b610107821061133a5761132c8360ff1661132760fd6113278760081c60e00189611259565b611259565b935061010682039150611302565b60078210611367576113608360ff16611327600785036113278760081c60e00189611259565b9050610b96565b6105ff8360ff166113278560081c8560051b0187611259565b6113d98282036113bd6113ad84600081518060001a8160011a60081b178160021a60101b17915050919050565b639e3779b90260131c611fff1690565b8060021b6040510182815160e01c1860e01b8151188152505050565b600101919050565b6180003860405139618000604051016020830180600d8551820103826002015b81811015611514576000805b50508051604051600082901a600183901a60081b1760029290921a60101b91909117639e3779b9810260111c617ffc16909101805160e081811c878603811890911b9091189091528401908183039084841061146957506114a4565b600184019350611fff821161149e578251600081901a600182901a60081b1760029190911a60101b17810361149e57506114a4565b5061140d565b8383106114b2575050611514565b600183039250858311156114d0576114cd878788860361128d565b96505b6114e4600985016003850160038501611264565b91506114f18782846112f9565b9650506115098461150486848601611380565b611380565b915050809350611401565b5050611526838384885185010361128d565b925050506040519150618000820180820391508183526020830160005b8381101561155b578281015182820152602001611543565b506000920191825250602001604052919050565b60008061157f83620cc394611956565b6115a9907ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd763200611c22565b90506115b96064620f4240611c96565b81121561063657610b966064620f4240611c96565b80516000908190815b81811015611651578481815181106115f1576115f1611bf3565b01602001517fff00000000000000000000000000000000000000000000000000000000000000166000036116315761162a600484611993565b925061163f565b61163c601084611993565b92505b8061164981611b90565b9150506115d7565b506105ff82610440611993565b60008061166a8361156f565b9050600061167661103c565b61167e6107ce565b63ffffffff1661168e9190611956565b6116966106ca565b61169e610aba565b6116a9906010611d52565b63ffffffff166116b99190611956565b6116c39190611993565b90506116d160066002611956565b6116dc90600a611b84565b6116e68284611956565b6105ff9190611a50565b60006020828403121561170257600080fd5b5035919050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b60006020828403121561174a57600080fd5b813567ffffffffffffffff8082111561176257600080fd5b818401915084601f83011261177657600080fd5b81358181111561178857611788611709565b604051601f82017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0908116603f011681019083821181831017156117ce576117ce611709565b816040528281528760208487010111156117e757600080fd5b826020860160208301376000928101602001929092525095945050505050565b60005b8381101561182257818101518382015260200161180a565b83811115611831576000848401525b50505050565b6020815260008251806020840152611856816040850160208701611807565b601f017fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffe0169190910160400192915050565b60006020828403121561189a57600080fd5b813573ffffffffffffffffffffffffffffffffffffffff81168114610b9657600080fd5b6000602082840312156118d057600080fd5b5051919050565b6000602082840312156118e957600080fd5b815167ffffffffffffffff81168114610b9657600080fd5b60006020828403121561191357600080fd5b815163ffffffff81168114610b9657600080fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b6000817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff048311821515161561198e5761198e611927565b500290565b600082198211156119a6576119a6611927565b500190565b600084516119bd818460208901611807565b80830190507f2e0000000000000000000000000000000000000000000000000000000000000080825285516119f9816001850160208a01611807565b60019201918201528351611a14816002840160208801611807565b0160020195945050505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601260045260246000fd5b600082611a5f57611a5f611a21565b500490565b600181815b80851115611abd57817fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04821115611aa357611aa3611927565b80851615611ab057918102915b93841c9390800290611a69565b509250929050565b600082611ad457506001610636565b81611ae157506000610636565b8160018114611af75760028114611b0157611b1d565b6001915050610636565b60ff841115611b1257611b12611927565b50506001821b610636565b5060208310610133831016604e8410600b8410161715611b40575081810a610636565b611b4a8383611a64565b807fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff04821115611b7c57611b7c611927565b029392505050565b6000610b968383611ac5565b60007fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff8203611bc157611bc1611927565b5060010190565b600082821015611bda57611bda611927565b500390565b600082611bee57611bee611a21565b500690565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b6000808212827f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff03841381151615611c5c57611c5c611927565b827f8000000000000000000000000000000000000000000000000000000000000000038412811615611c9057611c90611927565b50500190565b60007f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff600084136000841385830485118282161615611cd757611cd7611927565b7f80000000000000000000000000000000000000000000000000000000000000006000871286820588128184161615611d1257611d12611927565b60008712925087820587128484161615611d2e57611d2e611927565b87850587128184161615611d4457611d44611927565b505050929093029392505050565b600063ffffffff80831681851681830481118215151615611d7557611d75611927565b0294935050505056fea164736f6c634300080f000a",
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

// BaseFeeScalar is a free data retrieval call binding the contract method 0xc5985918.
//
// Solidity: function baseFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleCaller) BaseFeeScalar(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "baseFeeScalar")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// BaseFeeScalar is a free data retrieval call binding the contract method 0xc5985918.
//
// Solidity: function baseFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleSession) BaseFeeScalar() (uint32, error) {
	return _GasPriceOracle.Contract.BaseFeeScalar(&_GasPriceOracle.CallOpts)
}

// BaseFeeScalar is a free data retrieval call binding the contract method 0xc5985918.
//
// Solidity: function baseFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleCallerSession) BaseFeeScalar() (uint32, error) {
	return _GasPriceOracle.Contract.BaseFeeScalar(&_GasPriceOracle.CallOpts)
}

// BlobBaseFee is a free data retrieval call binding the contract method 0xf8206140.
//
// Solidity: function blobBaseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) BlobBaseFee(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "blobBaseFee")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// BlobBaseFee is a free data retrieval call binding the contract method 0xf8206140.
//
// Solidity: function blobBaseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) BlobBaseFee() (*big.Int, error) {
	return _GasPriceOracle.Contract.BlobBaseFee(&_GasPriceOracle.CallOpts)
}

// BlobBaseFee is a free data retrieval call binding the contract method 0xf8206140.
//
// Solidity: function blobBaseFee() view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) BlobBaseFee() (*big.Int, error) {
	return _GasPriceOracle.Contract.BlobBaseFee(&_GasPriceOracle.CallOpts)
}

// BlobBaseFeeScalar is a free data retrieval call binding the contract method 0x68d5dca6.
//
// Solidity: function blobBaseFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleCaller) BlobBaseFeeScalar(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "blobBaseFeeScalar")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// BlobBaseFeeScalar is a free data retrieval call binding the contract method 0x68d5dca6.
//
// Solidity: function blobBaseFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleSession) BlobBaseFeeScalar() (uint32, error) {
	return _GasPriceOracle.Contract.BlobBaseFeeScalar(&_GasPriceOracle.CallOpts)
}

// BlobBaseFeeScalar is a free data retrieval call binding the contract method 0x68d5dca6.
//
// Solidity: function blobBaseFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleCallerSession) BlobBaseFeeScalar() (uint32, error) {
	return _GasPriceOracle.Contract.BlobBaseFeeScalar(&_GasPriceOracle.CallOpts)
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

// GetL1FeeUpperBound is a free data retrieval call binding the contract method 0xf1c7a58b.
//
// Solidity: function getL1FeeUpperBound(uint256 _unsignedTxSize) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCaller) GetL1FeeUpperBound(opts *bind.CallOpts, _unsignedTxSize *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "getL1FeeUpperBound", _unsignedTxSize)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetL1FeeUpperBound is a free data retrieval call binding the contract method 0xf1c7a58b.
//
// Solidity: function getL1FeeUpperBound(uint256 _unsignedTxSize) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleSession) GetL1FeeUpperBound(_unsignedTxSize *big.Int) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetL1FeeUpperBound(&_GasPriceOracle.CallOpts, _unsignedTxSize)
}

// GetL1FeeUpperBound is a free data retrieval call binding the contract method 0xf1c7a58b.
//
// Solidity: function getL1FeeUpperBound(uint256 _unsignedTxSize) view returns(uint256)
func (_GasPriceOracle *GasPriceOracleCallerSession) GetL1FeeUpperBound(_unsignedTxSize *big.Int) (*big.Int, error) {
	return _GasPriceOracle.Contract.GetL1FeeUpperBound(&_GasPriceOracle.CallOpts, _unsignedTxSize)
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

// IsArsia is a free data retrieval call binding the contract method 0x1e2b6e7b.
//
// Solidity: function isArsia() view returns(bool)
func (_GasPriceOracle *GasPriceOracleCaller) IsArsia(opts *bind.CallOpts) (bool, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "isArsia")

	if err != nil {
		return *new(bool), err
	}

	out0 := *abi.ConvertType(out[0], new(bool)).(*bool)

	return out0, err

}

// IsArsia is a free data retrieval call binding the contract method 0x1e2b6e7b.
//
// Solidity: function isArsia() view returns(bool)
func (_GasPriceOracle *GasPriceOracleSession) IsArsia() (bool, error) {
	return _GasPriceOracle.Contract.IsArsia(&_GasPriceOracle.CallOpts)
}

// IsArsia is a free data retrieval call binding the contract method 0x1e2b6e7b.
//
// Solidity: function isArsia() view returns(bool)
func (_GasPriceOracle *GasPriceOracleCallerSession) IsArsia() (bool, error) {
	return _GasPriceOracle.Contract.IsArsia(&_GasPriceOracle.CallOpts)
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
// Solidity: function operatorFeeConstant() view returns(uint64)
func (_GasPriceOracle *GasPriceOracleCaller) OperatorFeeConstant(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "operatorFeeConstant")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint64)
func (_GasPriceOracle *GasPriceOracleSession) OperatorFeeConstant() (uint64, error) {
	return _GasPriceOracle.Contract.OperatorFeeConstant(&_GasPriceOracle.CallOpts)
}

// OperatorFeeConstant is a free data retrieval call binding the contract method 0x16d3bc7f.
//
// Solidity: function operatorFeeConstant() view returns(uint64)
func (_GasPriceOracle *GasPriceOracleCallerSession) OperatorFeeConstant() (uint64, error) {
	return _GasPriceOracle.Contract.OperatorFeeConstant(&_GasPriceOracle.CallOpts)
}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleCaller) OperatorFeeScalar(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _GasPriceOracle.contract.Call(opts, &out, "operatorFeeScalar")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleSession) OperatorFeeScalar() (uint32, error) {
	return _GasPriceOracle.Contract.OperatorFeeScalar(&_GasPriceOracle.CallOpts)
}

// OperatorFeeScalar is a free data retrieval call binding the contract method 0x4d5d9a2a.
//
// Solidity: function operatorFeeScalar() view returns(uint32)
func (_GasPriceOracle *GasPriceOracleCallerSession) OperatorFeeScalar() (uint32, error) {
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

// SetArsia is a paid mutator transaction binding the contract method 0x8f018a7b.
//
// Solidity: function setArsia() returns()
func (_GasPriceOracle *GasPriceOracleTransactor) SetArsia(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _GasPriceOracle.contract.Transact(opts, "setArsia")
}

// SetArsia is a paid mutator transaction binding the contract method 0x8f018a7b.
//
// Solidity: function setArsia() returns()
func (_GasPriceOracle *GasPriceOracleSession) SetArsia() (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetArsia(&_GasPriceOracle.TransactOpts)
}

// SetArsia is a paid mutator transaction binding the contract method 0x8f018a7b.
//
// Solidity: function setArsia() returns()
func (_GasPriceOracle *GasPriceOracleTransactorSession) SetArsia() (*types.Transaction, error) {
	return _GasPriceOracle.Contract.SetArsia(&_GasPriceOracle.TransactOpts)
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
