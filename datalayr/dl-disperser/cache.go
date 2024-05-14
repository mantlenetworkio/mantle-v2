package disperser

import (
	"bytes"
	"sync"

	rs "github.com/Layr-Labs/datalayr/lib/encoding/kzgEncoder"
)

// A simple cache that caches the first eligible item that it sees
type StoreCache struct {
	digest         [32]byte
	kzgCommit      [64]byte
	lowDegreeProof [64]byte
	frames         []rs.Frame
	mu             *sync.Mutex
}

func NewStoreCache() *StoreCache {
	var digest [32]byte
	var kzgCommit [64]byte
	var lowDegreeProof [64]byte
	return &StoreCache{
		digest:         digest,
		kzgCommit:      kzgCommit,
		lowDegreeProof: lowDegreeProof,
		frames:         make([]rs.Frame, 0),
		mu:             &sync.Mutex{},
	}
}

func (c *StoreCache) IsEmpty(u []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.frames) == 0
}

func (c *StoreCache) CheckHit(u []byte) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if bytes.Equal(u[:], c.digest[:]) {
		return true
	} else {
		return false
	}
}

func (c *StoreCache) Get() ([]byte, []byte, []rs.Frame) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.kzgCommit[:], c.lowDegreeProof[:], c.frames
}

func (c *StoreCache) Put(key []byte, kzgCommit, lowDegreeProof []byte, frames []rs.Frame) {
	c.mu.Lock()
	defer c.mu.Unlock()

	copy(c.digest[:], key[:])
	copy(c.kzgCommit[:], kzgCommit[:])
	copy(c.lowDegreeProof[:], lowDegreeProof[:])
	c.frames = frames
}
