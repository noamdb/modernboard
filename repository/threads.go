package repository

import (
	"errors"

	"github.com/lib/pq"
	"gitlab.com/noamdb/modernboard/utils"
)

// GetThreads returns list of threads with the first post
func (r *Repository) GetThreads(boardURI string, page int) ([]ThreadWithOP, error) {
	var threads []ThreadWithOP
	err := r.db.Select(&threads, `
	SELECT t.id, t.subject, t.is_sticky, t.is_locked, p.id AS post_id, p.author, p.body_html, p.tripcode, p.file_name, p.thumbnail_name, 
	p.created, posts.count - 1 as posts_count, images.count - 1 as images_count
	FROM threads AS t,
	LATERAL
		   (SELECT id, author, body_html, tripcode, file_name, thumbnail_name, created
			 FROM posts AS pos
			 WHERE pos.thread_id = t.id
			 ORDER BY created ASC
			 LIMIT 1) AS p,
	LATERAL
		   (SELECT created
			 FROM posts AS pos
			 WHERE pos.thread_id = t.id AND pos.bump = true
			ORDER BY pos.created DESC
			 LIMIT 1) AS lp,
 	LATERAL
 		   (SELECT id
 			FROM boards AS bo
			 WHERE bo.uri = $1) AS b,
	LATERAL
			 (SELECT COUNT(id)
			  FROM posts
			  WHERE thread_id = t.id) AS posts,
    LATERAL
			 (SELECT COUNT(id)
			  FROM posts
			  WHERE thread_id = t.id AND file_name <> '') AS images
	WHERE b.id = t.board_id AND t.deleted IS NOT true
	ORDER BY t.is_sticky DESC, lp.created DESC
	LIMIT $2 OFFSET $2*($3-1)`, boardURI, pageSize, page)
	return threads, err
}

// GetThread returns single thread with posts
func (r *Repository) GetThread(ThreadID int) (ThreadWithPosts, error) {
	var thread ThreadWithPosts
	err := r.db.Get(&thread, `
	SELECT subject, is_sticky, is_locked, 
	(SELECT json_agg(row_to_json(p))
	FROM (SELECT id, author, body_html, tripcode, file_name, thumbnail_name, file_original_name, 
		created, (SELECT array_agg(r.reply_id) FROM (SELECT reply_id FROM replies WHERE post_id = id) AS r) AS replies
		FROM posts
		WHERE posts.thread_id = t.id AND posts.deleted IS NOT true
		ORDER BY created ASC) AS p
	) AS posts 
	FROM threads AS t
	WHERE t.id = $1 AND t.deleted IS NOT true`, ThreadID)
	return thread, err
}

// CreateThread create a thread and the first post
func (r *Repository) CreateThread(ti ThreadInsert, pi PostInsert) (int, error) {
	tx := r.db.MustBegin()
	var threadID int

	rows, err := tx.NamedQuery(`
	INSERT INTO threads (board_id, subject, is_sticky, is_locked)
	SELECT id, :subject, :is_sticky, :is_locked  FROM boards WHERE uri = :board_uri 
	RETURNING id`, ti)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if !rows.Next() {
		tx.Rollback()
		return 0, errors.New("board not exists")
	}
	rows.Scan(&threadID)
	rows.Close()
	pi.ThreadID = threadID
	pi.AuthorID = utils.EncryptString(pi.IP)

	_, err = tx.NamedExec(`
	INSERT INTO posts (thread_id, author, body, body_html, tripcode, ip, author_id, bump, created, file_name, file_original_name, thumbnail_name)
	VALUES (:thread_id, :author, :body, :body_html, :tripcode, :ip, :author_id, :bump, current_timestamp, :file_name, :file_original_name, :thumbnail_name)`, pi)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	err = tx.Commit()
	return threadID, err
}

func (r *Repository) GetTrendingThreads() ([]TrendingThread, error) {
	var threads []TrendingThread
	err := r.db.Select(&threads, `
	SELECT t.id, t.subject, op.thumbnail_name, b.uri AS board_uri
	FROM threads AS t,
	LATERAL
	(SELECT created
		FROM posts AS pos
		WHERE pos.thread_id = t.id AND pos.deleted IS NOT true
		ORDER BY created DESC
		LIMIT 1) AS new_post,
	LATERAL
	(SELECT thumbnail_name, created
		FROM posts
		WHERE thread_id = t.id
		ORDER BY created ASC
		LIMIT 1) AS op,
     LATERAL 
     (SELECT uri
		FROM boards
		WHERE id = t.board_id) AS b
	WHERE t.deleted IS NOT true
	ORDER BY new_post.created DESC
	LIMIT 6`)
	return threads, err
}

