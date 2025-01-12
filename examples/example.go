package main

import (
	"fmt"

	"github.com/bugcacher/sieve-go"
)

func main() {
	// Initialize a new cache with a capacity of 3
	cache := sieve.NewCache[string, int](3)
	cache.Set("A", 1)
	cache.Set("B", 2)
	cache.Set("C", 3)

	// Get a value by key "B"
	valueB, err := cache.Get("B")
	if err != nil {
		fmt.Println("Error getting 'B':", err)
	} else {
		fmt.Printf("Value for key 'B': %d\n", valueB)
	}

	// Print the cache state (using the String method)
	fmt.Println("Cache state after adding 'A', 'B' and 'C':")
	fmt.Println(cache)

	// Get keys from the cache
	fmt.Printf("cache.Keys(): %v\n", cache.Keys()) // Expected output: [A, B]

	// Resize the cache to a new capacity of 2 and evict old items if needed
	evictedKeys := cache.Resize(2)
	fmt.Printf("Evicted keys after resizing: %v\n", evictedKeys)

	// Add more items to the cache
	cache.Set("D", 4)
	cache.Set("E", 5)

	// Print updated cache details
	fmt.Println("Cache state after resizing and adding more items:")
	fmt.Println(cache)

	// Check cache size and capacity
	fmt.Printf("cache.Size(): %v\n", cache.Size())         // Should print current size
	fmt.Printf("cache.Capacity(): %v\n", cache.Capacity()) // Should print the capacity (2 after resizing)

	// Check keys again after adding more items
	fmt.Printf("cache.Keys() after more adds: %v\n", cache.Keys()) // Expected output: [D, E] or similar

	// Evict an item (this will remove the least recently used item)
	evictedKey, err := cache.Evict()
	if err != nil {
		fmt.Println("Error during eviction:", err)
	} else {
		fmt.Printf("Evicted key: %v\n", evictedKey)
	}

	// Final state of the cache
	fmt.Println("Final Cache state after eviction:")
	fmt.Println(cache)

	// Try accessing a key that might have been evicted
	if value, err := cache.Get("A"); err != nil {
		fmt.Println("Error getting 'A':", err)
	} else {
		fmt.Printf("Value for 'A': %d\n", value)
	}

	// Clear the cache (removes all entries)
	cache.Clear()
	fmt.Println("Cache state after clearing:")
	fmt.Println(cache) // Should show empty cache
}
