package tasks

import (
	"gitlab.com/noamdb/modernboard/repository"
)

type Tasks struct {
	Repo *repository.Repository
}

// Run all tasks
// note - tasks with "true; <-ticker.C" will also run immidiatly
func (t Tasks) Run() {
	// go t.RunCacheTasks()
	go t.RunDBTasks()
}
