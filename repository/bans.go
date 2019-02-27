package repository

import "errors"

func (r *Repository) BanPoster(bpi BanPosterInsert) (string, error) {
	var IP string

	rows, err := r.db.NamedQuery(`
	INSERT INTO bans (ip, creator_id, reason, created)
	SELECT ip, :creator_id, :reason, current_timestamp
	FROM posts WHERE posts.id=:post_id
	RETURNING ip`, bpi)
	if err != nil {
		return "", err
	}
	if !rows.Next() {
		return "", errors.New("no IP found")
	}
	rows.Scan(&IP)
	rows.Close()
	return IP, err
}

func (r *Repository) BanIP(bi BanInsert) error {
	_, err := r.db.NamedExec(`
	INSERT INTO bans (ip, creator_id, reason, created)
	VALUES (:ip, :creator_id, :reason, current_timestamp)`, bi)
	return err
}

func (r *Repository) GetBans() ([]BanGet, error) {
	var b []BanGet
	err := r.db.Select(&b, `SELECT ip, reason FROM bans`)
	return b, err
}

func (r *Repository) DeleteOldBans(days int) error {
	_, err := r.db.Exec(`
	DELETE FROM bans
	WHERE created < current_date - $1 * interval '1 day'`, days)
	return err
}
