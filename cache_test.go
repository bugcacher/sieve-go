package sieve

import (
	"cmp"
	"fmt"
	"slices"
	"testing"
)

func assertEqual[K comparable](t *testing.T, expected, actual K, msg string) {
	if expected != actual {
		t.Errorf("%s expected %v, but got %v", msg, expected, actual)
	}
}

func assertEqualSlice[O cmp.Ordered](t *testing.T, expected, actual []O, msg string) {
	slices.Sort(expected)
	slices.Sort(actual)
	if slices.Compare(expected, actual) != 0 {
		t.Errorf("%s expected %v, but got %v", msg, expected, actual)
	}
}

func assertErrorEqual(t *testing.T, expected, actual error) {
	if expected != actual {
		t.Errorf("expected error: %v, but got %v", expected, actual)
	}
}

func assertErrorNil(t *testing.T, actual error) {
	if actual != nil {
		t.Errorf("expected error: nil, but got %v", actual)
	}
}

func TestCache_Capacity(t *testing.T) {
	capacity := int64(2)
	cache := NewCache[string, string](int64(capacity))
	assertEqual(t, capacity, cache.Capacity(), "Invalid capacity")
}

func TestCache_Size(t *testing.T) {
	capacity := int64(4)
	cache := NewCache[string, string](capacity)
	items := []string{"a", "b", "c"}
	for _, item := range items {
		cache.Set(item, item)
	}
	assertEqual(t, int64(len(items)), cache.Size(), "Invalid size")
	cache.Clear()

	items = append(items, "d", "e", "f")
	for _, item := range items {
		cache.Set(item, item)
	}
	assertEqual(t, min(int64(len(items)), cache.Capacity()), cache.Size(), "Invalid size")
}

func TestCache_SetAndGet(t *testing.T) {
	cache := NewCache[string, string](5)
	cache.Set("a", "A")
	cache.Set("b", "B")
	cache.Set("c", "C")

	// Get items from the cache
	tests := []struct {
		key      string
		expected string
		err      error
	}{
		{"a", "A", nil},
		{"b", "B", nil},
		{"d", "", ErrKeyNotFound},
	}

	for _, test := range tests {
		value, err := cache.Get(test.key)
		if test.err != nil {
			assertErrorEqual(t, test.err, err)
		} else {
			assertEqual(t, test.expected, value, fmt.Sprintf("For key %s", test.key))
		}
	}
}
func TestCache_MissingGet(t *testing.T) {
	cache := NewCache[string, string](2)
	key, value := "a", "A"
	cache.Set(key, value)
	retunredVal, err := cache.Get(key)
	assertErrorNil(t, err)
	assertEqual(t, value, retunredVal, fmt.Sprintf("For key %s", key))

	cache.Delete(key)
	_, err = cache.Get(key)
	assertErrorEqual(t, ErrKeyNotFound, err)
}

func TestCache_Evict(t *testing.T) {
	cache := NewCache[string, string](5)

	cache.Set("a", "A")
	cache.Set("b", "B")
	cache.Set("c", "C")

	evictedKey, err := cache.Evict()
	assertEqual(t, "a", evictedKey, "")
	assertErrorNil(t, err)

	cache.Clear()

	cache.Set("a", "A")
	cache.Set("b", "B")
	cache.Set("c", "C")

	cache.Get("a") // key `a` will be visited so it should not be evicted first

	evictedKey, err = cache.Evict()
	assertEqual(t, "b", evictedKey, "")
	assertErrorNil(t, err)

	cache.Clear()
	evictedKey, err = cache.Evict()
	assertErrorEqual(t, ErrEmptyCache, err)

	cache.Clear()

	cache.Set("a", "A")
	cache.Set("b", "B")
	cache.Set("c", "C")
	cache.Set("a", "A")

	cache.Get("b")
	cache.Get("c")

	evictedKey, err = cache.Evict()
	assertEqual(t, "a", evictedKey, "")
	assertErrorNil(t, err)

}

func TestCache_Resize(t *testing.T) {
	cache := NewCache[string, string](3)

	cache.Set("a", "A")
	cache.Set("b", "B")
	cache.Set("c", "C")

	// Resize the cache to a bigger capacity
	evictedKeys := cache.Resize(5)
	assertEqual(t, int64(5), cache.Capacity(), "Invalid capacity")
	assertEqual(t, int64(0), int64(len(evictedKeys)), "Invalid capacity")

	// Resize the cache to a smaller capacity
	evictedKeys = cache.Resize(2)
	assertEqual(t, int64(2), cache.Capacity(), "Invalid capacity")
	assertEqual(t, int64(1), int64(len(evictedKeys)), "Invalid capacity")
}

func TestCache_Items(t *testing.T) {
	cache := NewCache[string, string](3)

	cache.Set("a", "A")
	cache.Set("b", "B")
	cache.Set("c", "C")

	expectedKeys := []string{"a", "b", "c"}
	expectedValues := []string{"A", "B", "C"}

	var actualKeys, actualValues []string
	for _, item := range cache.Items() {
		actualKeys = append(actualKeys, item.Key)
		actualValues = append(actualValues, item.Value)
	}
	assertEqualSlice(t, expectedKeys, actualKeys, "")
	assertEqualSlice(t, expectedValues, actualValues, "")
	cache.String()
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache[string, string](5)
	cache.Set("a", "A")
	cache.Set("b", "B")
	expectedKeys := []string{"a", "b"}
	assertEqual(t, int64(2), cache.Size(), "")
	assertEqualSlice(t, expectedKeys, cache.Keys(), "")

	cache.Clear()
	expectedKeys = []string{}
	assertEqual(t, int64(0), cache.Size(), "")
	assertEqualSlice(t, expectedKeys, cache.Keys(), "")
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache[string, string](5)
	cache.Set("a", "A")
	tests := []struct {
		key      string
		expected string
		err      error
	}{
		{
			key:      "a",
			expected: "A",
			err:      nil,
		},
		{
			key:      "b",
			expected: "",
			err:      ErrKeyNotFound,
		},
	}
	for _, test := range tests {
		val, err := cache.Delete(test.key)
		if test.err != nil {
			assertErrorEqual(t, test.err, err)
		} else {
			assertEqual(t, test.expected, val, "")
		}
	}
}

func TestCache_Contains(t *testing.T) {
	cache := NewCache[string, string](5)

	cache.Set("a", "A")
	cache.Set("b", "B")

	assertEqual(t, true, cache.Contains("a"), "")
	assertEqual(t, true, cache.Contains("b"), "")
	assertEqual(t, false, cache.Contains("c"), "")
}
