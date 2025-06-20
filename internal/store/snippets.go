package store

import (
	"database/sql"
	"errors"
	"log"
	"time"
)

type Snippet struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

type PostgresSnippet struct {
	DB *sql.DB
}

func (m *PostgresSnippet) Insert(title string, content string, expires int) (int, error) {
	log.Printf("data layer title: %s, content: %s, expires: %d", title, content, expires)
	stmt := `INSERT INTO snippets (title, content, created, expires)
  VALUES ($1, $2, NOW(), NOW() + ($3 || ' days')::INTERVAL)
  RETURNING id
  `
	var id int
	err := m.DB.QueryRow(stmt, title, content, expires).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (m *PostgresSnippet) Get(id int) (Snippet, error) {
	stmt := "SELECT id, title, content, created, expires FROM snippets WHERE expires > NOW() and id=$1"
	row := m.DB.QueryRow(stmt, id)
	var s Snippet
	err := row.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Snippet{}, ErrNoRecord
		} else {
			return Snippet{}, err
		}
	}
	return s, nil
}

func (m *PostgresSnippet) Latest() ([]Snippet, error) {
	stmt := "SELECT id, title, content, created, expires FROM snippets WHERE expires > NOW() ORDER BY id DESC LIMIT 10"
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var snippets []Snippet

	for rows.Next() {
		var s Snippet
		err = rows.Scan(&s.ID, &s.Title, &s.Content, &s.Created, &s.Expires)
		if err != nil {
			return nil, err
		}
		snippets = append(snippets, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return snippets, nil
}
