package cuiutil

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// NoMin is a sentinel value meaning "no lower bound" for ParseIntArg.
const NoMin = math.MinInt

// NoMax is a sentinel value meaning "no upper bound" for ParseIntArg.
const NoMax = math.MaxInt

// ParseIntArg parses args[0] as an integer within [min, max].
// Returns (value, errMsg, ok). If ok is false, errMsg should be returned to the user.
// missingMsg is returned when args is empty.
// invalidMsg is returned when the value is non-numeric or out of range.
// If invalidMsg contains "%s", it is formatted with args[0] via fmt.Sprintf.
// Only a single %s verb is supported; %%s and multiple verbs are not handled.
func ParseIntArg(args []string, missingMsg, invalidMsg string, min, max int) (int, string, bool) {
	if len(args) < 1 {
		return 0, missingMsg, false
	}
	v, err := strconv.Atoi(args[0])
	if err != nil || v < min || v > max {
		msg := invalidMsg
		if strings.Contains(invalidMsg, "%s") {
			msg = fmt.Sprintf(invalidMsg, args[0])
		}
		return 0, msg, false
	}
	return v, "", true
}

// ParseOptionalInt parses args[idx] as an integer, returning defaultVal if absent or invalid.
func ParseOptionalInt(args []string, idx, defaultVal int) int {
	if len(args) <= idx {
		return defaultVal
	}
	v, err := strconv.Atoi(args[idx])
	if err != nil {
		return defaultVal
	}
	return v
}

// ParseIntSlice parses all elements of args as integers, returning skipped values.
func ParseIntSlice(args []string) ([]int, []string) {
	result := make([]int, 0, len(args))
	var skipped []string
	for _, s := range args {
		if v, err := strconv.Atoi(s); err == nil {
			result = append(result, v)
		} else {
			skipped = append(skipped, s)
		}
	}
	return result, skipped
}

// ParseBoundedIntSlice parses all elements of args as integers within [min, max],
// returning skipped or out-of-range values.
func ParseBoundedIntSlice(args []string, min, max int) ([]int, []string) {
	result := make([]int, 0, len(args))
	var skipped []string
	for _, s := range args {
		if v, err := strconv.Atoi(s); err == nil && v >= min && v <= max {
			result = append(result, v)
		} else {
			skipped = append(skipped, s)
		}
	}
	return result, skipped
}

// FormatSkippedWarning returns a warning string for skipped values.
// Returns an empty string if skipped is empty.
func FormatSkippedWarning(skipped []string) string {
	if len(skipped) == 0 {
		return ""
	}
	quoted := make([]string, len(skipped))
	for i, s := range skipped {
		quoted[i] = "'" + s + "'"
	}
	return "\033[33m警告: 無効な値 " + strings.Join(quoted, ", ") + " は無視されました\033[0m"
}
