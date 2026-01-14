package utils

import (
	"regexp"
	"strings"
)

var (
	reservedAliases = map[string]bool{
		"api":      true,
		"health":   true,
		"admin":    true,
		"login":    true,
		"register": true,
		"static":   true,
		"assets":   true,
	}

	aliasPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,30}$`)
)

func IsValidAlias(alias string) bool {
	lowerAlias := strings.ToLower(alias)

	if reservedAliases[lowerAlias] {
		return false
	}

	return aliasPattern.MatchString(alias)
}

func NormalizeAlias(alias string) string {
	return strings.ToLower(strings.TrimSpace(alias))
}
