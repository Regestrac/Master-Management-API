package utils

import "strings"

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
