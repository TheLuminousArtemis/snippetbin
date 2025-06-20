package store

import (
	"database/sql"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int
	Username  string
	Email     string
	Password  []byte
	CreatedAt time.Time
}

type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) Insert(username, email, password string) error {
	// log.Printf("data layer username: %s, email: %s, password: %s", username, email, password)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}

	stmt := "INSERT INTO users (username, email, password, created_at) VALUES($1, $2, $3, NOW())"
	_, err = m.DB.Exec(stmt, username, email, hashedPassword)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) Authenticate(email, password string) (int, error) {
	var id int
	var hashedPassword []byte

	stmt := "SELECT id, password FROM users WHERE email=$1"
	err := m.DB.QueryRow(stmt, email).Scan(&id, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}

	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	return id, err
}

func (m *UserModel) Exists(id int) (bool, error) {
	return false, nil
}
