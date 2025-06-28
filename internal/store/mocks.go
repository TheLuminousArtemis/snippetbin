package store

import (
	"context"
	"time"
)

func NewStorage() Storage {
	return Storage{
		Snippets: &MockSnippetStore{},
		Users:    &MockUserStore{},
	}
}

type MockSnippetStore struct {
	Snippet Snippet
}

func (m *MockSnippetStore) Insert(ctx context.Context, s *Snippet) (int, error) {
	return 2, nil
}

func (m *MockSnippetStore) Get(ctx context.Context, id int64) (*Snippet, error) {
	switch id {
	case 1:
		return &m.Snippet, nil
	default:
		return nil, ErrNoRecord
	}
}

type MockUserStore struct{}

var MockUser = User{
	ID:        1,
	Username:  "validusername",
	Email:     "valid@example.com",
	CreatedAt: time.Now(),
}

func (m *MockUserStore) Insert(ctx context.Context, u *User) error {
	if u.Username == "duplicateusername" {
		return ErrDuplicateUsername
	}
	if u.Email == "duplicate@example.com" {
		return ErrDuplicateEmail
	}
	return nil
}

func (m *MockUserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	switch email {
	case "valid@example.com":
		return &MockUser, nil
	default:
		return nil, ErrInvalidCredentials
	}
}

func (m *MockUserStore) Exists(ctx context.Context, id int) (bool, error) {
	switch id {
	case 1:
		return true, nil
	default:
		return false, nil
	}
}

func (m *MockUserStore) GetByID(ctx context.Context, id int) (*User, error) {
	switch id {
	case 1:
		return &MockUser, nil
	default:
		return nil, ErrInvalidCredentials
	}
}

func (m *MockUserStore) PasswordUpdate(ctx context.Context, id int, currentPassword string, newPassword string) error {
	return nil
}
