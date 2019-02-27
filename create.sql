-- \i C:/Users/noam/go/src/gitlab.com/noamdb/affiliator/create.sql


-- drop tables
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
-- drop tables

CREATE TYPE role AS ENUM ('admin', 'mod', 'janitor');

CREATE TABLE IF NOT EXISTS users(
  id   SERIAL PRIMARY KEY NOT NULL,
  name TEXT UNIQUE NOT NULL CONSTRAINT name_check CHECK  (length(name) <= 30),
  password TEXT NOT NULL CONSTRAINT password_check CHECK  (length(password) <= 300),
  role role NOT NULL,
  created TIMESTAMPTZ NOT NULL
  );

CREATE TABLE IF NOT EXISTS sessions(
  id   TEXT PRIMARY KEY NOT NULL,
  user_id INTEGER REFERENCES users ON DELETE CASCADE,
  created TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS boards
(
  id   SERIAL PRIMARY KEY NOT NULL,
  uri TEXT UNIQUE NOT NULL CONSTRAINT uri_check CHECK (length(uri) <= 10),
  title TEXT UNIQUE NOT NULL CONSTRAINT title_check CHECK (length(title) <= 20),
  priority INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS users_boards(
  user_id INTEGER NOT NULL REFERENCES users ON DELETE CASCADE,
  board_id INTEGER NOT NULL REFERENCES boards ON DELETE CASCADE,
  created TIMESTAMPTZ NOT NULL,
  CONSTRAINT users_boards_pkey PRIMARY KEY (user_id, board_id)
);

CREATE TABLE IF NOT EXISTS threads
(
  id SERIAL PRIMARY KEY NOT NULL,
  board_id INTEGER REFERENCES boards ON DELETE CASCADE,
  subject TEXT NOT NULL CONSTRAINT subject_check CHECK (length(subject) <= 300),
  is_sticky boolean NOT NULL,
  is_locked boolean NOT NULL,
  deleted BOOLEAN
);

CREATE TABLE IF NOT EXISTS posts
(
  id SERIAL PRIMARY KEY NOT NULL,
  thread_id INTEGER REFERENCES threads ON DELETE CASCADE,
  author TEXT   NOT NULL CONSTRAINT author_check CHECK (length(author) <= 40),
  body TEXT CONSTRAINT body_check CHECK (length(body) <= 50000),
  body_html TEXT CONSTRAINT body_html_check CHECK (length(body_html) <= 80000),
  tripcode TEXT CONSTRAINT tripcode_check CHECK (length(tripcode) <= 20),
  ip inet NOT NULL,
  author_id TEXT NOT NULL,
  bump boolean NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  file_name TEXT CONSTRAINT file_name_check CHECK (length(file_name) <= 200),
  file_original_name TEXT CONSTRAINT file_original_name_check CHECK  (length(file_original_name) <= 200),
  thumbnail_name TEXT CONSTRAINT thumbnail_name_check CHECK  (length(file_name) <= 200),
  deleted BOOLEAN
  -- size DECIMAL(2, 2) NOT NULL,
);

CREATE TABLE IF NOT EXISTS replies
(
  post_id int REFERENCES posts ON DELETE CASCADE,
  reply_id int REFERENCES posts ON DELETE CASCADE,
  PRIMARY KEY (post_id, reply_id)
);

CREATE TABLE IF NOT EXISTS reports
(
  id  SERIAL PRIMARY KEY NOT NULL,
  post_id INTEGER REFERENCES posts ON DELETE CASCADE,
  reason TEXT NOT NULL,
  ip inet NOT NULL,
  author_id TEXT NOT NULL,
  dismissed BOOLEAN NOT NULL,
  created TIMESTAMPTZ NOT NULL,
  UNIQUE (post_id, ip)
);

CREATE TABLE IF NOT EXISTS bans
(
  id SERIAL PRIMARY KEY NOT NULL,
  ip inet UNIQUE NOT NULL,
  creator_id INTEGER REFERENCES users ON DELETE CASCADE,
  reason TEXT NOT NULL,
  created TIMESTAMPTZ NOT NULL
);


-- default values
-- users
-- INSERT INTO users (name, password, role) VALUES ('admin', 'admin', 'admin');

-- boards
-- INSERT INTO boards (uri, title, priority) VALUES ('b', 'Random', 1);
-- INSERT INTO boards (uri, title, priority) VALUES ('pol', 'Politics', 2);

-- -- threads
-- INSERT INTO threads (board_id, subject, is_sticky, is_locked) VALUES (1, 'this is the first thread', false, false);
-- INSERT INTO threads (board_id, subject, is_sticky, is_locked) VALUES (1, 'woooohooooo second thread', false, false);
-- INSERT INTO threads (board_id, subject, is_sticky, is_locked) VALUES (2, 'first thread in new board', false, false);

-- -- threads
-- INSERT INTO posts (thread_id, author, body, body_html, tripcode, file_name, file_original_name, thumbnail_name, ip, author_id, bump, created) VALUES (1, 'anonymus', 'post 1', 'post 1', '', '1', '1', '1', '8.8.8.8', '1', true, '1.1.2018');
-- INSERT INTO posts (thread_id, author, body, body_html, tripcode, file_name, file_original_name, thumbnail_name, ip, author_id, bump, created) VALUES (1, 'anonymus', 'post 2', 'post 2', '', '2', '2', '2', '8.8.8.8', '1', true, '2.1.2018');
-- INSERT INTO posts (thread_id, author, body, body_html, tripcode, file_name, file_original_name, thumbnail_name, ip, author_id, bump, created) VALUES (2, 'anonymus', 'post 3', 'post 3', '', '3', '3', '3', '8.8.8.8', '1', true, '3.1.2018');
-- INSERT INTO posts (thread_id, author, body, body_html, tripcode, file_name, file_original_name, thumbnail_name, ip, author_id, bump, created) VALUES (3, 'anonymus', 'poll 1', 'post 4', '', '4', '4', '4', '8.8.8.8', '1', true, '3.1.2018');




