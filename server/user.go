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
	hasher   Hasher
	userRepo UserRepository
}

// Register and persist user with a hashed password
func (us *UserService) Register(ctx context.Context, name, email, passwd string) (*User, error) {

	// Create hash of password
	hash, err := us.hasher.Hash(passwd)
	if err != nil {
		return nil, err
	}

	// Register user and get the user model
	u := RegisterUser(name, email, hash)

	// Try to persist the user
	err = us.userRepo.Save(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// Authenticate user by email and password. If the credentials are valid, a user is returned and otherwise an error.
func (us *UserService) Authenticate(ctx context.Context, email, passwd string) (*User, error) {
	// Try to get user by their email
	u, err := us.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	// Match password of user found by email
	if err := us.hasher.Compare(u.HashedPassword, passwd); err != nil {
		return nil, err
	}

	return u, nil
}

// Instantiate a new user service with a user repoistory.
func NewUserService(userRepo UserRepository, hasher Hasher) *UserService {
	return &UserService{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`

	HashedPassword []byte `json:"-"`

	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

// See if user object is valid
func (u *User) Validate() {

}

func (u *User) CreateGame(name string) *Game {
	return CreateGame(u.ID, name)
}

// Register user by their information and return a user with a hash password.
func RegisterUser(name, email string, hashedPasswd []byte) *User {
	return &User{
		Name:           name,
		Email:          email,
		HashedPassword: hashedPasswd,
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now(),
	}
}
