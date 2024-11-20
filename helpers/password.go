package helpers

import (
	"fmt"

	passwordvalidator "github.com/wagslane/go-password-validator"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}
func ComparePassword(hashedPassword string, candidatePassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(candidatePassword))
}
func IsPasswordSecure(pwd string, entropy_threshold float64) (bool, float64, error) {
	entropy := passwordvalidator.GetEntropy(pwd)
	if entropy < entropy_threshold {
		return false, entropy, fmt.Errorf("Choose password that will be harder for others to guess. For example \"SaLam37+>?\"")
	}
	return true, entropy, nil
}
