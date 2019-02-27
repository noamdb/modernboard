package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/noamdb/modernboard/repository"
)

const boardsPrefix = "boards"

type BoardsCache struct {
	cache
}

func (c *BoardsCache) InsertBoards(boards []repository.Board) []byte {
	b, err := json.Marshal(boards)
	if err != nil {
		fmt.Println("error while inserting boards", err)
		return []byte{}
	}
	c.set(boardsPrefix, b, time.Hour*2)
	return b
}

func (c *BoardsCache) GetBoards() ([]byte, bool) {
	boards, exists := c.get(boardsPrefix)
	if exists {
		return boards.([]byte), exists
	}
	return nil, false
}
