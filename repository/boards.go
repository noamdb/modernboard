package repository

import (
	"errors"

	"github.com/lib/pq"
)

func (r *Repository) GetBoards() ([]Board, error) {
	var boards []Board
	err := r.db.Select(&boards, `
	SELECT id, uri, title
	FROM boards
	WHERE priority < 1000`)
	return boards, err
}

func (r *Repository) GetSpecificBoards(b pq.Int64Array) ([]Board, error) {
	var boards []Board
	err := r.db.Select(&boards, `
	SELECT id, uri, title
	FROM boards
	WHERE id = ANY($1)`, b)
	return boards, err
}

func (r *Repository) CreateBoard(bc BoardCreate) error {
	tx := r.db.MustBegin()
	var boardID int

	rows, err := tx.NamedQuery(`
	INSERT INTO boards (title, uri, priority)
	VALUES (:title, :uri, :priority)
	RETURNING id`, bc)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !rows.Next() {
		tx.Rollback()
		return errors.New("Failed board insert")
	}
	rows.Scan(&boardID)
	rows.Close()
	_, err = tx.Exec(`
	INSERT INTO users_boards (user_id, board_id, created)
	SELECT id, $1, current_timestamp FROM users WHERE role='admin'`, boardID)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()

	return err
}
