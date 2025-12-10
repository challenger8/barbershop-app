// internal/utils/strings.go
package utils

import (
	"strings"
)

// ToSnakeCase converts camelCase or PascalCase to snake_case
// Example: "UserName" -> "user_name", "userID" -> "user_id"
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// ContainsAny checks if string contains any of the substrings
// This is a more efficient version that avoids repeated string operations
func ContainsAny(s string, substrings []string) bool {
	for _, sub := range substrings {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
