package utils

import (
	"regexp"
	"strings"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9 ]+`)

func GenerateSlug(name string) string {
	// 1. Lowercase the string
	slug := strings.ToLower(name)

	// 2. Remove non-alphanumeric characters (except spaces)
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "")

	// 3. Replace spaces with dashes
	slug = strings.ReplaceAll(slug, " ", "-")

	// 4. Trim dashes from the ends (in case of leading/trailing spaces)
	return strings.Trim(slug, "-")
}
