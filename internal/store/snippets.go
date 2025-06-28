package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Snippet struct {
	ID    int64
	Title string
	// Content string
	Ciphertext []byte
	IV         []byte
	Created    time.Time
	Expires    time.Time
}

type PostgresSnippet struct {
	DB *sql.DB
}

func (m *PostgresSnippet) Insert(ctx context.Context, snippet *Snippet) (int, error) {
	// log.Printf("data layer title: %s, content: %s, expires: %d", title, content, expires)
	//	stmt := `INSERT INTO snippets (title, content, created, expires)
	// VALUES ($1, $2, NOW(), NOW() + ($3 || ' days')::INTERVAL)
	stmt := `INSERT INTO snippets (title, content, iv,created, expires)
  VALUES ($1, $2, $3,NOW(), $4)
  RETURNING id
  `
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	var id int
	err := m.DB.QueryRowContext(ctx, stmt, snippet.Title, snippet.Ciphertext, snippet.IV, snippet.Expires).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *PostgresSnippet) Get(ctx context.Context, id int64) (*Snippet, error) {
	stmt := "SELECT id, title, content, iv,created, expires FROM snippets WHERE expires > NOW() and id=$1"
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	row := m.DB.QueryRowContext(ctx, stmt, id)
	var s Snippet
	err := row.Scan(&s.ID, &s.Title, &s.Ciphertext, &s.IV, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	return &s, nil
}

// func (m *PostgresSnippet) Latest() ([]Snippet, error) {
// 	stmt := "SELECT id, title, content, created, expires FROM snippets WHERE expires > NOW() ORDER BY id DESC LIMIT 10"
// 	rows, err := m.DB.Query(stmt)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()
// 	var snippets []Snippet

// 	for rows.Next() {
// 		var s Snippet
// 		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
// 		if err != nil {
// 			return nil, err
// 		}
// 		snippets = append(snippets, s)
// 	}
// 	if err = rows.Err(); err != nil {
// 		return nil, err
// 	}
// 	return snippets, nil
// }
