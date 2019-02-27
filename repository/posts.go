package repository

import (
	"errors"

	"github.com/lib/pq"
	"gitlab.com/noamdb/modernboard/utils"
)

func (r *Repository) CreatePost(pi PostInsert) error {
	tx := r.db.MustBegin()
	var postID int
	pi.AuthorID = utils.EncryptString(pi.IP)
	rows, err := r.db.NamedQuery(`
	INSERT INTO posts (thread_id, author, body, body_html, tripcode, ip, author_id, bump, created, file_name, file_original_name, thumbnail_name)
	VALUES (:thread_id, :author, :body, :body_html, :tripcode, :ip, :author_id, :bump, current_timestamp, :file_name, :file_original_name, :thumbnail_name)
	RETURNING id`, pi)
	if err != nil {
		tx.Rollback()
		return err
	}
	if !rows.Next() {
		tx.Rollback()
		return errors.New("board not exists")
	}
	rows.Scan(&postID)
	rows.Close()

	if len(pi.Replies) > 0 {
		_, err = tx.Exec(`
		INSERT INTO replies (post_id, reply_id)
		SELECT id, $1 FROM posts WHERE id = ANY($2) AND thread_id = $3`, postID, pi.Replies, pi.ThreadID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	return err
}

func (r *Repository) ReportPost(report ReportInsert) error {
	report.AuthorID = utils.EncryptString(report.IP)
	_, err := r.db.NamedExec(`
	INSERT INTO reports (reason, post_id, ip, author_id, created, dismissed)
	VALUES (:reason, :post_id, :ip, :author_id, current_timestamp, false)`, report)
	return err
}

func (r *Repository) GetPostsAfter(postID int) ([]PostSelect, error) {
	var posts []PostSelect
	err := r.db.Select(&posts, `
	SELECT id, author, body_html, tripcode, file_name, thumbnail_name, file_original_name, 
	posts.created, (SELECT array_agg(r.reply_id) FROM (SELECT reply_id FROM replies WHERE post_id = id) AS r) AS replies
	FROM posts,
		LATERAL (SELECT thread_id, created
		FROM posts 
		WHERE id = $1
		ORDER BY created ASC LIMIT 1) AS after_post
	WHERE posts.thread_id=after_post.thread_id AND (posts.created > after_post.created) AND posts.deleted IS NOT true
	ORDER BY created ASC`, postID)
	return posts, err
}

// DismissReport if user has permission on the board
func (r *Repository) DismissReport(reportID int, boards pq.Int64Array) error {
	_, err := r.db.Exec(`
			UPDATE reports SET dismissed=true
			WHERE reports.id=$1
			AND EXISTS(SELECT posts.id FROM posts
            	INNER JOIN threads ON threads.id=posts.thread_id 
           		WHERE posts.id=post_id 
		   		AND board_id=ANY($2))`, reportID, boards)
	return err

}

// GetReportedPosts returns the reported posts for the given boards,
// if post is an OP it will return the thread data as well
func (r *Repository) GetReportedPosts(boards pq.Int64Array, page int) ([]ReportedPost, error) {
	var posts []ReportedPost
	err := r.db.Select(&posts, `
	SELECT p.id, author, body_html, tripcode, author_id, file_name, thumbnail_name, 
	file_original_name, created, thread_id, boards.uri AS board_uri, is_op, reports,
	CASE WHEN is_op = true
	THEN subject
	END AS subject
		FROM(
		SELECT posts.id, posts.thread_id, op.thread_id IS NOT NULL AS is_op, author, body_html, tripcode, author_id, file_name, thumbnail_name, file_original_name, created,
		(SELECT json_agg(row_to_json(r)) FROM (SELECT id, reason, author_id, created FROM reports WHERE post_id=posts.id AND dismissed=false ORDER BY created DESC) AS r) AS reports
		FROM posts
		LEFT JOIN (SELECT DISTINCT ON(thread_id) id, thread_id FROM posts as posts2
				WHERE id=posts2.id
				ORDER BY thread_id, created ASC) AS op ON posts.id = op.id
		WHERE deleted IS NOT true) p
	LEFT JOIN threads ON threads.id=p.thread_id AND threads.deleted IS NOT true
	LEFT JOIN boards ON boards.id=threads.board_id
 	WHERE board_id=ANY($1) AND reports IS NOT NULL
	ORDER BY (reports->0->>'created')::timestamp DESC
		LIMIT $2 OFFSET $2*($3-1)`, boards, pageSize, page)
	return posts, err
}

// MarkPostDeleted if user has permission on board
func (r *Repository) MarkPostDeleted(PostID int, boards pq.Int64Array) error {
	_, err := r.db.Exec(`
			UPDATE posts SET deleted=true
			FROM threads
			WHERE threads.id=posts.thread_id AND board_id=ANY($2) AND posts.id=$1`,
		PostID, boards)
	return err

}

// DeletePosts delete posts that are marked as deleted and return their files
func (r *Repository) DeletePosts(days int) ([]PostFiles, error) {
	tx := r.db.MustBegin()
	var f []PostFiles
	// Select all the files of the posts
	err := tx.Select(&f,
		`SELECT file_name, thumbnail_name FROM posts
		WHERE deleted=true AND created < current_date - $1 * interval '1 day'`,
		days)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// Delete OLD posts that are marked as deleted
	_, err = tx.Exec(`
	DELETE FROM posts
	WHERE deleted=true AND created < current_date - $1 * interval '1 day'`, days)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return f, nil
}
