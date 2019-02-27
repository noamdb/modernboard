package controllers

import (
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"gitlab.com/noamdb/modernboard/cache"
	"gitlab.com/noamdb/modernboard/repository"
)

type HomeResource struct {
	Repo             *repository.Repository
	TrendingThreadsC *cache.TrendingThreadsCache
}

func (rs HomeResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route(`/`, func(r chi.Router) {
		r.Get(`/trending`, rs.Trending)
	})

	return r
}

func (rs HomeResource) Trending(w http.ResponseWriter, r *http.Request) {
	j, exists := rs.TrendingThreadsC.GetThreads()
	if exists {
		w.Write(j)
		return
	}
	threads, err := rs.Repo.GetTrendingThreads()

	if err != nil {
		log.WithFields(log.Fields{
			"event": "trending threads",
			"error": err,
		}).Error("could not retrieve threads from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j = rs.TrendingThreadsC.InsertThreads(threads)
	w.Write(j)
}
