package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte // raw data we are cashing
}

type Cache struct {
	entry map[string]cacheEntry
	mutex sync.Mutex
}

func (c *Cache) Add(key string, val []byte) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry := cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.entry[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	// The bool should be true if the entry was found and false if it wasn't
	entry, exists := c.entry[key]
	if !exists {
		return nil, false
	} else {
		return entry.val, true
	}

}

func (c *Cache) reapLoop(interval time.Duration) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		<-ticker.C // wait for the next tick
		c.mutex.Lock()
		for k, v := range c.entry {
			if time.Since(v.createdAt) > interval {
				delete(c.entry, k)
				fmt.Println("Evicted:", k)
			}
		}
		c.mutex.Unlock()
	}

}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{entry: make(map[string]cacheEntry)}
	go cache.reapLoop(interval)
	return cache
}
