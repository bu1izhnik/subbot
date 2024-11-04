package tools

import "strings"

func GetChannelUsername(username string) string {
	if username[0] == '@' {
		return username[1:]
	} else if strings.HasPrefix(username, "t.me/") {
		return username[5:]
	} else if strings.HasPrefix(username, "https://t.me/") {
		return username[13:]
	}
	return username
}
