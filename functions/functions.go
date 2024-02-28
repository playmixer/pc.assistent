package functions

import (
	"regexp"
	"strconv"
	"strings"
)

func CleanString(text string, d string) string {
	reg := regexp.MustCompile("[^a-zA-Zа-яА-Я0-9/-]+")
	cleanedString := reg.ReplaceAllString(text, d)
	return cleanedString
}

func IsInt(text string) bool {
	text = strings.TrimLeft(text, "0")
	_, err := strconv.ParseInt(text, 0, 10)
	return err == nil
}

func StrToInt(n string) int {
	n = strings.TrimLeft(n, "0")
	v, err := strconv.ParseInt(n, 0, 10)
	if err != nil {
		return 0
	}

	return int(v)
}
