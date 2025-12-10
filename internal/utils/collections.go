// internal/utils/collections.go
package utils

// BuildStringSet creates a map[string]bool from a slice for O(1) lookup
// Useful for checking if a value exists in a set
func BuildStringSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, item := range items {
		set[item] = true
	}
	return set
}

// StringInSlice checks if a string exists in a slice
func StringInSlice(s string, slice []string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
