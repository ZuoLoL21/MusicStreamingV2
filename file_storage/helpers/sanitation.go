package helpers

import (
	"fmt"
	"regexp"
)

const uuidRegex = "^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$"

func ValidateUUID(uuid string) bool {
	matched, err := regexp.MatchString(uuidRegex, uuid)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return matched
}
