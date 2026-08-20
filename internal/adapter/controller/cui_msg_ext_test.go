//go:build test

package controller_test

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// The same helpers as cui_msg_test.go, for the test files in the external
// package. See there for why the assertions go through i18n.
func msgCardIndexRequired() string { return i18n.MarkError(i18n.T("cardIndexRequired")) }

func msgInvalidCardIndex(val string) string {
	return i18n.MarkError(i18n.Tf("invalidCardIndex", "val", val))
}

// msgInvalidCardIndexPrefix is the part of the message before the offending
// value, for the assertions that only check that the rejection happened.
// Split on a sentinel rather than on ": " so it holds in any language.
func msgInvalidCardIndexPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidCardIndex", "val", "\x00"), "\x00", 2)[0]
	// Trim the punctuation that introduces the value, so the stem also matches
	// invalidCardIndexNotANumber -- the assertions using this only care that the
	// index was rejected, not which of the two wordings said so.
	return strings.TrimRight(stem, ":. ")
}

// The two card-index prompts that carry a usage example.
func msgCardIndexRequiredField() string { return i18n.MarkError(i18n.T("cardIndexRequiredField")) }

func msgCardIndexRequiredCapture() string { return i18n.MarkError(i18n.T("cardIndexRequiredCapture")) }

// msgInvalidDiscardIndexPrefix is msgInvalidCardIndexPrefix for the discard pile.
func msgInvalidDiscardIndexPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidDiscardIndex", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

// The CPU-difficulty rejections. The prefix form is for the assertions that
// only check that the value was refused; see msgInvalidCardIndexPrefix.
func msgInvalidCpuDifficulty(val string) string {
	return i18n.MarkError(i18n.Tf("invalidCpuDifficulty", "val", val))
}

func msgInvalidCpuDifficultyPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidCpuDifficulty", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgCpuDifficultyRequired() string { return i18n.MarkError(i18n.T("cpuDifficultyRequired")) }

// msgCpuDifficultyRequiredAlt is the wording used by the games whose difficulty
// scale starts at Normal.
func msgCpuDifficultyRequiredAlt() string { return i18n.MarkError(i18n.T("cpuDifficultyRequiredAlt")) }

// Generated per-family renderers for the #5384 migration; see the note on
// msgCardIndexRequired for why the tests go through i18n at all.
func msgPointLimitRequired() string { return i18n.MarkError(i18n.T("pointLimitRequired")) }

func msgInvalidPointLimit(val string) string {
	return i18n.MarkError(i18n.Tf("invalidPointLimit", "val", val))
}

func msgInvalidPointLimitPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidPointLimit", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgBetAmountRequired() string { return i18n.MarkError(i18n.T("betAmountRequired")) }

func msgInvalidBetAmountPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidBetAmount", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgTargetScoreRequired() string { return i18n.MarkError(i18n.T("targetScoreRequired")) }

func msgInvalidTargetScore(val string) string {
	return i18n.MarkError(i18n.Tf("invalidTargetScore", "val", val))
}

func msgAnteAmountRequired() string { return i18n.MarkError(i18n.T("anteAmountRequired")) }

func msgInvalidAnteAmountPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidAnteAmount", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgInvalidPlayerCount(val string) string {
	return i18n.MarkError(i18n.Tf("invalidPlayerCount", "val", val))
}

func msgInvalidPlayerCountPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidPlayerCount", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgAmountRequired() string { return i18n.MarkError(i18n.T("amountRequired")) }

func msgInvalidAmountPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidAmountNotANumber", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgInvalidBetAmount(val string) string {
	return i18n.MarkError(i18n.Tf("invalidBetAmount", "val", val))
}

// msgKey renders any rejection the controllers raise through invalidArg, and
// msgStem takes the part before the interpolated value and before the hint.
// Assertions have to go through i18n or they pin the English wording, which is
// the bug the suit/bid keys were added to fix -- an assertion on the English
// literal keeps passing in Japanese mode while the player reads English.
func msgKey(key string, params ...string) string {
	return i18n.MarkError(i18n.Tf(key, params...))
}

func msgStem(key string) string {
	stem := strings.SplitN(i18n.Tf(key, "val", "\x00"), "\x00", 2)[0]
	if i := strings.Index(stem, " ("); i >= 0 {
		stem = stem[:i]
	}
	return strings.TrimRight(stem, ":.。 ")
}

// msgRejected reports whether a reply carries the rejection marker. It replaces
// assertions that looked for the English word "required" or "Invalid": those
// pass in Japanese mode only while the reply is still English, so they were
// testing the bug rather than the behaviour.
func msgRejected(out string) bool {
	_, isErr := i18n.StripErrorPrefix(out)
	return isErr
}

// msgUsage renders a usage line the way the controllers do -- through i18n and
// marked as a rejection, because the command did not run.
func msgUsage(key string) string {
	return i18n.MarkError(i18n.T(key))
}
