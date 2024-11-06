package util

import "golang.org/x/exp/rand"

const (
	themLettersNShit = "abcdefghijklmnoprstuwxyzABCDEFGHIJKLMNOPRSTUWXYZ1234567890-_"
)

func RandomString(n int) string {
	if n < 1 {
		return ""
	}

	name := ""

	for range n {
		name += string(themLettersNShit[rand.Intn(len(themLettersNShit))])
	}

	return name
}
