package sieve

import (
	"container/list"
	"errors"
	"fmt"
)

var (
	ErrKeyNotFound     = errors.New("key not found")
	ErrEmptyCache      = errors.New("cache is empty")
	ErrInvalidCapacity = errors.New("capacity should be greater than 0")
)

type Key comparable
type Value any

type keyNodeMap[K Key, V *list.Element] map[K]V

type nodeEntry[K comparable, V Value] struct {
	visited bool
	key     K
	value   V
}

type Item[K Key, V Value] struct {
	Key   K
	Value V
}

type Cache[K Key, V Value] struct {
	capacity int64
	size     int64
	q        *list.List
	keysMap  keyNodeMap[K, *list.Element]
	hand     *list.Element
}

func NewCache[K Key, V Value](capacity int64) *Cache[K, V] {
	cache := &Cache[K, V]{
		capacity: capacity,
	}
	cache.init()
	return cache
}

func (c *Cache[K, V]) Size() int64 {
	return c.size
}

func (c *Cache[K, V]) Capacity() int64 {
	return c.capacity
}

func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	if _, err := c.Get(key); err == nil {
		return
	}
	if c.size == c.capacity {
		c.evict()
	}
	element := c.q.PushFront(&nodeEntry[K, V]{key: key, value: value})
	c.keysMap[key] = element
	c.size += 1
}

func (c *Cache[K, V]) Get(key K) (V, error) {
	var value V
	ele, ok := c.keysMap[key]
	if !ok {
		return value, ErrKeyNotFound
	}
	entry := ele.Value.(*nodeEntry[K, V])
	entry.visited = true
	value = entry.value
	return value, nil
}

func (c *Cache[K, V]) Keys() []K {
	var keys []K
	for k, _ := range c.keysMap {
		keys = append(keys, k)
	}
	return keys
}

func (c *Cache[K, V]) Items() []Item[K, V] {
	var items []Item[K, V]
	for k, ele := range c.keysMap {
		entry := ele.Value.(*nodeEntry[K, V])
		items = append(items, Item[K, V]{Key: k, Value: entry.value})
	}
	return items
}

func (c *Cache[K, V]) Print() {
	for curr := c.q.Front(); curr != nil; curr = curr.Next() {
		ele := curr.Value.(*nodeEntry[K, V])
		fmt.Printf("%s: %v\t", ele.value, ele.visited)
	}
	fmt.Println("\n")
}

func (c *Cache[K, V]) Clear() {
	c.init()
}

func (c *Cache[K, V]) Delete(key K) (V, error) {
	var value V
	ele, ok := c.keysMap[key]
	if !ok {
		return value, ErrKeyNotFound
	}
	delete(c.keysMap, key)
	entry := c.q.Remove(ele).(*nodeEntry[K, V])
	value = entry.value
	return value, nil
}

func (c *Cache[K, V]) Evict() (K, error) {
	return c.evict()
}

func (c *Cache[K, V]) Contains(key K) bool {
	_, ok := c.keysMap[key]
	return ok
}

func (c *Cache[K, V]) Resize(newCapacity int64) []K {
	var evictedKeys []K
	if newCapacity >= c.capacity {
		c.capacity = newCapacity
		return evictedKeys
	}
	keysToEvictCount := c.Size() - newCapacity
	for keysToEvictCount > 0 {
		fmt.Println(keysToEvictCount)
		if key, err := c.evict(); err == nil {
			evictedKeys = append(evictedKeys, key)
		}
		keysToEvictCount -= 1
	}
	c.capacity = newCapacity
	return evictedKeys
}

func (c *Cache[K, V]) init() {
	c.size = 0
	c.q = list.New()
	c.keysMap = make(keyNodeMap[K, *list.Element], c.capacity)
	c.hand = nil
}

// TODO: test when nothing to evict
func (c *Cache[K, V]) evict() (K, error) {
	var evictedKey K
	if c.size == 0 {
		return evictedKey, ErrEmptyCache
	}
	curr := c.hand
	if curr == nil {
		curr = c.q.Back()
	}
	for {
		entry := curr.Value.(*nodeEntry[K, V])
		if !entry.visited {
			break
		}
		entry.visited = false
		curr = curr.Prev()
		if curr == nil {
			curr = c.q.Back()
		}
	}
	c.hand = curr.Prev()
	c.q.Remove(curr)
	c.size -= 1
	entry := curr.Value.(*nodeEntry[K, V])
	delete(c.keysMap, entry.key)
	evictedKey = entry.key
	return evictedKey, nil
}
