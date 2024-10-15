package util

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func pointerToString(s string) *string {
	return &s
}

// Make this more secure
func MakeHMACKey() (*string, error) {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)

	if err != nil {
		return nil, fmt.Errorf("error generating a random secret: %v", err)
	}

	return pointerToString(base64.StdEncoding.EncodeToString(secret)), nil
}
