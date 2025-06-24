package store

import (
	"context"
	"database/sql"
	"time"
)

var QueryTimeOutDuration = 5 * time.Second

type Storage struct {
	Snippets interface {
		Insert(context.Context, *Snippet) (int, error)
		Get(context.Context, int) (*Snippet, error)
		// Latest() ([]Snippet, error)
	}
	Users interface {
		Insert(context.Context, *User) error
		GetByEmail(context.Context, string) (*User, error)
		Exists(context.Context, int) (bool, error)
	}
}

func NewPostgresStore(db *sql.DB) Storage {
	return Storage{
		Snippets: &PostgresSnippet{DB: db},
		Users:    &PostgresUserModel{DB: db},
	}
}
