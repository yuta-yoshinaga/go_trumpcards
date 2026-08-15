package cuiutil

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
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

// WithParsedInt はParseIntArgの結果を処理するヘルパー。
// パース失敗時は (errMsg, true) を返し、成功時は fn(value) の結果を (result, true) で返す。
// 戻り値の bool は常に true（コマンドが処理済みであることを示す）。
func WithParsedInt(args []string, missingMsg, invalidMsg string, min, max int, fn func(int) string) (string, bool) {
	v, errMsg, ok := ParseIntArg(args, missingMsg, invalidMsg, min, max)
	if !ok {
		return errMsg, true
	}
	return fn(v), true
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
	return color.Yellow(i18n.Tf("skippedWarning", "values", strings.Join(quoted, ", ")))
}

// PrependSkippedWarning prepends a warning to result if skipped is non-empty.
func PrependSkippedWarning(result string, skipped []string) string {
	if w := FormatSkippedWarning(skipped); w != "" {
		return w + "\n" + result
	}
	return result
}

// ParseIntArgKeys is ParseIntArg with i18n keys instead of message strings.
//
// ParseIntArg takes the two messages as Go literals, and all 588 call sites
// passed English ones -- so a player running in Japanese got a Japanese board,
// Japanese prompts and a Japanese "unknown command", then `Invalid card index:
// zz.` the moment they mistyped an argument (issue #5384). Taking keys instead
// puts these messages through the same i18n path as every other string the CUI
// prints.
//
// invalidKey is rendered with the offending argument as {{val}}; a message that
// does not name the value simply omits the placeholder.
// An empty missingKey means "this call cannot be reached with no argument";
// it yields an empty message rather than i18n.T("")'s literal empty key.
func ParseIntArgKeys(args []string, missingKey, invalidKey string, min, max int) (int, string, bool) {
	if len(args) < 1 {
		if missingKey == "" {
			return 0, "", false
		}
		return 0, i18n.T(missingKey), false
	}
	v, err := strconv.Atoi(args[0])
	if err != nil || v < min || v > max {
		return 0, i18n.Tf(invalidKey, "val", args[0]), false
	}
	return v, "", true
}

// WithParsedIntKeys is WithParsedInt with i18n keys instead of message strings.
// See ParseIntArgKeys.
func WithParsedIntKeys(args []string, missingKey, invalidKey string, min, max int, fn func(int) string) (string, bool) {
	v, errMsg, ok := ParseIntArgKeys(args, missingKey, invalidKey, min, max)
	if !ok {
		return errMsg, true
	}
	return fn(v), true
}
