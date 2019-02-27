package controllers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/spf13/viper"
)

type FilesResource struct {
}

var fs http.Handler

func (rs FilesResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route("/*", func(r chi.Router) {
		r.Get("/*", rs.Get)
	})

	fs = http.StripPrefix("/static/", http.FileServer(http.Dir(viper.GetString("static_path"))))

	return r
}

// TODO: images not showing when opening in new tab
func (rs FilesResource) Get(w http.ResponseWriter, r *http.Request) {
	file := strings.Split(r.URL.Path, "/")

	// default FileServer behavior is to list the files in the directory.
	// we dont want this so we make sure the path is not a folder
	// folder names are short but file names are long strings
	if len(file[len(file)-1]) < 15 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// time.Sleep(2 * time.Second)

	fs.ServeHTTP(w, r)
}
