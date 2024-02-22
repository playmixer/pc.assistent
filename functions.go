package main

import (
	"regexp"
	"strconv"
)

func cleanString(text string, d string) string {
	reg := regexp.MustCompile("[^a-zA-Zа-яА-Я0-9/-]+")
	cleanedString := reg.ReplaceAllString(text, d)
	return cleanedString
}

func isInt(text string) bool {
	_, err := strconv.ParseInt(text, 0, 10)
	return err == nil
}

func strToInt(n string) int {
	v, err := strconv.ParseInt(n, 0, 10)
	if err != nil {
		return 0
	}

	return int(v)
}
