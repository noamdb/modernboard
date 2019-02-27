package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/go-chi/chi"
	"gitlab.com/noamdb/modernboard/repository"
	"golang.org/x/crypto/acme/autocert"
)


func NewServer(mux *chi.Mux) *http.Server {
	srv := http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      mux,
	} 


	return &srv
}

func StartServer(repo *repository.Repository) {
	log.Println("configuring server...")
	api := NewRouter(repo)
	server := NewServer(api)
	if viper.GetString("environment") == "production" {
		log.Println("configuring tls")
		server.Addr = ":443"
		m := autocert.Manager{
			Cache:      autocert.DirCache(viper.GetString("cert_dir")),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(viper.GetString("domain")),
		}
		server.TLSConfig = &tls.Config{
			GetCertificate: m.GetCertificate,
		}
		go http.ListenAndServe(":80", m.HTTPHandler(nil))
		server.ListenAndServeTLS("", "")
	} else {
		server.Addr = ":3333"
		server.ListenAndServe()
	}

}