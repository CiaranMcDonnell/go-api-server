package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"runtime"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 47104 // 46 MiB (OWASP recommended)
	argonThreads = 1
	argonKeyLen  = 32
	argonSaltLen = 16
)

// hashSem limits concurrent argon2id operations to CPU count,
// preventing goroutine thrashing and memory storms under load.
var hashSem = make(chan struct{}, runtime.NumCPU())

func HashPassword(password string) (string, error) {
	hashSem <- struct{}{}
	defer func() { <-hashSem }()

	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory, argonTime, argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)

	return encoded, nil
}

func CheckPassword(password, encodedHash string) bool {
	hashSem <- struct{}{}
	defer func() { <-hashSem }()

	salt, hash, err := parseArgon2Hash(encodedHash)
	if err != nil {
		return false
	}

	candidate := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return subtle.ConstantTimeCompare(hash, candidate) == 1
}

func parseArgon2Hash(encoded string) (salt, hash []byte, err error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 {
		return nil, nil, fmt.Errorf("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return nil, nil, fmt.Errorf("unsupported algorithm: %s", parts[1])
	}

	salt, err = base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, fmt.Errorf("decoding salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, fmt.Errorf("decoding hash: %w", err)
	}

	return salt, hash, nil
}
