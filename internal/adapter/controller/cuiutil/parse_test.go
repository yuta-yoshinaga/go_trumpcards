package cuiutil_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
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
	assert.Equal(t, []int{}, cuiutil.ParseIntSlice([]string{}))
}

func TestParseIntSlice_AllValid(t *testing.T) {
	assert.Equal(t, []int{1, 2, 3}, cuiutil.ParseIntSlice([]string{"1", "2", "3"}))
}

func TestParseIntSlice_SkipsInvalid(t *testing.T) {
	assert.Equal(t, []int{1, 3}, cuiutil.ParseIntSlice([]string{"1", "abc", "3"}))
}

func TestParseIntSlice_AllInvalid(t *testing.T) {
	assert.Equal(t, []int{}, cuiutil.ParseIntSlice([]string{"x", "y"}))
}

func TestParseIntSlice_NegativeValues(t *testing.T) {
	assert.Equal(t, []int{-1, 2, -3}, cuiutil.ParseIntSlice([]string{"-1", "2", "-3"}))
}

func TestParseIntSlice_Nil(t *testing.T) {
	assert.Equal(t, []int{}, cuiutil.ParseIntSlice(nil))
}

// --- ParseBoundedIntSlice ---

func TestParseBoundedIntSlice_Empty(t *testing.T) {
	assert.Equal(t, []int{}, cuiutil.ParseBoundedIntSlice([]string{}, 0, 4))
}

func TestParseBoundedIntSlice_AllInRange(t *testing.T) {
	assert.Equal(t, []int{0, 2, 4}, cuiutil.ParseBoundedIntSlice([]string{"0", "2", "4"}, 0, 4))
}

func TestParseBoundedIntSlice_SkipsBelowMin(t *testing.T) {
	assert.Equal(t, []int{1}, cuiutil.ParseBoundedIntSlice([]string{"-1", "1"}, 0, 4))
}

func TestParseBoundedIntSlice_SkipsAboveMax(t *testing.T) {
	assert.Equal(t, []int{3}, cuiutil.ParseBoundedIntSlice([]string{"5", "3"}, 0, 4))
}

func TestParseBoundedIntSlice_SkipsNonNumeric(t *testing.T) {
	assert.Equal(t, []int{2}, cuiutil.ParseBoundedIntSlice([]string{"abc", "2"}, 0, 4))
}

func TestParseBoundedIntSlice_AllInvalid(t *testing.T) {
	assert.Equal(t, []int{}, cuiutil.ParseBoundedIntSlice([]string{"abc", "-1", "5", "99"}, 0, 4))
}

func TestParseBoundedIntSlice_AtBoundaries(t *testing.T) {
	assert.Equal(t, []int{0, 4}, cuiutil.ParseBoundedIntSlice([]string{"0", "4"}, 0, 4))
}

func TestParseBoundedIntSlice_Nil(t *testing.T) {
	assert.Equal(t, []int{}, cuiutil.ParseBoundedIntSlice(nil, 0, 4))
}
