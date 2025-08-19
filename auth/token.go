package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const tokenBytes = 32

// A 32 bit random, hex encoded string.
type Token string

func NewToken() Token {
	token := make([]byte, tokenBytes)
	_, err := rand.Read(token)
	if err != nil {
		panic(fmt.Sprintf("error reading bytes for session token: %v", err))
	}

	return Token(hex.EncodeToString(token))
}
