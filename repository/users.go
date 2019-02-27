package repository

import "github.com/lib/pq"

func (r *Repository) CreateUser(user UserCreate) error {
	_, err := r.db.NamedExec(`
	INSERT INTO users (name, password, role, created)
	VALUES (:name, :password, :role, current_timestamp)`, user)

	return err
}

func (r *Repository) GetUser(username string) (UserLoginGet, error) {
	var u UserLoginGet
	err := r.db.Get(&u, `
	SELECT id, role, password FROM users
	Where name = $1`, username)
	return u, err
}

func (r *Repository) CreateSession(id string, username string) error {
	_, err := r.db.Exec(`
	INSERT INTO sessions (id, user_id, created)
	SELECT $1, id, current_timestamp FROM users WHERE name = $2`, id, username)
	return err
}

func (r *Repository) GetUserDetails(session_id string) (User, error) {
	var p User
	err := r.db.Get(&p, `	
	SELECT id, role, password, array_remove(array_agg(b.board_id), NULL) AS boards FROM users
	INNER JOIN (SELECT user_id FROM sessions
				WHERE id=$1) AS session
	ON users.id = session.user_id
	LEFT JOIN  (
		SELECT board_id
		FROM users_boards) as b ON user_id = session.user_id
		GROUP BY id`, session_id)
	return p, err
}

func (r *Repository) DeleteSession(session string) error {
	_, err := r.db.Exec(`
	DELETE FROM sessions
	WHERE id = $1`, session)
	return err
}

func (r *Repository) DeleteOldSessions(days int) error {
	_, err := r.db.Exec(`
	DELETE FROM sessions
	WHERE created < current_date - $1 * interval '1 day'`, days)
	return err
}

func (r *Repository) ChangePassword(userID int, newPassword string) error {
	_, err := r.db.Exec(`
			UPDATE users SET password=$1
			WHERE id=$2`, newPassword, userID)
	return err
}

func (r *Repository) GetUsers() ([]UserSelect, error) {
	var u []UserSelect
	err := r.db.Select(&u, `SELECT id, name, role, created FROM users`)
	return u, err
}

func (r *Repository) DeleteUser(userID int) error {
	_, err := r.db.Exec(`DELETE FROM users WHERE id=$1 AND role!='admin'`, userID)
	return err
}

func (r *Repository) GetBoardUsers(boardURI string, boards pq.Int64Array) ([]BoardUser, error) {
	var u []BoardUser
	err := r.db.Select(&u, `
	SELECT users.id, name, role, users_boards.created FROM users
	INNER JOIN boards ON uri=$1 AND boards.id=ANY($2)
	INNER JOIN users_boards ON user_id=users.id AND board_id=boards.id`, boardURI, boards)
	return u, err
}

func (r *Repository) DeleteBoardUser(boardURI string, userID int, boards pq.Int64Array) error {
	_, err := r.db.Exec(`
	DELETE FROM users_boards
	WHERE board_id = (SELECT boards.id FROM boards WHERE boards.uri=$1 AND boards.id=ANY($3))
	AND EXISTS((SELECT users.id FROM users WHERE users.id=$2 AND users.role!='admin'))
	AND user_id=2`, boardURI, userID, boards)
	return err
}

func (r *Repository) CreateBoardUser(uri string, name string, boards pq.Int64Array) error {
	_, err := r.db.Exec(`
	INSERT INTO users_boards (board_id, user_id, created)
	SELECT (SELECT boards.id FROM boards WHERE boards.uri=$1 AND boards.id=ANY($3)),
	(SELECT users.id FROM users WHERE users.name=$2), current_timestamp`,
		uri, name, boards)

	return err
}
