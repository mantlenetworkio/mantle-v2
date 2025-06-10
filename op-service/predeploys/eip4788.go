package predeploys

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

// EIP-4788 defines a deterministic deployment transaction that deploys the beacon-block-roots contract.
// To embed the contract in genesis, we want the deployment-result, not the contract-creation tx input code.
// Since the contract deployment result is deterministic and the same across every chain,
// the bytecode can be easily verified by comparing it with chains like Goerli.
// During deployment it does not modify any contract storage, the storage starts empty.
// See https://goerli.etherscan.io/tx/0xdf52c2d3bbe38820fff7b5eaab3db1b91f8e1412b56497d88388fb5d4ea1fde0
// And https://eips.ethereum.org/EIPS/eip-4788
var (
	EIP4788ContractAddr     = params.BeaconRootsAddress
	EIP4788ContractCode     = params.BeaconRootsCode
	EIP4788ContractCodeHash = common.HexToHash("0xf57acd40259872606d76197ef052f3d35588dadf919ee1f0e3cb9b62d3f4b02c")
	EIP4788ContractDeployer = common.HexToAddress("0x0B799C86a49DEeb90402691F1041aa3AF2d3C875")
)
