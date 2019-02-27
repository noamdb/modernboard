package controllers

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"strings"

	"gitlab.com/noamdb/modernboard/cache"
	"gitlab.com/noamdb/modernboard/repository"
	"gitlab.com/noamdb/modernboard/utils"
)

// Authorize - Authenticate user from cookie then Authorize
func Authorize(repo *repository.Repository, minAllowed string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			p, err := repo.GetUserDetails(cookie.Value)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !utils.CheckPermission(minAllowed, p.Role) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			ctx := context.WithValue(r.Context(), "user", p)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func paginate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		ctx := context.WithValue(r.Context(), "page", page)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func BlockBanned(bc *cache.BanCache) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "GET" {
				next.ServeHTTP(w, r)
				return
			}
			IP, _, _ := net.SplitHostPort(r.RemoteAddr)

			if _, exists := bc.IsBanned(IP); exists {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORS Set CORS headers
func CORS(origins []string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			origin := strings.ToLower(r.Header.Get("Origin"))
			for _, o := range origins {
				if o == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
			if r.Method == "OPTIONS" {
				w.Header().Set("Access-Control-Allow-Methods", "DELETE")
				w.WriteHeader(http.StatusOK)
			}
			next.ServeHTTP(w, r)
		})
	}
}
