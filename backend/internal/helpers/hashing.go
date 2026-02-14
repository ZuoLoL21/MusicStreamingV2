package helpers

import (
	"github.com/matthewhartstonge/argon2"
)

func Encode(password string) string {
	argon := argon2.DefaultConfig()

	encoded, err := argon.HashEncoded([]byte(password))
	if err != nil {
		panic(err)
	}

	return string(encoded)
}

func Verify(password string, encodedPassword string) bool {
	ok, err := argon2.VerifyEncoded([]byte(password), []byte(encodedPassword))
	if err != nil {
		panic(err)
	}

	return ok
}
