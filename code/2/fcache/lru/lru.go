package lru

import "container/list"

type (
	// Cache the main struct to manage the cache
	Cache struct {
		// the maximum bytes can be used in the cache
		maxBytes int64
		// the number of bytes has been used, both the length of key and value are counted
		nbytes int64
		// the List to organize the entries and to realise the LRU
		ll *list.List
		// store the key and according value, aim to retrieve a key in O(1) complexity
		cache map[string]*list.Element
		// optional and executed when an entry is purged
		// may call it callback function ?
		OnEvicted func(key string, value Value)
	}

	//	entry the type of node in the list
	//	the reason why entry has the key field is to remove the head node with key in the map
	entry struct {
		key   string
		value Value
	}

	// Value use Len to Count how many bytes it takes
	Value interface {
		Len() int
	}
)

// New Initialise the cache, set the maximum bytes can be used and extra function called when a entry evicted
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get retrieve the value and ok
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// RemoveOldest discard the oldest entry
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		// remove the entry from list
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		// remove the entry from map
		delete(c.cache, kv.key)
		// update the now bytes
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		// execute the optional OnEvicted if exist
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// when the key exists, update it
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(len(kv.key)) + int64(kv.value.Len())
		kv.value = value
	} else {
		// when the key is not existing, add it
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	// remove the oldest entry to maintain the maximum bytes constraint
	// when maxBytes is zero, it means there is no constraint
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
