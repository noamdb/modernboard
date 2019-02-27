package controllers

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gitlab.com/noamdb/modernboard/media"
	"gitlab.com/noamdb/modernboard/repository"
	"gitlab.com/noamdb/modernboard/utils"
)

type PostsResource struct {
	Repo *repository.Repository
}

func (rs PostsResource) PostsRoutes() chi.Router {
	r := chi.NewRouter()

	r.Route("/{postID:[0-9]+}", func(r chi.Router) {
		r.Get(`/after`, rs.ListAfter)
		r.Post(`/reports`, rs.Report)
		r.Group(func(r chi.Router) {
			r.Use(Authorize(rs.Repo, utils.JANITOR))
			r.Delete(`/`, rs.Delete)
		})
	})

	r.Route("/reports", func(r chi.Router) {
		r.Use(Authorize(rs.Repo, utils.JANITOR))
		r.Route("/", func(r chi.Router) {
			r.Use(paginate)
			r.Get("/", rs.ReportedPosts)
		})
		r.Delete("/{reportID:[0-9]+}", rs.DismissReport)
	})
	return r
}
func (rs PostsResource) ThreadRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post(`/`, rs.Create)
	return r
}

func (rs PostsResource) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
	r.ParseMultipartForm(10 << 20)
	threadID, _ := strconv.Atoi(chi.URLParam(r, "threadID"))
	pc := PostCreate{threadID: threadID,
		author:   r.PostFormValue("author"),
		tripcode: r.PostFormValue("tripcode"),
		body:     r.PostFormValue("body")}
	if !pc.valid() {
		log.WithFields(log.Fields{
			"event": "create post",
		}).Info("inavlid data")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var fileName, thumbnailName, fileOriginalName string
	file, handler, err := r.FormFile("file")
	if err == nil {
		defer file.Close()

		fileName, thumbnailName, err = media.HandleFile(file, 250)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if len(handler.Filename) > 50 {
			handler.Filename = handler.Filename[:50]
		}
		fileOriginalName = handler.Filename
	}
	html, replies := utils.HTMLAndReplies(pc.body)
	pi := repository.PostInsert{ThreadID: pc.threadID, Author: pc.author,
		Tripcode: utils.EncryptString(pc.tripcode),
		Body:     pc.body, BodyHTML: html,
		FileName: fileName, FileOriginalName: fileOriginalName,
		ThumbnailName: thumbnailName, Bump: true,
		Replies: pq.Int64Array(replies),
	}
	pi.IP, _, _ = net.SplitHostPort(r.RemoteAddr)

	err = rs.Repo.CreatePost(pi)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "create post",
			"error": err,
		}).Error("could not create post in db")
		media.DeleteFileAndThumbnail(fileName, thumbnailName)
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(pi)
}

func (rs PostsResource) ListAfter(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(chi.URLParam(r, "postID"))

	posts, err := rs.Repo.GetPostsAfter(postID)

	if err != nil {
		log.WithFields(log.Fields{
			"event": "list after",
			"error": err,
		}).Error("could not get posts after", postID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(posts)
}

func (rs PostsResource) Report(w http.ResponseWriter, r *http.Request) {
	report := &Report{}
	err := json.NewDecoder(r.Body).Decode(report)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "report",
			"error": err,
		}).Error("could not decode report")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !report.valid() {
		log.WithFields(log.Fields{
			"event": "report",
			"error": err,
		}).Error("report is not valid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	postID, _ := strconv.Atoi(chi.URLParam(r, "postID"))

	rep := repository.ReportInsert{Reason: report.Reason, PostID: postID}
	rep.IP, _, _ = net.SplitHostPort(r.RemoteAddr)

	err = rs.Repo.ReportPost(rep)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "report",
			"error": err,
		}).Error("could not save report in db", postID)
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (rs PostsResource) DismissReport(w http.ResponseWriter, r *http.Request) {
	ReportID, _ := strconv.Atoi(chi.URLParam(r, "reportID"))

	rs.Repo.DismissReport(ReportID,
		r.Context().Value("user").(repository.User).Boards)
}

func (rs PostsResource) ReportedPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := rs.Repo.GetReportedPosts(r.Context().Value("user").(repository.User).Boards,
		r.Context().Value("page").(int))
	if err != nil {
		log.WithFields(log.Fields{
			"event": "get reported posts",
			"error": err,
		}).Error("could not get reported posts")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(posts)

}

func (rs PostsResource) Delete(w http.ResponseWriter, r *http.Request) {
	postID, _ := strconv.Atoi(chi.URLParam(r, "postID"))

	err := rs.Repo.MarkPostDeleted(postID,
		r.Context().Value("user").(repository.User).Boards)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "delete post",
			"error": err,
		}).Error("could not delete post", postID)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
