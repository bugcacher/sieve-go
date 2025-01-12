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

type Key comparable // Represents the type of the key in the cache, must support comparison
type Value any      // Represents the type of the value in the cache, can be any type

// keyNodeMap maps keys to linked list elements, used for O(1) lookup in the cache
type keyNodeMap[K Key, V *list.Element] map[K]V

// nodeEntry stores cache data (key-value pair) and the visited flag for eviction
type nodeEntry[K comparable, V Value] struct {
	visited bool // Tracks whether this entry has been visited during eviction cycle
	key     K    // The key for this entry
	value   V    // The value for this entry
}

// Item represents a key-value pair for exporting cache data
type Item[K Key, V Value] struct {
	Key   K // The key in the cache
	Value V // The value in the cache
}

// Cache represents a Sieve cache with a given capacity
type Cache[K Key, V Value] struct {
	capacity int64                        // The maximum number of items the cache can hold
	size     int64                        // The current number of items in the cache
	q        *list.List                   // Doubly linked list to maintain cache order
	keysMap  keyNodeMap[K, *list.Element] // Map of keys to their respective list elements
	hand     *list.Element                // Points to the hand in the cache, used for eviction tracking
}

// NewCache initializes a new cache with the given capacity
func NewCache[K Key, V Value](capacity int64) *Cache[K, V] {
	cache := &Cache[K, V]{capacity: capacity}
	cache.init()
	return cache
}

// Size returns the current number of items in the cache
func (c *Cache[K, V]) Size() int64 {
	return c.size
}

// Capacity returns the maximum capacity of the cache
func (c *Cache[K, V]) Capacity() int64 {
	return c.capacity
}

// Set adds a new key-value pair to the cache, evicting an entry if necessary
func (c *Cache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	if _, err := c.Get(key); err == nil {
		return // If key already exists, skip inserting
	}
	if c.size == c.capacity {
		c.evict()
	}
	element := c.q.PushFront(&nodeEntry[K, V]{key: key, value: value})
	c.keysMap[key] = element
	c.size++
}

// Get retrieves the value for a given key, returns an error if the key is not found
func (c *Cache[K, V]) Get(key K) (V, error) {
	var value V
	ele, ok := c.keysMap[key]
	if !ok {
		return value, ErrKeyNotFound // Return error if key not found
	}
	entry := ele.Value.(*nodeEntry[K, V])
	entry.visited = true // Mark this entry as visited
	value = entry.value
	return value, nil
}

// Keys returns a slice of all keys currently in the cache
func (c *Cache[K, V]) Keys() []K {
	var keys []K
	for k := range c.keysMap {
		keys = append(keys, k)
	}
	return keys
}

// Items returns a slice of all key-value pairs (Item) currently in the cache
func (c *Cache[K, V]) Items() []Item[K, V] {
	var items []Item[K, V]
	for k, ele := range c.keysMap {
		entry := ele.Value.(*nodeEntry[K, V])
		items = append(items, Item[K, V]{Key: k, Value: entry.value})
	}
	return items
}

// Print outputs the current cache state, showing each key's value and visited status
func (c *Cache[K, V]) Print() {
	for curr := c.q.Front(); curr != nil; curr = curr.Next() {
		ele := curr.Value.(*nodeEntry[K, V])
		fmt.Printf("%s: %v\t", ele.value, ele.visited)
	}
	fmt.Println("\n")
}

// Clear resets the cache to its initial empty state
func (c *Cache[K, V]) Clear() {
	c.init()
}

// Delete removes a key-value pair from the cache and returns the value, or an error if not found
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

// Evict removes the first unvisited entry to the left of hand from the cache and returns the evicted key
func (c *Cache[K, V]) Evict() (K, error) {
	return c.evict()
}

// Contains checks whether the cache contains a given key
func (c *Cache[K, V]) Contains(key K) bool {
	_, ok := c.keysMap[key]
	return ok
}

// Resize changes the capacity of the cache, evicting items if necessary
func (c *Cache[K, V]) Resize(newCapacity int64) []K {
	var evictedKeys []K
	if newCapacity >= c.capacity {
		c.capacity = newCapacity
		return evictedKeys // No need to evict if new capacity is greater than or equal to current capacity
	}
	// Evict items if the new capacity is smaller
	keysToEvictCount := c.Size() - newCapacity
	for keysToEvictCount > 0 {
		if key, err := c.evict(); err == nil {
			evictedKeys = append(evictedKeys, key)
		}
		keysToEvictCount--
	}
	c.capacity = newCapacity
	return evictedKeys
}

// init reinitializes the cache to an empty state
func (c *Cache[K, V]) init() {
	c.size = 0
	c.q = list.New()
	c.keysMap = make(keyNodeMap[K, *list.Element], c.capacity)
	c.hand = nil
}

// evict removes the first unvisited entry to the left of hand from the cache, returns the evicted key
func (c *Cache[K, V]) evict() (K, error) {
	var evictedKey K
	if c.size == 0 {
		return evictedKey, ErrEmptyCache
	}
	curr := c.hand
	if curr == nil {
		curr = c.q.Back() // Start from the back if no hand set
	}
	// Traverse to find an unvisited node to evict
	for {
		entry := curr.Value.(*nodeEntry[K, V])
		if !entry.visited {
			break // Found an unvisited entry to evict
		}
		entry.visited = false // Mark the entry as not visited
		curr = curr.Prev()
		if curr == nil {
			curr = c.q.Back() // Loop around if we've reached the beginning
		}
	}
	c.hand = curr.Prev() // Update hand for next eviction
	c.q.Remove(curr)
	c.size--
	entry := curr.Value.(*nodeEntry[K, V])
	delete(c.keysMap, entry.key)
	evictedKey = entry.key
	return evictedKey, nil
}
