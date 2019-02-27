package main

import (
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gitlab.com/noamdb/modernboard/cache"
	"gitlab.com/noamdb/modernboard/controllers"
	"gitlab.com/noamdb/modernboard/repository"
)

func NewRouter(repo *repository.Repository) *chi.Mux {
	c := &cache.Cache{Repository: repo}
	c.Init()

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(180 * time.Second))
	r.Use(controllers.CORS(viper.GetStringSlice("cors_domains")))
	r.Use(controllers.BlockBanned(c.BanCache))
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("."))
	})

	threadR := controllers.ThreadsResource{repo, c.ThreadsPageCache, c.ThreadCache}
	usersR := controllers.UsersResource{repo}
	boardsR := controllers.BoardsResource{repo, c.BoardsCache}
	r.Mount("/static", controllers.FilesResource{}.Routes())
	r.Mount("/home", controllers.HomeResource{repo, c.TrendingThreadsCache}.Routes())
	r.Mount("/users", usersR.Routes())
	r.Mount("/ban", controllers.BansResource{repo, c.BanCache}.Routes())
	r.Route(`/boards`, func(r chi.Router) {
		r.Mount("/", boardsR.Routes())
		r.Mount("/manage", boardsR.ManageRoutes())
		r.Route(`/{boardURI:[a-zA-z0-9]{1,10}}`, func(r chi.Router) {
			r.Mount("/threads", threadR.ThreadsRoutes())
			r.Mount("/users", usersR.BoardRoutes())
		})
		r.Route(`/threads`, func(r chi.Router) {
			r.Route(`/{threadID:[0-9]+}`, func(r chi.Router) {
				r.Mount(`/posts`, controllers.PostsResource{repo}.ThreadRoutes())
				r.Mount("/", threadR.ThreadRoutes())

			})
			r.Mount(`/posts`, controllers.PostsResource{repo}.PostsRoutes())

		})
	})

	return r
}
