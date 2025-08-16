package util

import (
	"regexp"
)

func IsValidCEP(cep string) bool {
	cep = regexp.MustCompile(`[- ]`).ReplaceAllString(cep, "")
	if len(cep) != 8 {
		return false
	}
	if matched, _ := regexp.MatchString(`^\d{8}$`, cep); !matched {
		return false
	}
	return true
}
