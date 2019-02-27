// Package cache is a modification of https://github.com/patrickmn/go-cache
package cache

import (
	"fmt"
	"sync"
	"time"

	"gitlab.com/noamdb/modernboard/repository"
)

type Cache struct {
	*BanCache
	*BoardsCache
	// *UserBoardsCache
	*TrendingThreadsCache
	*ThreadsPageCache
	*ThreadCache
	*repository.Repository
}

type item struct {
	Value      interface{}
	Expiration int64
}

type cache struct {
	items map[string]item
	mtx   sync.RWMutex
}

func (c *cache) set(k string, x interface{}, d time.Duration) {
	var e int64
	if d > 0 {
		e = time.Now().Add(d).Unix()
	}
	c.mtx.Lock()
	c.items[k] = item{
		Value:      x,
		Expiration: e,
	}
	c.mtx.Unlock()
}

func (c *cache) get(k string) (interface{}, bool) {
	c.mtx.RLock()
	item, found := c.items[k]
	if !found {
		c.mtx.RUnlock()
		return "", false
	}
	c.mtx.RUnlock()
	return item.Value, true
}

func (c *cache) run(d time.Duration) {
	for range time.Tick(d) {
		c.deleteExpired()
	}
}

func (c *cache) deleteExpired() {
	now := time.Now().Unix()
	c.mtx.Lock()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
		}
	}
	c.mtx.Unlock()
}

func new() cache {
	return cache{
		items: make(map[string]item),
	}
}

func (c *cache) Flush() {
	c.mtx.Lock()
	c.items = map[string]item{}
	c.mtx.Unlock()
}

func (c *Cache) Init() {
	fmt.Println("initializig cache")
	c.BanCache = &BanCache{new(), c.Repository}
	c.BoardsCache = &BoardsCache{new()}
	c.TrendingThreadsCache = &TrendingThreadsCache{new()}
	c.ThreadsPageCache = &ThreadsPageCache{new()}
	c.ThreadCache = &ThreadCache{new()}
	go c.BoardsCache.run(time.Hour)
	go c.BanCache.scheduleRefresh()
	go c.TrendingThreadsCache.run(time.Minute * 1)
	go c.ThreadsPageCache.run(time.Second * 3)
	go c.ThreadCache.run(time.Second * 4)
}
