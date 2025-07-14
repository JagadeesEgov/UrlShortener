package utils

import (
	"net/url"
	"regexp"
)

var urlRegex = regexp.MustCompile(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`)

func ValidateURL(u string) bool {
	if !urlRegex.MatchString(u) {
		return false
	}
	_, err := url.ParseRequestURI(u)
	return err == nil
} 