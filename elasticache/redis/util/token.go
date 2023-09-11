package util

import (
	"crypto/rand"
	"math/big"
)

const (
	validTokenCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!&#$^<->"
	validIDCharacters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateToken returns a token of n length that is generated using crypto/rand
// may return an error
func GenerateToken(n int) (string, error) {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(validTokenCharacters))))
		if err != nil {
			return "", err
		}
		b[i] = validTokenCharacters[index.Int64()]
	}
	return string(b), nil
}

// GenerateID returns an alphanumeric ID of length n that is generated using crypto/rand
// may return an error
func GenerateID(n int) (string, error) {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(validIDCharacters))))
		if err != nil {
			return "", err
		}
		b[i] = validIDCharacters[index.Int64()]
	}
	return string(b), nil
}
