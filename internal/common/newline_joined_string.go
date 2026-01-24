package common

import "strings"

func NewlineJoinedString[T any](items []T, mapper func(T) string) string {
	s := make([]string, len(items))
	for i, item := range items {
		s[i] = mapper(item)
	}
	return strings.Join(s, "\n")
}
