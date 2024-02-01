package l1blockhashcache

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/log"
)

var OpNodeBlockHashCache map[common.Hash]common.Hash

func init() {
	OpNodeBlockHashCache = make(map[common.Hash]common.Hash)
}

func SetCacheBlockHash(hash common.Hash, cacheHash common.Hash) {
	log.Info("op-node hash cache SetCacheBlockHash", "hash", hash.String(), "cacheHash", cacheHash.String())
	OpNodeBlockHashCache[hash] = cacheHash
}

func GetCacheBlockHash(hash common.Hash) common.Hash {
	cacheBlockHash, ok := OpNodeBlockHashCache[hash]
	log.Info("op-node hash cache GetCacheBlockHash", "hash", hash.String(), "cacheHash", cacheBlockHash.String())
	if !ok {
		return hash
	}
	return cacheBlockHash
}
