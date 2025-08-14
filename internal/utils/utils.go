package utils

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
)

func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func MatchScore(title, searchKey, desc string, useDesc bool) int {
	if title == searchKey {
		return 100
	}
	if strings.HasPrefix(title, searchKey) {
		return 90
	}
	if strings.Contains(title, searchKey) {
		return 70
	}
	if useDesc && strings.Contains(desc, searchKey) {
		return 50
	}
	return 0
}

func GenerateInviteCode(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // handle properly in production
	}
	code := base64.URLEncoding.EncodeToString(b)
	return strings.ToUpper(code[:length])
}
