package store

import (
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

func (m *PostgresUserModel) Insert(user *User) error {
	stmt := "INSERT INTO users (username, email, password, created_at) VALUES($1, $2, $3, NOW())"
	_, err := m.DB.Exec(stmt, user.Username, user.Email, user.Password.hash)
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

func (m *PostgresUserModel) GetByEmail(email string) (*User, error) {
	// var id int
	// var hashedPassword []byte
	var user User
	stmt := "SELECT id, password FROM users WHERE email=$1"
	err := m.DB.QueryRow(stmt, email).Scan(&user.ID, &user.Password.hash)
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

func (m *PostgresUserModel) Exists(id int) (bool, error) {
	return false, nil
}
