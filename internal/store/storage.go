package store

import "database/sql"

type Storage struct {
	Snippets interface {
		Insert(string, string, int) (int, error)
		Get(int) (Snippet, error)
		Latest() ([]Snippet, error)
	}
	Users interface {
		Insert(*User) error
		GetByEmail(string) (*User, error)
		Exists(int) (bool, error)
	}
}

func NewPostgresStore(db *sql.DB) Storage {
	return Storage{
		Snippets: &PostgresSnippet{DB: db},
		Users:    &PostgresUserModel{DB: db},
	}
}
