package utils

import (
	"strings"
)

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "_")
}

// ToKebabCase converts a string to kebab-case
func ToKebabCase(s string) string {
	words := splitWords(s)
	return strings.Join(words, "-")
}

// Pluralize converts a word to plural form (simple implementation)
func Pluralize(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

// Singularize converts a word to singular form (simple implementation)
func Singularize(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "es") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return s[:len(s)-1]
	}
	return s
}

// splitWords splits a string into words
func splitWords(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}
	}

	var words []string
	var currentWord strings.Builder

	for i, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
			continue
		}

		if i > 0 && isUpperCase(r) && !isUpperCase(rune(s[i-1])) {
			if currentWord.Len() > 0 {
				words = append(words, strings.ToLower(currentWord.String()))
				currentWord.Reset()
			}
		}

		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		words = append(words, strings.ToLower(currentWord.String()))
	}

	return words
}

// isUpperCase checks if a rune is uppercase
func isUpperCase(r rune) bool {
	return r >= 'A' && r <= 'Z'
}
