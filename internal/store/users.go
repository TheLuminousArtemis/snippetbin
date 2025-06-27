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

func (m *PostgresUserModel) GetByID(ctx context.Context, id int) (*User, error) {
	var user User
	stmt := "SELECT username, email, created_at FROM users WHERE id=$1"
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	err := m.DB.QueryRowContext(ctx, stmt, id).Scan(&user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		} else {
			return nil, err
		}
	}
	return &user, nil
}

func (m *PostgresUserModel) PasswordUpdate(ctx context.Context, id int, currentPassword string, newPassword string) error {
	return withTx(ctx, m.DB, func(tx *sql.Tx) error {
		user, err := m.getPasswordByID(ctx, tx, id)
		if err != nil {
			return err
		}

		if err := user.Password.Compare(currentPassword); err != nil {
			return ErrInvalidCredentials
		}

		user.Password.Set(newPassword)
		return m.updatePassword(ctx, tx, user)
	})
}

func (m *PostgresUserModel) getPasswordByID(ctx context.Context, tx *sql.Tx, id int) (*User, error) {
	var user User
	stmt := "SELECT password from users WHERE ID = $1"
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()
	err := tx.QueryRowContext(ctx, stmt, id).Scan(&user.Password.hash)
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (m *PostgresUserModel) updatePassword(ctx context.Context, tx *sql.Tx, u *User) error {
	stmt := "UPDATE users SET password = $1 WHERE id = $2"
	ctx, cancel := context.WithTimeout(ctx, QueryTimeOutDuration)
	defer cancel()

	_, err := tx.ExecContext(ctx, stmt, u.Password.hash, u.ID)
	if err != nil {
		return err
	}

	return nil
}
