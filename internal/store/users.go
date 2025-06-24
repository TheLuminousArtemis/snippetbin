package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Username  string
	Email     string
	Password  password
	CreatedAt time.Time
}

type password struct {
	text *string
	hash []byte
}

func (p *password) Set(plaintext string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintext), 12)
	if err != nil {
		return err
	}
	p.hash = hash
	return nil
}

func (p *password) Compare(plaintext string) error {
	return bcrypt.CompareHashAndPassword(p.hash, []byte(plaintext))
}

type PostgresUserModel struct {
	DB *sql.DB
}

func (m *PostgresUserModel) Insert(ctx context.Context, user *User) error {
	stmt := "INSERT INTO users (username, email, password, created_at) VALUES($1, $2, $3, NOW())"
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	_, err := m.DB.ExecContext(ctx, stmt, user.Username, user.Email, user.Password.hash)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			switch pgErr.Constraint {
			case "users_email_key":
				return ErrDuplicateEmail
			case "users_username_key":
				return ErrDuplicateUsername
			}
		}
		return err
	}
	return nil
}

func (m *PostgresUserModel) GetByEmail(ctx context.Context, email string) (*User, error) {
	// var id int
	// var hashedPassword []byte
	var user User
	stmt := "SELECT id, password FROM users WHERE email=$1"
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, stmt, email).Scan(&user.ID, &user.Password.hash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		} else {
			return nil, err
		}
	}

	// err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	// if err != nil {
	// 	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
	// 		return 0, ErrInvalidCredentials
	// 	} else {
	// 		return 0, err
	// 	}
	// }
	return &user, err
}

func (m *PostgresUserModel) Exists(ctx context.Context, id int) (bool, error) {
	var exists bool

	stmt := "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)"
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(&exists)
	return exists, err
}
