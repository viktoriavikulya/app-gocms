package authn

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type PasswordHasher interface {
	Hash(secret string) (string, error)
	Verify(secret string, encoded string) bool
}

type Argon2idHasher struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func DefaultPasswordHasher() Argon2idHasher {
	return Argon2idHasher{Memory: 64 * 1024, Iterations: 3, Parallelism: 2, SaltLength: 16, KeyLength: 32}
}

func (h Argon2idHasher) Hash(secret string) (string, error) {
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("secret is required")
	}
	salt := make([]byte, h.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	sum := argon2.IDKey([]byte(secret), salt, h.Iterations, h.Memory, h.Parallelism, h.KeyLength)
	return fmt.Sprintf(
		"argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.Memory,
		h.Iterations,
		h.Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(sum),
	), nil
}

func (h Argon2idHasher) Verify(secret string, encoded string) bool {
	parts := strings.Split(strings.TrimSpace(encoded), "$")
	if len(parts) != 5 || parts[0] != "argon2id" {
		return false
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[2], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}
	sum := argon2.IDKey([]byte(secret), salt, iterations, memory, parallelism, uint32(len(expected)))
	return subtle.ConstantTimeCompare(sum, expected) == 1
}
