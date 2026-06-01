package main

import "testing"

func TestGenerateShortCodeLength(t *testing.T) {
	length := 6

	code := generateShortCode(length)

	if len(code) != length {
		t.Fatalf("expected short code length %d, got %d", length, len(code))
	}
}

func TestGenerateShortCodeCharacterSet(t *testing.T) {
	code := generateShortCode(100)

	for _, char := range code {
		if !containsRune(shortCodeCharset, char) {
			t.Fatalf("generated short code contains invalid character %q", char)
		}
	}
}

func TestGenerateShortCodeWithNonPositiveLength(t *testing.T) {
	tests := []int{0, -1}

	for _, length := range tests {
		code := generateShortCode(length)
		if code != "" {
			t.Fatalf("expected empty short code for length %d, got %q", length, code)
		}
	}
}

func containsRune(value string, target rune) bool {
	for _, char := range value {
		if char == target {
			return true
		}
	}

	return false
}