// GetThreadsManage returns list of threads with the first post for managers
func (r *Repository) GetThreadsManage(boardURI string, page int) ([]ThreadManageWithOP, error) {
	var threads []ThreadManageWithOP
	err := r.db.Select(&threads, `
	SELECT t.id, t.subject, t.is_sticky, t.is_locked, p.id AS post_id, p.author, p.body_html, p.tripcode, p.author_id, p.file_name, p.thumbnail_name, 
	p.created, posts.count - 1 as posts_count, images.count - 1 as images_count, p.reports
	FROM threads AS t,
	LATERAL
		   (SELECT id, author, body_html, tripcode, author_id, file_name, thumbnail_name, created,
			(SELECT json_agg(row_to_json(r)) FROM (SELECT id, reason, author_id, created FROM reports WHERE post_id=pos.id AND dismissed=false) AS r) AS reports
			 FROM posts AS pos
			 WHERE pos.thread_id = t.id
			 ORDER BY created ASC
			 LIMIT 1) AS p,
	LATERAL
		   (SELECT created
			 FROM posts AS pos
			 WHERE pos.thread_id = t.id AND pos.bump = true AND pos.deleted IS NOT true
			ORDER BY pos.created DESC
			 LIMIT 1) AS lp,
 	LATERAL
 		   (SELECT id
 			FROM boards AS bo
			 WHERE bo.uri = $1) AS b,
	LATERAL
			 (SELECT COUNT(id)
			  FROM posts
			  WHERE thread_id = t.id) AS posts,
    LATERAL
			 (SELECT COUNT(id)
			  FROM posts
			  WHERE thread_id = t.id AND file_name <> '') AS images
	WHERE b.id = t.board_id AND t.deleted IS NOT true
	ORDER BY t.is_sticky DESC, lp.created DESC
	LIMIT $2 OFFSET $2*($3-1)`, boardURI, pageSize, page)
	return threads, err
}

// GetThreadManage returns single thread with posts for managers
func (r *Repository) GetThreadManage(ThreadID int) (ThreadWithPosts, error) {
	var thread ThreadWithPosts
	err := r.db.Get(&thread, `
	SELECT subject, is_sticky, is_locked, 
	(SELECT json_agg(row_to_json(p))
	FROM (SELECT id, author, body_html, tripcode, author_id, file_name, thumbnail_name, file_original_name, 
		created, (SELECT array_agg(r.reply_id) FROM (SELECT reply_id FROM replies WHERE post_id = posts.id) AS r) AS replies,
		(SELECT json_agg(row_to_json(r)) FROM (SELECT id, reason, author_id, created FROM reports WHERE post_id=posts.id AND dismissed=false) AS r) AS reports
		 FROM posts 
		 WHERE posts.thread_id = t.id AND posts.deleted IS NOT true
		 ORDER BY created ASC) AS p
	) AS posts 
	FROM threads AS t
	WHERE t.id = $1 AND t.deleted IS NOT true`, ThreadID)
	return thread, err
}

// MarkThreadDeleted if user has permission on board
func (r *Repository) MarkThreadDeleted(threadID int, boards pq.Int64Array) error {
	_, err := r.db.Exec(`
			UPDATE threads SET deleted=true
			WHERE board_id=ANY($2) AND id=$1`,
		threadID, boards)
	return err

}

// DeleteThreads delete threads that are marked as deleted and return their posts files
func (r *Repository) DeleteThreads(days int) ([]PostFiles, error) {
	tx := r.db.MustBegin()
	var f []PostFiles
	// Select all files of the posts of the deleted thread
	err := tx.Select(&f, `SELECT file_name, thumbnail_name FROM threads
	INNER JOIN posts ON posts.thread_id=threads.id
	WHERE threads.deleted=true AND file_name <> '' AND
	EXISTS(SELECT created FROM posts AS p
		WHERE p.thread_id=threads.id AND p.created < current_date - $1 * interval '1 day')`,
		days)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	// Delete OLD threads that are marked as deleted or have no posts
	_, err = tx.Exec(`
	DELETE FROM threads
	WHERE (deleted=true AND EXISTS(SELECT created FROM posts AS p 
								   WHERE p.thread_id=threads.id
									AND p.created < current_date - $1 * interval '1 day'))
     	OR ((SELECT count(thread_id) FROM posts WHERE thread_id=threads.id)=0)`, days)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return f, err
}

func (r *Repository) ToggleSticky(threadID int, boards pq.Int64Array) error {
	_, err := r.db.Exec(`
			UPDATE threads SET is_sticky = NOT is_sticky
			WHERE board_id=ANY($2) AND id=$1`,
		threadID, boards)
	return err
}

func (r *Repository) ToggleLock(threadID int, boards pq.Int64Array) error {
	_, err := r.db.Exec(`
			UPDATE threads SET is_locked = NOT is_locked
			WHERE board_id=ANY($2) AND id=$1`,
		threadID, boards)
	return err
}
