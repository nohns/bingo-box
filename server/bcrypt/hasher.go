package bcrypt

import (
	"errors"

	bingo "github.com/nohns/bingo-box/server"
	"golang.org/x/crypto/bcrypt"
)

type Hasher struct{}

// Try to hash the given password.
func (h *Hasher) Hash(passwd string) ([]byte, error) {

	hash, err := bcrypt.GenerateFromPassword([]byte(passwd), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// Try to match the given password against the hash. Returns domain error if passwords does not match
func (h *Hasher) Compare(hash []byte, passwd string) error {

	if err := bcrypt.CompareHashAndPassword(hash, []byte(passwd)); err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return bingo.ErrPasswordMismatch
		default:
			return err
		}
	}

	return nil
}

func NewHasher() *Hasher {
	return &Hasher{}
}
