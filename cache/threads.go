package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/noamdb/modernboard/repository"
)

const trending = "trending"

type TrendingThreadsCache struct {
	cache
}

func (c *TrendingThreadsCache) InsertThreads(threads []repository.TrendingThread) []byte {
	j, err := json.Marshal(threads)
	if err != nil {
		fmt.Println("error while inserting user boards", err)
		return []byte{}
	}
	c.set(trending, j, time.Minute*2)
	return j
}

func (c *TrendingThreadsCache) GetThreads() ([]byte, bool) {
	threads, exists := c.get(trending)
	if exists {
		return threads.([]byte), exists
	}
	return nil, false
}

type ThreadsPageCache struct {
	cache
}

func (c *ThreadsPageCache) InsertPage(URI string, page int, threads []repository.ThreadWithOP) []byte {
	j, err := json.Marshal(threads)
	if err != nil {
		fmt.Println("error while inserting threads page", err)
		return []byte{}
	}

	c.set(fmt.Sprintf("%s_%d", URI, page), j, time.Second*6)
	return j
}

func (c *ThreadsPageCache) GetPage(URI string, page int) ([]byte, bool) {
	j, exists := c.get(fmt.Sprintf("%s_%d", URI, page))
	if exists {
		return j.([]byte), exists
	}
	return nil, false
}

type ThreadCache struct {
	cache
}

func (c *ThreadCache) InsertThread(threadID int, thread repository.ThreadWithPosts) []byte {
	j, err := json.Marshal(thread)
	if err != nil {
		fmt.Println("error while inserting thread", err)
		return []byte{}
	}
	c.set(string(threadID), j, time.Second*7)
	return j
}

func (c *ThreadCache) GetThread(threadID int) ([]byte, bool) {
	threads, exists := c.get(string(threadID))
	if exists {
		return threads.([]byte), exists
	}
	return nil, false
}
