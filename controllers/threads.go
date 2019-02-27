package controllers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
	"gitlab.com/noamdb/modernboard/cache"
	"gitlab.com/noamdb/modernboard/media"
	"gitlab.com/noamdb/modernboard/repository"
	"gitlab.com/noamdb/modernboard/utils"
)

type ThreadsResource struct {
	Repo         *repository.Repository
	ThreadsPageC *cache.ThreadsPageCache
	ThreadCacheC *cache.ThreadCache
}

func (rs ThreadsResource) ThreadsRoutes() chi.Router {
	r := chi.NewRouter()

	r.Group(func(r chi.Router) {
		r.Use(paginate)
		r.Get("/", rs.List)
	})
	r.Post(`/`, rs.Create)
	r.Group(func(r chi.Router) {
		r.Use(Authorize(rs.Repo, utils.JANITOR))
		r.Use(paginate)
		r.Get("/manage", rs.ListManage)
	})

	return r
}
func (rs ThreadsResource) ThreadRoutes() chi.Router {
	r := chi.NewRouter()

	r.Get(`/`, rs.Get)
	r.Group(func(r chi.Router) {
		r.Use(Authorize(rs.Repo, utils.JANITOR))
		r.Get("/manage", rs.GetManage)
		r.Delete("/", rs.Delete)
	})
	r.Group(func(r chi.Router) {
		r.Use(Authorize(rs.Repo, utils.MOD))
		r.Post("/stick", rs.ToggleSticky)
		r.Post("/lock", rs.ToggleLock)
	})

	return r
}

func (rs ThreadsResource) List(w http.ResponseWriter, r *http.Request) {
	boardURI := chi.URLParam(r, "boardURI")
	page := r.Context().Value("page").(int)

	j, exists := rs.ThreadsPageC.GetPage(boardURI, page)
	if exists {
		w.Write(j)
		return
	}
	threads, err := rs.Repo.GetThreads(boardURI, page)

	if err != nil {
		log.WithFields(log.Fields{
			"event":     "list threads",
			"error":     err,
			"board_uri": boardURI,
			"page":      page,
		}).Error("could not retrieve threads")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	j = rs.ThreadsPageC.InsertPage(boardURI, page, threads)
	w.Write(j)
}

func (rs ThreadsResource) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	r.ParseMultipartForm(10 << 20)
	tc := threadCreate{boardURI: chi.URLParam(r, "boardURI"),
		subject:  r.PostFormValue("subject"),
		author:   r.PostFormValue("author"),
		tripcode: r.PostFormValue("tripcode"),
		body:     r.PostFormValue("body")}
	if !tc.valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileName, thumbnailName, err := media.HandleFile(file, 250)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "create thread",
			"error": err,
		}).Error("could not create files")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ti := repository.ThreadInsert{BoardURI: tc.boardURI, Subject: tc.subject}
	if len(handler.Filename) > 50 {
		handler.Filename = handler.Filename[:50]
	}
	html, _ := utils.HTMLAndReplies(tc.body)

	pi := repository.PostInsert{Author: tc.author,
		Tripcode: utils.EncryptString(tc.tripcode),
		Body:     tc.body, BodyHTML: html,
		FileName: fileName, FileOriginalName: handler.Filename,
		ThumbnailName: thumbnailName, Bump: true,
	}
	pi.IP, _, _ = net.SplitHostPort(r.RemoteAddr)

	threadID, err := rs.Repo.CreateThread(ti, pi)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "create thread",
			"error": err,
		}).Error("could not save thread in db")
		media.DeleteFileAndThumbnail(fileName, thumbnailName)
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(struct {
		ID int `json:"id"`
	}{threadID})
}

func (rs ThreadsResource) ListManage(w http.ResponseWriter, r *http.Request) {
	boardURI := chi.URLParam(r, "boardURI")
	threads, err := rs.Repo.GetThreadsManage(boardURI, r.Context().Value("page").(int))

	if err != nil {
		log.WithFields(log.Fields{
			"event":     "list threads for managers",
			"error":     err,
			"board_uri": boardURI,
		}).Error("could not retrieve threads")
		fmt.Println(err.Error())
		return
	}
	json.NewEncoder(w).Encode(threads)
}

func (rs ThreadsResource) Get(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.Atoi(chi.URLParam(r, "threadID"))
	j, exists := rs.ThreadCacheC.GetThread(threadID)
	if exists {
		w.Write(j)
		return
	}
	thread, err := rs.Repo.GetThread(threadID)

	if err != nil {
		log.WithFields(log.Fields{
			"event":     "get thread",
			"error":     err,
			"thread_id": threadID,
		}).Error("could not retrieve thread from db")
		fmt.Println(err.Error())
		return
	}
	j = rs.ThreadCacheC.InsertThread(threadID, thread)
	w.Write(j)
}

func (rs ThreadsResource) GetManage(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.Atoi(chi.URLParam(r, "threadID"))

	thread, err := rs.Repo.GetThreadManage(threadID)

	if err != nil {
		log.WithFields(log.Fields{
			"event":     "get thread for managers",
			"error":     err,
			"thread_id": threadID,
		}).Error("could not retrieve thread from db")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(thread)
}

func (rs ThreadsResource) Delete(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.Atoi(chi.URLParam(r, "threadID"))

	err := rs.Repo.MarkThreadDeleted(threadID,
		r.Context().Value("user").(repository.User).Boards)
	if err != nil {
		log.WithFields(log.Fields{
			"event":     "delete thread",
			"error":     err,
			"thread_id": threadID,
		}).Error("could not delete thread")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (rs ThreadsResource) ToggleSticky(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.Atoi(chi.URLParam(r, "threadID"))

	err := rs.Repo.ToggleSticky(threadID,
		r.Context().Value("user").(repository.User).Boards)
	if err != nil {
		log.WithFields(log.Fields{
			"event":     "toggle sticky",
			"error":     err,
			"thread_id": threadID,
		}).Error("could not toggle sticky")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (rs ThreadsResource) ToggleLock(w http.ResponseWriter, r *http.Request) {
	threadID, _ := strconv.Atoi(chi.URLParam(r, "threadID"))

	err := rs.Repo.ToggleLock(threadID,
		r.Context().Value("user").(repository.User).Boards)
	if err != nil {
		log.WithFields(log.Fields{
			"event":     "toggle lock",
			"error":     err,
			"thread_id": threadID,
		}).Error("could not toggle lock")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
