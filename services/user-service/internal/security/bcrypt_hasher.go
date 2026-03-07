package security

import "golang.org/x/crypto/bcrypt"

const defaultBcryptCost = 12

// BcryptHasher implements PasswordHasher with bcrypt.
type BcryptHasher struct{}

func NewBcryptHasher() *BcryptHasher {
	return &BcryptHasher{}
}

func (h *BcryptHasher) Hash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), defaultBcryptCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (h *BcryptHasher) Compare(hash string, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
