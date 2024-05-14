package disperser

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/Layr-Labs/datalayr/common/logging"
)

type CodedDataCache struct {
	maxCacheSize    uint64
	currCacheSize   uint64
	cache           map[[32]byte]*Store
	expireDuration  int64
	expireByUnixSec map[[32]byte]int64
	metrics         *Metrics
	cleanPeriod     time.Duration
	mu              *sync.Mutex
	logger          *logging.Logger
}

// expireDuration in sec
func NewCodedDataCache(
	maxCacheSize uint64,
	expireDuration int64,
	cleanPeriod time.Duration,
	metrics *Metrics,
	logger *logging.Logger,
) *CodedDataCache {
	return &CodedDataCache{
		maxCacheSize:    maxCacheSize,
		currCacheSize:   0,
		expireDuration:  expireDuration,
		cleanPeriod:     cleanPeriod,
		metrics:         metrics,
		cache:           make(map[[32]byte]*Store),
		expireByUnixSec: make(map[[32]byte]int64),
		mu:              &sync.Mutex{},
		logger:          logger,
	}
}

func (c *CodedDataCache) Add(store *Store) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, found := c.cache[store.HeaderHash]
	if found {
		return fmt.Errorf("Unable to add item to coded Cache. already found")
	}

	// check if we have enough space
	if c.currCacheSize+uint64(store.UpperBoundBytes()) >= c.maxCacheSize {
		err := fmt.Errorf("Unable to add item to coded Cache. space is full. max  %v. curr space %v", c.maxCacheSize, c.currCacheSize)
		c.logger.Error().Err(err).Msg("Cache does not have sufficient size")
		return err
	}

	c.currCacheSize += uint64(store.UpperBoundBytes())
	c.cache[store.HeaderHash] = store
	c.expireByUnixSec[store.HeaderHash] = time.Now().Unix() + c.expireDuration
	c.logger.Info().Msg("Cache added data")
	return nil
}

// return a store obj, in case expire loop deletes it
func (c *CodedDataCache) Get(headerHash [32]byte) (Store, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	store, found := c.cache[headerHash]
	if !found {
		return Store{}, fmt.Errorf("Unable to get coded Cache. %v not found", hex.EncodeToString(headerHash[:]))
	} else {
		return *store, nil
	}
}

func (c *CodedDataCache) Delete(headerHash [32]byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	store, found := c.cache[headerHash]
	if found {
		c.currCacheSize -= uint64(store.UpperBoundBytes())
		delete(c.cache, headerHash)
		delete(c.expireByUnixSec, headerHash)
		return nil
	} else {
		return fmt.Errorf("Unable to delete coded cache %v. not found", hex.EncodeToString(headerHash[:]))
	}
}

func (c *CodedDataCache) StartExpireLoop() {

	go func() {
		ticker := time.NewTicker(c.cleanPeriod)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				c.logger.Debug().Msgf("Cache auto clean, curr size %v", c.currCacheSize)
				c.mu.Lock()
				now := t.Unix()
				items := make([][32]byte, 0)
				// collect expired items
				for h, expiration := range c.expireByUnixSec {
					if expiration < now {
						items = append(items, h)
					}
				}

				// delete expired items
				for _, headerHash := range items {
					store, found := c.cache[headerHash]
					if !found {
						c.logger.Error().Msg("Data structure cache expireByUnixSec are not consistent")
					}
					c.currCacheSize -= uint64(store.UpperBoundBytes())
					delete(c.cache, headerHash)
					delete(c.expireByUnixSec, headerHash)
				}
				c.mu.Unlock()
			}
		}
	}()
}
