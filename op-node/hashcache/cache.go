package hashcache

import "github.com/ethereum/go-ethereum/common"

var OpNodeBlockHashCache map[common.Hash]common.Hash

func init() {
	OpNodeBlockHashCache = make(map[common.Hash]common.Hash)
}
