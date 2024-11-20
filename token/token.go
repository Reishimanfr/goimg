package token

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateOpaque(n int) (string, error) {
	bytes := make([]byte, n)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
