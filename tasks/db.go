package tasks

import (
	"fmt"
	"time"

	"gitlab.com/noamdb/modernboard/media"
)

func (t Tasks) ClearThreads() {
	files, err := t.Repo.DeleteThreads(15)
	if err != nil {
		fmt.Println("error while deleting threads", err.Error())
		return
	}
	for _, f := range files {
		media.DeleteFileAndThumbnail(f.FileName, f.ThumbnailName)
	}
}

func (t Tasks) ClearPosts() {
	files, err := t.Repo.DeletePosts(15)
	if err != nil {
		fmt.Println("error while deleting posts", err.Error())
		return
	}
	for _, f := range files {
		media.DeleteFileAndThumbnail(f.FileName, f.ThumbnailName)
	}
}

func (t Tasks) ClearSessions() {
	err := t.Repo.DeleteOldSessions(15)
	if err != nil {
		fmt.Println("error while deleting sessions", err.Error())
		return
	}
}

func (t Tasks) ClearBans() {
	err := t.Repo.DeleteOldBans(15)
	if err != nil {
		fmt.Println("error while deleting bans", err.Error())
		return
	}
}

func (t Tasks) RunDBTasks() {
	ticker := time.NewTicker(time.Hour * 24 * 15)
	go func() {
		for ; true; <-ticker.C {
			fmt.Println("running DB tasks")
			t.ClearThreads()
			t.ClearPosts()
			t.ClearSessions()
			t.ClearBans()
		}
	}()
}
