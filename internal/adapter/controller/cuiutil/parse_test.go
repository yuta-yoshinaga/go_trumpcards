package cuiutil_test

import (
	"math"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// --- ParseIntArg ---

func TestParseIntArg_MissingArg(t *testing.T) {
	v, msg, ok := cuiutil.ParseIntArg([]string{}, "missing", "invalid", 0, 10)
	assert.False(t, ok)
	assert.Equal(t, "missing", msg)
	assert.Equal(t, 0, v)
}

func TestParseIntArg_NonNumeric(t *testing.T) {
	v, msg, ok := cuiutil.ParseIntArg([]string{"abc"}, "missing", "invalid", 0, 10)
	assert.False(t, ok)
	assert.Equal(t, "invalid", msg)
	assert.Equal(t, 0, v)
}

func TestParseIntArg_BelowMin(t *testing.T) {
	v, msg, ok := cuiutil.ParseIntArg([]string{"-1"}, "missing", "invalid", 0, 10)
	assert.False(t, ok)
	assert.Equal(t, "invalid", msg)
	assert.Equal(t, 0, v)
}

func TestParseIntArg_AboveMax(t *testing.T) {
	v, msg, ok := cuiutil.ParseIntArg([]string{"11"}, "missing", "invalid", 0, 10)
	assert.False(t, ok)
	assert.Equal(t, "invalid", msg)
	assert.Equal(t, 0, v)
}

func TestParseIntArg_AtMin(t *testing.T) {
	v, msg, ok := cuiutil.ParseIntArg([]string{"0"}, "missing", "invalid", 0, 10)
	assert.True(t, ok)
	assert.Empty(t, msg)
	assert.Equal(t, 0, v)
}

func TestParseIntArg_AtMax(t *testing.T) {
	v, msg, ok := cuiutil.ParseIntArg([]string{"10"}, "missing", "invalid", 0, 10)
	assert.True(t, ok)
	assert.Empty(t, msg)
	assert.Equal(t, 10, v)
}

func TestParseIntArg_InRange(t *testing.T) {
	v, msg, ok := cuiutil.ParseIntArg([]string{"5"}, "missing", "invalid", 0, 10)
	assert.True(t, ok)
	assert.Empty(t, msg)
	assert.Equal(t, 5, v)
}

func TestParseIntArg_InvalidMsgWithFormatVerb(t *testing.T) {
	// When invalidMsg contains %s, it is formatted with args[0].
	v, msg, ok := cuiutil.ParseIntArg([]string{"bad"}, "missing", "Invalid value: %s.", 0, 10)
	assert.False(t, ok)
	assert.Equal(t, "Invalid value: bad.", msg)
	assert.Equal(t, 0, v)
}

func TestParseIntArg_InvalidMsgWithFormatVerbOutOfRange(t *testing.T) {
	// Out-of-range value also uses the format string.
	v, msg, ok := cuiutil.ParseIntArg([]string{"99"}, "missing", "Invalid: %s. Enter 0-10.", 0, 10)
	assert.False(t, ok)
	assert.Equal(t, "Invalid: 99. Enter 0-10.", msg)
	assert.Equal(t, 0, v)
}

func TestParseIntArg_NoLowerBound(t *testing.T) {
	v, _, ok := cuiutil.ParseIntArg([]string{"-9999"}, "missing", "invalid", cuiutil.NoMin, 10)
	assert.True(t, ok)
	assert.Equal(t, -9999, v)
}

func TestParseIntArg_NoUpperBound(t *testing.T) {
	v, _, ok := cuiutil.ParseIntArg([]string{"9999999"}, "missing", "invalid", 0, cuiutil.NoMax)
	assert.True(t, ok)
	assert.Equal(t, 9999999, v)
}

func TestParseIntArg_NoBounds(t *testing.T) {
	v, _, ok := cuiutil.ParseIntArg([]string{"-42"}, "missing", "invalid", cuiutil.NoMin, cuiutil.NoMax)
	assert.True(t, ok)
	assert.Equal(t, -42, v)
}

func TestParseIntArg_ExtraArgsIgnored(t *testing.T) {
	// Only args[0] is read.
	v, _, ok := cuiutil.ParseIntArg([]string{"3", "extra"}, "missing", "invalid", 0, 10)
	assert.True(t, ok)
	assert.Equal(t, 3, v)
}

func TestParseIntArg_NoBoundsNegativeMath(t *testing.T) {
	v, _, ok := cuiutil.ParseIntArg([]string{"-1"}, "m", "i", math.MinInt, math.MaxInt)
	assert.True(t, ok)
	assert.Equal(t, -1, v)
}

// --- WithParsedInt ---

func TestWithParsedInt_ParseFails(t *testing.T) {
	result, cont := cuiutil.WithParsedInt([]string{}, "missing", "invalid", 0, 10, func(v int) string {
		return "should not be called"
	})
	assert.True(t, cont)
	assert.Equal(t, "missing", result)
}

func TestWithParsedInt_ParseSuccess(t *testing.T) {
	result, cont := cuiutil.WithParsedInt([]string{"5"}, "missing", "invalid", 0, 10, func(v int) string {
		return "ok"
	})
	assert.True(t, cont)
	assert.Equal(t, "ok", result)
}

func TestWithParsedInt_OutOfRange(t *testing.T) {
	result, cont := cuiutil.WithParsedInt([]string{"11"}, "missing", "Invalid: %s.", 0, 10, func(v int) string {
		return "should not be called"
	})
	assert.True(t, cont)
	assert.Equal(t, "Invalid: 11.", result)
}

// --- ParseOptionalInt ---

func TestParseOptionalInt_AbsentIdx(t *testing.T) {
	assert.Equal(t, -1, cuiutil.ParseOptionalInt([]string{}, 0, -1))
}

func TestParseOptionalInt_IdxBeyondSlice(t *testing.T) {
	assert.Equal(t, 0, cuiutil.ParseOptionalInt([]string{"5"}, 1, 0))
}

func TestParseOptionalInt_NonNumeric(t *testing.T) {
	assert.Equal(t, -1, cuiutil.ParseOptionalInt([]string{"abc"}, 0, -1))
}

func TestParseOptionalInt_Valid(t *testing.T) {
	assert.Equal(t, 7, cuiutil.ParseOptionalInt([]string{"7"}, 0, -1))
}

func TestParseOptionalInt_NegativeValue(t *testing.T) {
	assert.Equal(t, -5, cuiutil.ParseOptionalInt([]string{"-5"}, 0, 0))
}

func TestParseOptionalInt_IndexOne(t *testing.T) {
	assert.Equal(t, 3, cuiutil.ParseOptionalInt([]string{"1", "3"}, 1, 0))
}

func TestParseOptionalInt_DefaultWhenEmpty(t *testing.T) {
	assert.Equal(t, 42, cuiutil.ParseOptionalInt(nil, 0, 42))
}

// --- ParseIntSlice ---

func TestParseIntSlice_Empty(t *testing.T) {
	result, skipped := cuiutil.ParseIntSlice([]string{})
	assert.Equal(t, []int{}, result)
	assert.Empty(t, skipped)
}

func TestParseIntSlice_AllValid(t *testing.T) {
	result, skipped := cuiutil.ParseIntSlice([]string{"1", "2", "3"})
	assert.Equal(t, []int{1, 2, 3}, result)
	assert.Empty(t, skipped)
}

func TestParseIntSlice_SkipsInvalid(t *testing.T) {
	result, skipped := cuiutil.ParseIntSlice([]string{"1", "abc", "3"})
	assert.Equal(t, []int{1, 3}, result)
	assert.Equal(t, []string{"abc"}, skipped)
}

func TestParseIntSlice_AllInvalid(t *testing.T) {
	result, skipped := cuiutil.ParseIntSlice([]string{"x", "y"})
	assert.Equal(t, []int{}, result)
	assert.Equal(t, []string{"x", "y"}, skipped)
}

func TestParseIntSlice_NegativeValues(t *testing.T) {
	result, skipped := cuiutil.ParseIntSlice([]string{"-1", "2", "-3"})
	assert.Equal(t, []int{-1, 2, -3}, result)
	assert.Empty(t, skipped)
}

func TestParseIntSlice_Nil(t *testing.T) {
	result, skipped := cuiutil.ParseIntSlice(nil)
	assert.Equal(t, []int{}, result)
	assert.Empty(t, skipped)
}

// --- ParseBoundedIntSlice ---

func TestParseBoundedIntSlice_Empty(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice([]string{}, 0, 4)
	assert.Equal(t, []int{}, result)
	assert.Empty(t, skipped)
}

func TestParseBoundedIntSlice_AllInRange(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice([]string{"0", "2", "4"}, 0, 4)
	assert.Equal(t, []int{0, 2, 4}, result)
	assert.Empty(t, skipped)
}

func TestParseBoundedIntSlice_SkipsBelowMin(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice([]string{"-1", "1"}, 0, 4)
	assert.Equal(t, []int{1}, result)
	assert.Equal(t, []string{"-1"}, skipped)
}

func TestParseBoundedIntSlice_SkipsAboveMax(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice([]string{"5", "3"}, 0, 4)
	assert.Equal(t, []int{3}, result)
	assert.Equal(t, []string{"5"}, skipped)
}

func TestParseBoundedIntSlice_SkipsNonNumeric(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice([]string{"abc", "2"}, 0, 4)
	assert.Equal(t, []int{2}, result)
	assert.Equal(t, []string{"abc"}, skipped)
}

func TestParseBoundedIntSlice_AllInvalid(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice([]string{"abc", "-1", "5", "99"}, 0, 4)
	assert.Equal(t, []int{}, result)
	assert.Equal(t, []string{"abc", "-1", "5", "99"}, skipped)
}

func TestParseBoundedIntSlice_AtBoundaries(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice([]string{"0", "4"}, 0, 4)
	assert.Equal(t, []int{0, 4}, result)
	assert.Empty(t, skipped)
}

func TestParseBoundedIntSlice_Nil(t *testing.T) {
	result, skipped := cuiutil.ParseBoundedIntSlice(nil, 0, 4)
	assert.Equal(t, []int{}, result)
	assert.Empty(t, skipped)
}

// --- FormatSkippedWarning ---

func TestFormatSkippedWarning_Empty(t *testing.T) {
	assert.Equal(t, "", cuiutil.FormatSkippedWarning(nil))
	assert.Equal(t, "", cuiutil.FormatSkippedWarning([]string{}))
}

func TestFormatSkippedWarning_Single(t *testing.T) {
	result := cuiutil.FormatSkippedWarning([]string{"abc"})
	assert.Equal(t, "\033[33m警告: 無効な値 'abc' は無視されました\033[0m", result)
}

func TestFormatSkippedWarning_Multiple(t *testing.T) {
	result := cuiutil.FormatSkippedWarning([]string{"abc", "5"})
	assert.Equal(t, "\033[33m警告: 無効な値 'abc', '5' は無視されました\033[0m", result)
}

// --- PrependSkippedWarning ---

func TestPrependSkippedWarning_NoSkipped(t *testing.T) {
	assert.Equal(t, "result", cuiutil.PrependSkippedWarning("result", nil))
	assert.Equal(t, "result", cuiutil.PrependSkippedWarning("result", []string{}))
}

func TestPrependSkippedWarning_WithSkipped(t *testing.T) {
	result := cuiutil.PrependSkippedWarning("game output", []string{"abc"})
	assert.Contains(t, result, "'abc'")
	assert.Contains(t, result, "game output")
	assert.True(t, strings.Index(result, "'abc'") < strings.Index(result, "game output"))
}

// ParseOptionalIntKeys は「省略」と「打ち間違い」を区別する。既定値に差し替える
// 実装では `p abc` が 0 番を出してしまい、プレイヤーが選んでいない手が通る。
func TestParseOptionalIntKeys(t *testing.T) {
	t.Run("absent takes the default", func(t *testing.T) {
		v, msg, ok := cuiutil.ParseOptionalIntKeys(nil, 0, 7, "invalidCardIndex")
		assert.True(t, ok)
		assert.Equal(t, 7, v)
		assert.Empty(t, msg)
	})
	t.Run("index past the end takes the default", func(t *testing.T) {
		v, _, ok := cuiutil.ParseOptionalIntKeys([]string{"0"}, 1, -1, "invalidCardIndex")
		assert.True(t, ok)
		assert.Equal(t, -1, v)
	})
	t.Run("a valid argument wins over the default", func(t *testing.T) {
		v, _, ok := cuiutil.ParseOptionalIntKeys([]string{"3"}, 0, 7, "invalidCardIndex")
		assert.True(t, ok)
		assert.Equal(t, 3, v)
	})
	t.Run("a typo is refused, not defaulted", func(t *testing.T) {
		v, msg, ok := cuiutil.ParseOptionalIntKeys([]string{"abc"}, 0, 7, "invalidCardIndex")
		assert.False(t, ok)
		assert.Zero(t, v)
		body, isErr := i18n.StripErrorPrefix(msg)
		assert.True(t, isErr, "the refusal has to be marked or callers cannot tell it from output")
		assert.Contains(t, body, "abc")
	})
}
