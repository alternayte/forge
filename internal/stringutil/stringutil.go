package stringutil

import (
	"strings"
	"unicode"
)

// Plural naively pluralizes a string.
func Plural(s string) string {
	if strings.HasSuffix(s, "y") {
		return s[:len(s)-1] + "ies"
	}
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") || strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	return s + "s"
}

// Snake converts PascalCase to snake_case.
func Snake(s string) string {
	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := runes[i-1]
			if prev >= 'a' && prev <= 'z' {
				result.WriteRune('_')
			} else if i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z' {
				result.WriteRune('_')
			}
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// Kebab converts PascalCase to kebab-case (e.g., "BlogPost" -> "blog-post").
func Kebab(s string) string {
	if len(s) == 0 {
		return s
	}

	var result strings.Builder
	runes := []rune(s)

	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			prev := runes[i-1]
			if unicode.IsLower(prev) {
				result.WriteRune('-')
			} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
				result.WriteRune('-')
			}
		}
		result.WriteRune(unicode.ToLower(r))
	}

	return result.String()
}

// LowerCamel converts PascalCase to camelCase (e.g., "BlogPost" -> "blogPost").
func LowerCamel(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	return strings.ToLower(string(runes[0:1])) + string(runes[1:])
}

// Pluralize returns singular or plural form based on count.
func Pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}
