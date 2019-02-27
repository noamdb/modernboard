package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"gitlab.com/noamdb/modernboard/cache"
	"gitlab.com/noamdb/modernboard/repository"
	"gitlab.com/noamdb/modernboard/utils"
)

type BansResource struct {
	Repo *repository.Repository
	Bc   *cache.BanCache
}

func (rs BansResource) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(Authorize(rs.Repo, utils.MOD))
	r.Post(`/posts/{postID:[0-9]{1,20}}`, rs.BanPoster)
	r.Post(`/ip`, rs.BanIP)

	return r
}

func (rs BansResource) BanPoster(w http.ResponseWriter, r *http.Request) {
	b := &BanPosterCreate{}
	err := json.NewDecoder(r.Body).Decode(b)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !b.valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	postID, _ := strconv.Atoi(chi.URLParam(r, "postID"))

	bpi := repository.BanPosterInsert{PostID: postID,
		CreatorID: r.Context().Value("user").(repository.User).ID,
		Reason:    b.Reason}
	IP, err := rs.Repo.BanPoster(bpi)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "ban poster",
			"error": err,
		}).Error("could not ban poster", postID)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	go rs.Bc.InsertBan(IP, bpi.Reason)
}

func (rs BansResource) BanIP(w http.ResponseWriter, r *http.Request) {
	b := &BanIPCreate{}
	err := json.NewDecoder(r.Body).Decode(b)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !b.valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bi := repository.BanInsert{IP: b.IP,
		CreatorID: r.Context().Value("user").(repository.User).ID,
		Reason:    b.Reason}
	err = rs.Repo.BanIP(bi)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "ban ip",
			"error": err,
		}).Error("could not ban ip", b)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	go rs.Bc.InsertBan(bi.IP, bi.Reason)
}
