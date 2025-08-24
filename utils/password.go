package utils

import (
	"fmt"

	"github.com/matthewhartstonge/argon2"
)

var argon = argon2.DefaultConfig()

func HashPassword(password string) string {
	encoded, err := argon.HashEncoded([]byte(password))
	if err != nil {
		panic(fmt.Sprintf("password hashing error: %v", err))
	}

	return string(encoded)
}

func VerifyPassword(password, hash string) (bool, error) {
	ok, err := argon2.VerifyEncoded([]byte(password), []byte(hash))
	if err != nil {
		return false, err
	}

	return ok, nil
}

func MustVerifyPassword(password, hash string) bool {
	ok, err := VerifyPassword(password, hash)
	if err != nil {
		panic(fmt.Sprintf("password verification error: %v", err))
	}

	return ok
}
