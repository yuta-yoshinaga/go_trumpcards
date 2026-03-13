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

// ParseIntSlice parses all elements of args as integers, silently skipping invalid entries.
func ParseIntSlice(args []string) []int {
	result := make([]int, 0, len(args))
	for _, s := range args {
		if v, err := strconv.Atoi(s); err == nil {
			result = append(result, v)
		}
	}
	return result
}

// ParseBoundedIntSlice parses all elements of args as integers within [min, max],
// silently skipping invalid or out-of-range entries.
func ParseBoundedIntSlice(args []string, min, max int) []int {
	result := make([]int, 0, len(args))
	for _, s := range args {
		if v, err := strconv.Atoi(s); err == nil && v >= min && v <= max {
			result = append(result, v)
		}
	}
	return result
}
