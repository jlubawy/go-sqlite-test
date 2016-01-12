package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	sqlite3 "github.com/mattn/go-sqlite3"
)

const createUsersTableStr = `CREATE TABLE IF NOT EXISTS users (
    id   INTEGER PRIMARY KEY ASC,
    name TEXT UNIQUE
)`

const createCommentsTableStr = `CREATE TABLE IF NOT EXISTS comments (
    id      INTEGER PRIMARY KEY ASC,
    user_id INTEGER,
    date    DATETIME,
    comment TEXT,

    FOREIGN KEY(user_id) REFERENCES users(id)
)`

const insertUserStr = "INSERT INTO users (id, name) VALUES (NULL, $1)"
const insertCommentStr = "INSERT INTO comments (id, user_id, date, comment) VALUES (NULL, (SELECT id FROM users WHERE name=$1), $2, $3)"

const selectCommentStr = `SELECT
	comments.id, users.name, comments.date, comments.comment
		FROM comments
		INNER JOIN users
			ON user_id = users.id
		WHERE users.name = $1`

var usernames = []string{
	"jlubawy",
	"anonymous",
}

var commentDesc = []struct {
	Username string
	Comment  string
}{
	{"jlubawy", "Comment 0"},
	{"anonymous", "Comment 1"},
	{"jlubawy", "Comment 2"},
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Open connection
	db, err := sql.Open("sqlite3", "test.db")
	if err != nil {
		log.Fatal(err)
	}

	// Ping connection
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Create and initialize users table
	{
		if _, err = db.Exec(createUsersTableStr, nil); err != nil {
			log.Fatal(err)
		}

		userInsertStmt, err := db.Prepare(insertUserStr)
		if err != nil {
			log.Fatal(err)
		}

		for _, username := range usernames {
			_, err := userInsertStmt.Exec(username)

			if err != nil {
				switch t := err.(type) {
				default:
					log.Fatal(t)
				case sqlite3.Error:
					switch t.Code {
					default:
						log.Fatal(t)
					case sqlite3.ErrConstraint:
						// Ignore username constraint errors
					}
				}
			}
		}

		userInsertStmt.Close()
	}

	// Create and initialize comments table
	{

		if _, err = db.Exec(createCommentsTableStr, nil); err != nil {
			log.Fatal(err)
		}

		commentInsertStmt, err := db.Prepare(insertCommentStr)
		if err != nil {
			log.Fatal(err)
		}

		for _, desc := range commentDesc {
			if _, err := commentInsertStmt.Exec(desc.Username, time.Now().UTC(), desc.Comment); err != nil {
				log.Fatal(err)
			}

			time.Sleep(1 * time.Second)
		}

		commentInsertStmt.Close()
	}

	// Select all comments from jlubawy username
	{
		rows, err := db.Query(selectCommentStr, "jlubawy")
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var commentID int
			var username string
			var date time.Time
			var comment string

			if err := rows.Scan(&commentID, &username, &date, &comment); err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%d) User '%s' said '%s' at '%s'\n", commentID, username, comment, date)
		}

		// Check for query errors
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}
	}

	// Close connection
	if err = db.Close(); err != nil {
		log.Fatal(err)
	}
}
