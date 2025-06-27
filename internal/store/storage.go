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
		GetByID(context.Context, int) (*User, error)
		PasswordUpdate(context.Context, int, string, string) error
	}
}

func NewPostgresStore(db *sql.DB) Storage {
	return Storage{
		Snippets: &PostgresSnippet{DB: db},
		Users:    &PostgresUserModel{DB: db},
	}
}

func withTx(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
