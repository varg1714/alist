package virtual_file

import "strings"

func AllowUA(userAgent string, allowUaKeyword string) bool {

	split := strings.Split(allowUaKeyword, ",")

	for _, keyword := range split {
		if strings.Contains(userAgent, keyword) {
			return true
		}
	}

	return false

}
