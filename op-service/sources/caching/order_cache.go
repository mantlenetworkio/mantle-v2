package caching

import (
	"github.com/google/btree"
	"sync"
)

type item[V any] struct {
	number uint64
	value  V
}

func (a item[V]) Less(b btree.Item) bool {
	return a.number < b.(item[V]).number
}

type OrderCache[V any] struct {
	m       Metrics
	label   string
	data    *btree.BTree
	lock    sync.Mutex
	maxSize int
}

func NewOrderCache[V any](m Metrics, label string, maxSize int) *OrderCache[V] {
	return &OrderCache[V]{
		m:       m,
		label:   label,
		data:    btree.New(32), //
		maxSize: maxSize,
	}
}

func (v *OrderCache[V]) Add(key uint64, value V) bool {
	defer v.lock.Unlock()
	v.lock.Lock()

	// check if  it already exits
	if v.data.Has(item[V]{number: key}) {
		return false
	}

	// check if full
	if v.data.Len() >= v.maxSize {
		return false
	}

	// add new item
	v.data.ReplaceOrInsert(item[V]{
		number: key,
		value:  value,
	})
	if v.m != nil {
		v.m.CacheAdd(v.label, v.data.Len(), false)
	}
	return true
}

func (v *OrderCache[V]) AddIfNotFull(key uint64, value V) (success bool, isFull bool) {
	defer v.lock.Unlock()
	v.lock.Lock()

	// check if  it already exits
	if v.data.Has(item[V]{number: key}) {
		v.data.ReplaceOrInsert(item[V]{
			number: key,
			value:  value,
		})
		return true, false
	}

	// check if full
	if v.data.Len() >= v.maxSize {
		return false, true
	}

	// add new item
	v.data.ReplaceOrInsert(item[V]{
		number: key,
		value:  value,
	})
	if v.m != nil {
		v.m.CacheAdd(v.label, v.data.Len(), false)
	}
	return true, false
}

func (v *OrderCache[V]) IsFull() bool {
	return v.data.Len() >= v.maxSize
}

func (v *OrderCache[V]) Get(key uint64, recordMetrics bool) (V, bool) {
	defer v.lock.Unlock()
	v.lock.Lock()

	i := v.data.Get(item[V]{number: key})
	if i == nil {
		var zero V
		return zero, false
	}
	if v.m != nil && recordMetrics {
		v.m.CacheGet(v.label, true)
	}
	return i.(item[V]).value, true
}

func (v *OrderCache[V]) RemoveAll() {
	defer v.lock.Unlock()
	v.lock.Lock()
	v.data.Clear(false)
}

func (v *OrderCache[V]) RemoveLessThan(p uint64) (isRemoved bool) {
	defer v.lock.Unlock()
	v.lock.Lock()

	// Traverse and delete elements less than p
	v.data.Ascend(func(i btree.Item) bool {
		block := i.(item[V])
		if block.number < p {
			v.data.Delete(i)
			isRemoved = true
			return true // continue
		}
		return false // stop
	})
	return
}

func (v *OrderCache[V]) RemoveGreaterThan(p uint64) (isRemoved bool) {
	defer v.lock.Unlock()
	v.lock.Lock()

	// Traverse and delete elements less than p
	v.data.Ascend(func(i btree.Item) bool {
		block := i.(item[V])
		if block.number > p {
			v.data.Delete(i)
			isRemoved = true
			return true // continue
		}
		return false // stop
	})
	return
}

func (v *OrderCache[V]) AddAndRemove(key uint64, value V) {
	defer v.lock.Unlock()
	v.lock.Lock()

	// check if full
	if v.data.Len() == v.maxSize {
		v.data.DeleteMin()
	}

	// add new item
	v.data.ReplaceOrInsert(item[V]{
		number: key,
		value:  value,
	})
	if v.m != nil {
		v.m.CacheAdd(v.label, v.data.Len(), false)
	}
}
