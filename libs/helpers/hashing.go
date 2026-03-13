package helpers

import (
	"github.com/matthewhartstonge/argon2"
)

// Encode hashes a password using Argon2id with default parameters.
// Takes a plaintext password string and returns the Argon2-encoded hash.
//
// The function uses panic on error since encoding failures are considered fatal.
func Encode(password string) string {
	argon := argon2.DefaultConfig()

	encoded, err := argon.HashEncoded([]byte(password))
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

// Verify compares a plaintext password against an Argon2-encoded hash.
// Takes the plaintext password and the previously encoded hash.
// Returns true if the password matches the hash, false otherwise.
//
// The function uses panic on error since verification failures are considered fatal.
func Verify(password string, encodedPassword string) bool {
	ok, err := argon2.VerifyEncoded([]byte(password), []byte(encodedPassword))
	if err != nil {
		panic(err)
	}

	return ok
}
