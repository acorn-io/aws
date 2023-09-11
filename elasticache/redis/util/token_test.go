package util

import (
	"strings"
	"testing"
)

func TestGeneratePassword(t *testing.T) {
	pass, err := GenerateToken(24)
	if err != nil {
		t.Fatal(err)
	}

	for _, char := range pass {
		if !strings.Contains(validTokenCharacters, string(char)) {
			t.Errorf("unexpected character: %s", string(char))
		}
	}
}
