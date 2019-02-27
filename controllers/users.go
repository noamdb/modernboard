package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
	log "github.com/sirupsen/logrus"
	"gitlab.com/noamdb/modernboard/repository"
	"gitlab.com/noamdb/modernboard/utils"
	"golang.org/x/crypto/bcrypt"
)

type UsersResource struct {
	Repo *repository.Repository
}

func (rs *UsersResource) Routes() chi.Router {
	r := chi.NewRouter()
	r.Route(`/register`, func(r chi.Router) {
		r.Use(Authorize(rs.Repo, utils.ADMIN))
		r.Post(`/`, rs.Register)
	})
	r.Route(`/changePassword`, func(r chi.Router) {
		r.Use(Authorize(rs.Repo, utils.JANITOR))
		r.Post(`/`, rs.ChangePassword)
	})
	r.Post(`/login`, rs.Login)
	r.Post(`/logout`, rs.Logout)
	r.Group(func(r chi.Router) {
		r.Use(Authorize(rs.Repo, utils.ADMIN))
		r.Get(`/`, rs.list)
		r.Delete(`/{userID:[0-9]+}`, rs.deleteUser)
	})
	return r
}

func (rs *UsersResource) BoardRoutes() chi.Router {
	r := chi.NewRouter()
	r.Use(Authorize(rs.Repo, utils.ADMIN))
	r.Get(`/`, rs.ListBoardUsers)
	r.Post(`/`, rs.createBoardUser)
	r.Delete(`/{userID:[0-9]+}`, rs.DeleteBoardUser)
	return r
}

func (rs UsersResource) Register(w http.ResponseWriter, r *http.Request) {
	user := &repository.UserCreate{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !user.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	user.Password = string(hashedPassword)
	err = rs.Repo.CreateUser(*user)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "create user",
			"error": err,
		}).Error("could not create user")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (rs UsersResource) Login(w http.ResponseWriter, r *http.Request) {
	user := &UserLogin{}
	err := json.NewDecoder(r.Body).Decode(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !user.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	savedUser, err := rs.Repo.GetUser(user.Name)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(savedUser.Password), []byte(user.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	uuid := uuid.Must(uuid.NewV4()).String()

	err = rs.Repo.CreateSession(uuid, user.Name)
	if err != nil {
		log.WithFields(log.Fields{
			"event": "create session",
			"error": err,
		}).Error("could not create session")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cookie := http.Cookie{Name: "session", Value: uuid, Path: "/",
		Expires: time.Now().Add(10 * 24 * time.Hour), HttpOnly: true}
	http.SetCookie(w, &cookie)
	res := UserLoginResponse{ID: savedUser.ID, Role: savedUser.Role}
	json.NewEncoder(w).Encode(res)

}

func (rs UsersResource) Logout(w http.ResponseWriter, r *http.Request) {
	session, err := r.Cookie("session")
	if err == nil {
		if err := rs.Repo.DeleteSession(session.Value); err != nil {
			log.WithFields(log.Fields{
				"event": "delete session",
				"error": err,
			}).Error("could not delete session")
		}
	}
	cookie := &http.Cookie{Name: "session", Value: "", Expires: time.Unix(0, 0),
		Path: "/", HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

func (rs UsersResource) ChangePassword(w http.ResponseWriter, r *http.Request) {
	cp := &ChangePassword{}
	err := json.NewDecoder(r.Body).Decode(cp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !cp.valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	u := r.Context().Value("user").(repository.User)
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password),
		[]byte(cp.OldPassword)); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cp.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	err = rs.Repo.ChangePassword(u.ID, string(hashedPassword))
	if err != nil {
		log.WithFields(log.Fields{
			"event": "change password",
			"error": err,
		}).Error("could not change password")
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (rs UsersResource) list(w http.ResponseWriter, r *http.Request) {
	users, err := rs.Repo.GetUsers()

	if err != nil {
		log.WithFields(log.Fields{
			"event": "list users",
			"error": err,
		}).Error("could not list users")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (rs UsersResource) deleteUser(w http.ResponseWriter, r *http.Request) {
	deleteID, _ := strconv.Atoi(chi.URLParam(r, "userID"))
	user := r.Context().Value("user").(repository.User)

	if user.ID == deleteID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := rs.Repo.DeleteUser(deleteID)

	if err != nil {
		log.WithFields(log.Fields{
			"event":   "delete user",
			"error":   err,
			"user_id": deleteID,
		}).Error("could not delete user")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func (rs UsersResource) ListBoardUsers(w http.ResponseWriter, r *http.Request) {
	boardURI := chi.URLParam(r, "boardURI")
	user := r.Context().Value("user").(repository.User)

	users, err := rs.Repo.GetBoardUsers(boardURI, user.Boards)

	if err != nil {
		log.WithFields(log.Fields{
			"event":     "list board users",
			"error":     err,
			"board_uri": boardURI,
		}).Error("could not list board users")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(users)
}

func (rs UsersResource) createBoardUser(w http.ResponseWriter, r *http.Request) {
	boardURI := chi.URLParam(r, "boardURI")
	boardUser := &BoardUserCreate{}
	user := r.Context().Value("user").(repository.User)

	err := json.NewDecoder(r.Body).Decode(boardUser)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !boardUser.Valid() {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err = rs.Repo.CreateBoardUser(boardURI, boardUser.Name, user.Boards)

	if err != nil {
		log.WithFields(log.Fields{
			"event":     "create board user",
			"error":     err,
			"board_uri": boardURI,
		}).Error("could not create board user")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
func (rs UsersResource) DeleteBoardUser(w http.ResponseWriter, r *http.Request) {
	boardURI := chi.URLParam(r, "boardURI")
	deleteID, _ := strconv.Atoi(chi.URLParam(r, "userID"))
	user := r.Context().Value("user").(repository.User)

	if user.ID == deleteID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	err := rs.Repo.DeleteBoardUser(boardURI, deleteID, user.Boards)

	if err != nil {
		log.WithFields(log.Fields{
			"event":     "delete board user",
			"error":     err,
			"board_uri": boardURI,
		}).Error("could not delete board user")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}
