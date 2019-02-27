package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/lib/pq"

	"github.com/jmoiron/sqlx/types"
	"gitlab.com/noamdb/modernboard/utils"
)

type NullString struct {
	sql.NullString
}

func (ns *NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}

type ThreadInsert struct {
	BoardURI string `json:"board_uri"`
	Subject  string `json:"subject"`
	IsSticky bool   `json:"is_sticky"`
	IsLocked bool   `json:"is_locked"`
}

type ThreadWithOP struct {
	ID            int       `json:"id"`
	PostID        int       `json:"post_id"`
	Subject       string    `json:"subject"`
	IsLocked      bool      `json:"is_locked"`
	IsSticky      bool      `json:"is_sticky"`
	Author        string    `json:"author"`
	Tripcode      string    `json:"tripcode"`
	BodyHTML      string    `json:"body_html"`
	ThumbnailName string    `json:"thumbnail_name"`
	FileName      string    `json:"file_name"`
	Created       time.Time `json:"created"`
	PostsCount    string    `json:"posts_count"`
	ImagesCount   string    `json:"images_count"`
}

type ThreadManageWithOP struct {
	ThreadWithOP
	AuthorID string         `json:"author_id"`
	Reports  types.JSONText `json:"reports"`
}

type ThreadWithPosts struct {
	Subject  string         `json:"subject"`
	IsLocked bool           `json:"is_locked"`
	IsSticky bool           `json:"is_sticky"`
	Tripcode string         `json:"tripcode"`
	Posts    types.JSONText `json:"posts"`
}

type PostSelect struct {
	ID               int           `json:"id"`
	Author           string        `json:"author"`
	Tripcode         string        `json:"tripcode"`
	BodyHTML         string        `json:"body_html"`
	ThumbnailName    string        `json:"thumbnail_name"`
	FileName         string        `json:"file_name"`
	FileOriginalName string        `json:"file_original_name"`
	Created          time.Time     `json:"created"`
	Replies          pq.Int64Array `json:"replies"`
}

type PostInsert struct {
	ThreadID         int           `json:"thread_id"`
	Author           string        `json:"author"`
	Tripcode         string        `json:"tripcode"`
	Body             string        `json:"body"`
	BodyHTML         string        `json:"body_html"`
	FileName         string        `json:"file_name"`
	FileOriginalName string        `json:"file_original_name"`
	ThumbnailName    string        `json:"thumbnail_name"`
	IP               string        `json:"ip"`
	AuthorID         string        `json:"author_id"`
	Bump             bool          `json:"bump"`
	Created          time.Time     `json:"created"`
	Replies          pq.Int64Array `json:"replies"`
}

type UserLoginGet struct {
	ID       int    `json:"id"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type UserCreate struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func (uc UserCreate) Valid() bool {
	return utils.ValidLength(uc.Name, 1, 20) &&
		utils.ValidLength(uc.Password, 4, 30) &&
		utils.PermissionExists(uc.Role)
}

type UserSelect struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Role    string    `json:"role"`
	Created time.Time `json:"created"`
}

type BoardUser struct {
	UserSelect
}

type Board struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Uri   string `json:"uri"`
}

type BoardCreate struct {
	Title    string `json:"title"`
	Uri      string `json:"uri"`
	Priority int    `json:"priority"`
}

func (bc BoardCreate) Valid() bool {
	return utils.ValidLength(bc.Title, 1, 15) &&
		utils.ValidLength(bc.Uri, 1, 10)
}

type User struct {
	ID       int           `json:"id"`
	Password string        `json:"password"`
	Role     string        `json:"role"`
	Boards   pq.Int64Array `json:"boards"`
}

type TrendingThread struct {
	ID            int    `json:"id"`
	BoardURI      string `json:"board_uri"`
	Subject       string `json:"subject"`
	ThumbnailName string `json:"thumbnail_name"`
}

type ReportInsert struct {
	Reason   string    `json:"reason"`
	PostID   int       `json:"post_id"`
	IP       string    `json:"ip"`
	AuthorID string    `json:"author_id"`
	Created  time.Time `json:"created"`
}

type ReportedPost struct {
	ID               int            `json:"id"`
	ThreadID         int            `json:"thread_id"`
	BoardURI         string         `json:"board_uri"`
	Subject          NullString     `json:"subject"`
	Author           string         `json:"author"`
	AuthorID         string         `json:"author_id"`
	Tripcode         string         `json:"tripcode"`
	BodyHTML         string         `json:"body_html"`
	ThumbnailName    string         `json:"thumbnail_name"`
	FileName         string         `json:"file_name"`
	FileOriginalName string         `json:"file_original_name"`
	Created          time.Time      `json:"created"`
	IsOP             bool           `json:"is_op"`
	Reports          types.JSONText `json:"reports"`
}

type BanInsert struct {
	IP        string `json:"ip"`
	CreatorID int    `json:"creator_id"`
	Reason    string `json:"reason"`
}
type BanPosterInsert struct {
	PostID    int    `json:"post_id"`
	CreatorID int    `json:"creator_id"`
	Reason    string `json:"reason"`
}

type BanGet struct {
	IP     string `json:"ip"`
	Reason string `json:"reason"`
}

type PostFiles struct {
	ThumbnailName string `json:"thumbnail_name"`
	FileName      string `json:"file_name"`
}
