package utils

import (
	"crypto/rand"
	"golang.org/x/crypto/scrypt"
	"io"
)

const (
	pw_salt_bytes = 32
	pw_hash_bytes = 64
)

func CryptGenSalt() (string, error) {
	salt := make([]byte, pw_salt_bytes)
	_, err := io.ReadFull(rand.Reader, salt)
	return string(salt), err
}

func CryptGenHash(pw, salt string) (string, error) {
	hpw, err := scrypt.Key([]byte(pw), []byte(salt), 32768, 8, 1, pw_hash_bytes)
	if err != nil {
		return "", err
	}
	return string(hpw), err
}
