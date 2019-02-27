package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"gitlab.com/noamdb/modernboard/cache"
	"gitlab.com/noamdb/modernboard/repository"
	"gitlab.com/noamdb/modernboard/utils"
)

type BoardsResource struct {
	Repo    *repository.Repository
	BoardsC *cache.BoardsCache
	// UserBoardsC *cache.UserBoardsCache
}

func (rs BoardsResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route(`/`, func(r chi.Router) {
		r.Get(`/`, rs.List)
		r.Group(func(r chi.Router) {
			r.Use(Authorize(rs.Repo, utils.ADMIN))
			r.Post(`/`, rs.Create)
		})
	})

	return r
}
func (rs BoardsResource) ManageRoutes() chi.Router {
	r := chi.NewRouter()

	r.Use(Authorize(rs.Repo, utils.JANITOR))
	r.Get(`/`, rs.ListManage)
	return r

}
func (rs BoardsResource) List(w http.ResponseWriter, r *http.Request) {
	b, exists := rs.BoardsC.GetBoards()
	if exists {
		w.Write(b)
		return
	}
	boards, err := rs.Repo.GetBoards()

	if err != nil {
		log.WithFields(log.Fields{
			"event": "list boards",
			"error": err,
		}).Error("could not get boards from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	b = rs.BoardsC.InsertBoards(boards)
	w.Write(b)
}

func (rs BoardsResource) Create(w http.ResponseWriter, r *http.Request) {
	board := &repository.BoardCreate{}
	err := json.NewDecoder(r.Body).Decode(board)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !board.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = rs.Repo.CreateBoard(*board)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "create board",
			"error": err,
		}).Error("could not create board")
		w.WriteHeader(http.StatusBadRequest)
	}
	rs.BoardsC.Flush()
}

func (rs BoardsResource) ListManage(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(repository.User)

	boards, err := rs.Repo.GetSpecificBoards(user.Boards)

	if err != nil {
		log.WithFields(log.Fields{
			"event": "get managers board list",
			"error": err,
		}).Error("could not retrieve board list from db")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(boards)
}
