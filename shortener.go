package main

import (
	"math/rand"
	"time"
)

const shortCodeCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateShortCode(length int) string {
	if length <= 0 {
		return ""
	}

	rand.Seed(time.Now().UnixNano())

	code := make([]byte, length)

	for i := range code {
		code[i] = shortCodeCharset[rand.Intn(len(shortCodeCharset))]
	}

	return string(code)
}
