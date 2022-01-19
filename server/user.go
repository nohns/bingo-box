package bingo

import (
	"context"
	"errors"
	"time"
)

var (
	ErrUserNotFound      = errors.New("bingo: user was not found")
	ErrUserAlreadyExists = errors.New("bingo: user was already exists")
	ErrPasswordMismatch  = errors.New("bingo: passwords are not matching")
)

type UserRepository interface {
	Get(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Save(ctx context.Context, user *User) error
}

type Hasher interface {
	Hash(passwd string) ([]byte, error)
	Compare(hash []byte, passwd string) error
}

type UserService struct {
	userRepo UserRepository
}

// Authenticate user by email and password. If the credentials are valid, a user is returned and otherwise an error.
func (us *UserService) Authenticate(ctx context.Context, email, passwd string) (*User, error) {
	// Try to get user by their email
	u, err := us.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// Instantiate a new user service with a user repoistory.
func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	GameCredits int

	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// See if user object is valid
func (u *User) Validate() error {
	return nil
}

func (u *User) CreateGame(name string) *Game {
	return CreateGame(u.ID, name)
}

// Register user by their information and return a user with a hash password.
func RegisterUser(name, email string) *User {
	return &User{
		Name:      name,
		Email:     email,
		UpdatedAt: time.Now(),
		CreatedAt: time.Now(),
	}
}
