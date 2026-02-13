package general

import (
	"regexp"
)

var uuidPattern = regexp.MustCompile("^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")

func ValidateUUID(uuid string) bool {
	return uuidPattern.MatchString(uuid)
}
