package tasks

// import (
// 	"fmt"
// 	"time"
// )

// func (t Tasks) ReloadBans() {
// 	bans, err := t.Repo.GetBans()
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return
// 	}
// 	t.BanCache.Flush()
// 	for _, ban := range bans {
// 		t.BanCache.InsertBan(ban.IP, ban.Reason)
// 	}
// 	fmt.Println("reload bans")
// }

// // func (t Tasks) ReloadBoards() {
// // 	boards, err := t.Repo.GetBoards()

// // 	if err != nil {
// // 		fmt.Println(err.Error())
// // 		return
// // 	}
// // 	b, err := json.Marshal(boards)
// // 	t.BoardsCache.RefreshBoards(b)
// // 	fmt.Println(t.BoardsCache.Get("boards"))
// // 	fmt.Println("reload boards")
// // }

// func (t Tasks) RunCacheTasks() {
// 	// add here chache ruun function if cache has expiry e.g - go t.BanCache.Run()
// 	// reload bans every 30 minutes
// 	tickerThirtyMinutes := time.NewTicker(time.Minute * 30)
// 	go func() {
// 		for ; true; <-tickerThirtyMinutes.C {
// 			t.ReloadBans()
// 		}
// 	}()

// 	// reload boards every 2 hours
// 	// tickerTwoHours := time.NewTicker(time.Hour * 2)
// 	// go func() {
// 	// 	for ; true; <-tickerTwoHours.C {
// 	// 		t.ReloadBoards()
// 	// 	}
// 	// }()
// }
