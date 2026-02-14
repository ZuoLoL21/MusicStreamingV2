package middleware

import (
	"regexp"
	"strings"
)

type Route struct {
	Route  string
	Method string
}

type RoutePattern struct {
	Pattern *regexp.Regexp
	Method  string
}

func (t RoutePattern) Matches(other Route) bool {
	return t.Pattern.MatchString(other.Route) && t.Method == other.Method
}

var ANY = "[^\\/\\\\]*"
var UUID = "[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}"

var SLASH = "\\/"
var START = "^"
var END = "$"

func BuildRegex(parts ...string) *regexp.Regexp {
	var sb strings.Builder

	sb.WriteString(START)
	for _, part := range parts {
		sb.WriteString(SLASH)
		sb.WriteString(part)
	}
	sb.WriteString(END)

	return regexp.MustCompile(sb.String())
}
