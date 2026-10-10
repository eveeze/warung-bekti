package password_test

import (
	"testing"

	"github.com/eveeze/warung-backend/internal/pkg/password"
)

func TestPasswordHashAndCheck(t *testing.T) {
	pass := "secretPassword123!"
	hash, err := password.Hash(pass)
	if err != nil {
		t.Fatalf("expected no error hashing password: %v", err)
	}
	if hash == "" || hash == pass {
		t.Fatalf("invalid hash returned")
	}

	if err := password.Check(pass, hash); err != nil {
		t.Errorf("expected password check to succeed: %v", err)
	}

	if err := password.Check("wrongPassword", hash); err == nil {
		t.Errorf("expected password check to fail for incorrect password")
	}

	// Bcrypt returns an error for inputs over 72 bytes
	veryLongPassword := string(make([]byte, 100))
	_, errLong := password.Hash(veryLongPassword)
	if errLong == nil {
		t.Logf("expected error for overly long password if bcrypt enforces max length")
	}
}
