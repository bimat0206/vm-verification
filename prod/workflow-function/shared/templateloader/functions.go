package templateloader

import (
	"fmt"
	"strings"
	"text/template"
)

// DefaultFunctions contains all built-in template functions
var DefaultFunctions = template.FuncMap{
	// String functions
	"split":  strings.Split,
	"join":   strings.Join,
	"upper":  strings.ToUpper,
	"lower":  strings.ToLower,
	"trim":   strings.TrimSpace,
	"printf": fmt.Sprintf,
	
	// Math functions
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
	"mul": func(a, b int) int { return a * b },
	"div": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	},
	
	// Comparison functions
	"gt": func(a, b int) bool { return a > b },
	"lt": func(a, b int) bool { return a < b },
	"eq": func(a, b interface{}) bool { return a == b },
	"ne": func(a, b interface{}) bool { return a != b },
	"ge": func(a, b int) bool { return a >= b },
	"le": func(a, b int) bool { return a <= b },
	
	// Array/slice functions
	"index": indexFunction,
	"len":   lenFunction,
	"first": firstFunction,
	"last":  lastFunction,
	
	// Utility functions
	"ordinal": ordinalFunction,
	"repeat":  repeatFunction,
	"default": defaultFunction,
	
	// Template-specific functions
	"formatArray": formatArrayFunction,
	"contains":    containsFunction,

	// Machine structure functions
	"lastRowLabel":  lastRowLabelFunction,
	"maxSlotNumber": maxSlotNumberFunction,
}

// indexFunction safely gets an element from a slice
func indexFunction(arr []string, i int) string {
	if i < 0 || i >= len(arr) {
		return ""
	}
	return arr[i]
}

// lenFunction returns the length of various types
func lenFunction(v interface{}) int {
	switch val := v.(type) {
	case string:
		return len(val)
	case []string:
		return len(val)
	case []interface{}:
		return len(val)
	case map[string]interface{}:
		return len(val)
	default:
		return 0
	}
}

// firstFunction gets the first element from a slice
func firstFunction(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[0]
}

// lastFunction gets the last element from a slice
func lastFunction(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return arr[len(arr)-1]
}

// ordinalFunction converts numbers to ordinal words (1st, 2nd, 3rd, etc.)
func ordinalFunction(num int) string {
	if num <= 0 {
		return fmt.Sprintf("%d", num)
	}
	
	switch num {
	case 1:
		return "first"
	case 2:
		return "second"
	case 3:
		return "third"
	case 4:
		return "fourth"
	case 5:
		return "fifth"
	case 6:
		return "sixth"
	case 7:
		return "seventh"
	case 8:
		return "eighth"
	case 9:
		return "ninth"
	case 10:
		return "tenth"
	case 11:
		return "eleventh"
	case 12:
		return "twelfth"
	default:
		// For numbers beyond 12, use numeric ordinals
		suffix := "th"
		lastDigit := num % 10
		lastTwoDigits := num % 100
		
		// Special cases for 11th, 12th, 13th
		if lastTwoDigits >= 11 && lastTwoDigits <= 13 {
			suffix = "th"
		} else {
			switch lastDigit {
			case 1:
				suffix = "st"
			case 2:
				suffix = "nd"
			case 3:
				suffix = "rd"
			default:
				suffix = "th"
			}
		}
		return fmt.Sprintf("%d%s", num, suffix)
	}
}

// repeatFunction repeats a string n times
func repeatFunction(s string, count int) string {
	if count <= 0 {
		return ""
	}
	return strings.Repeat(s, count)
}

// defaultFunction returns a default value if the input is empty/nil
func defaultFunction(defaultVal interface{}, value interface{}) interface{} {
	if value == nil {
		return defaultVal
	}
	
	switch v := value.(type) {
	case string:
		if v == "" {
			return defaultVal
		}
	case []string:
		if len(v) == 0 {
			return defaultVal
		}
	case map[string]interface{}:
		if len(v) == 0 {
			return defaultVal
		}
	}
	
	return value
}

// formatArrayFunction formats a string array to a comma-separated string
func formatArrayFunction(arr []string) string {
	if len(arr) == 0 {
		return ""
	}
	return strings.Join(arr, ", ")
}

// containsFunction checks if a string contains a substring
func containsFunction(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Additional helper functions for specific template needs

// CapitalizeFunction capitalizes the first letter of a string
func CapitalizeFunction(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}

// TitleFunction converts a string to title case
func TitleFunction(s string) string {
	return strings.Title(strings.ToLower(s))
}

// PluralizeFunctionfn simple pluralization (adds 's' if count != 1)
func PluralizeFunction(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

// RangeFunction creates a slice of integers from start to end
func RangeFunction(start, end int) []int {
	if start > end {
		return []int{}
	}
	
	result := make([]int, end-start+1)
	for i := range result {
		result[i] = start + i
	}
	return result
}

// PadLeftFunction pads a string with spaces to the left
func PadLeftFunction(s string, length int) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(" ", length-len(s))
	return padding + s
}

// PadRightFunction pads a string with spaces to the right
func PadRightFunction(s string, length int) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(" ", length-len(s))
	return s + padding
}

// lastRowLabelFunction returns the last row label based on row count
// For PREVIOUS_VS_CURRENT templates, this provides a default structure
func lastRowLabelFunction(rowCount ...interface{}) string {
	// Default to 5 rows (A-E) if no row count provided
	count := 5

	// If row count is provided, use it
	if len(rowCount) > 0 {
		if rc, ok := rowCount[0].(int); ok && rc > 0 {
			count = rc
		}
	}

	// Convert count to last row label (A=1, B=2, C=3, etc.)
	if count <= 0 {
		return "A"
	}
	if count > 26 {
		count = 26 // Limit to Z
	}

	return string(rune('A' + count - 1))
}

// maxSlotNumberFunction returns the maximum slot number based on column count
// For PREVIOUS_VS_CURRENT templates, this provides a default structure
func maxSlotNumberFunction(columnCount ...interface{}) string {
	// Default to 8 slots (01-08) if no column count provided
	count := 8

	// If column count is provided, use it
	if len(columnCount) > 0 {
		if cc, ok := columnCount[0].(int); ok && cc > 0 {
			count = cc
		}
	}

	// Format as zero-padded number
	return fmt.Sprintf("%02d", count)
}

// RegisterExtraFunctions adds additional convenience functions to the default set
func RegisterExtraFunctions(funcMap template.FuncMap) {
	funcMap["capitalize"] = CapitalizeFunction
	funcMap["title"] = TitleFunction
	funcMap["pluralize"] = PluralizeFunction
	funcMap["range"] = RangeFunction
	funcMap["padLeft"] = PadLeftFunction
	funcMap["padRight"] = PadRightFunction
}