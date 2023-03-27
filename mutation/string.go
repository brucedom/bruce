package mutation

import (
	"regexp"
	"strings"
)

// StripNonAlnum removes non-alphanumeric characters for a string, useful for filenames without spaces / punctuation.
func StripNonAlnum(str string) string {
	j := 0
	s := []byte(str)
	for _, b := range s {
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') {
			s[j] = b
			j++
		}
	}
	return string(s[:j])
}

// StripExtraWhitespace is a utility function to remove extra white space from a string.
func StripExtraWhitespace(str string) string {
	space := regexp.MustCompile(`\s+`)
	return space.ReplaceAllString(str, " ")
}

// StripExtraWhitespaceFB is a utility function to remove extra white space from a string and also strip starting and ending spaces.
func StripExtraWhitespaceFB(str string) string {
	return strings.TrimLeft(strings.TrimRight(StripExtraWhitespace(str), " "), " ")
}
