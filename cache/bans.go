package cache

import (
	"fmt"
	"time"

	"gitlab.com/noamdb/modernboard/repository"
)

type BanCache struct {
	cache
	*repository.Repository
}

func (c *BanCache) InsertBan(IP string, reason string) {
	c.set(IP, reason, -1)
}

func (c *BanCache) IsBanned(IP string) (string, bool) {
	reason, exists := c.get(IP)
	return reason.(string), exists
}

func (c *BanCache) Refresh() {
	bans, err := c.Repository.GetBans()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	c.Flush()
	for _, ban := range bans {
		c.InsertBan(ban.IP, ban.Reason)
	}
	fmt.Println("refresh bans")
}

func (c *BanCache) scheduleRefresh() {
	ticker := time.NewTicker(time.Minute * 30)
	go func() {
		for ; true; <-ticker.C {
			c.Refresh()
		}
	}()
}
