//go:build test

package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// msgCardIndexRequired and msgInvalidCardIndex render the two card-index
// rejections the way the controllers do.
//
// The tests used to spell these as English literals, which worked only while
// the controllers passed English literals of their own (issue #5384). Going
// through i18n keeps them correct in whichever language the suite runs in, and
// makes them fail if the key is renamed or deleted rather than if the wording
// is polished.
//
// cui_msg_ext_test.go carries the same helpers for the external test package,
// plus the ones only it needs.
func msgCardIndexRequired() string { return i18n.MarkError(i18n.T("cardIndexRequired")) }

func msgInvalidCardIndex(val string) string {
	return i18n.MarkError(i18n.Tf("invalidCardIndex", "val", val))
}

func msgInvalidCpuDifficultyPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidCpuDifficulty", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgCpuDifficultyRequired() string { return i18n.MarkError(i18n.T("cpuDifficultyRequired")) }

func msgBetAmountRequired() string { return i18n.MarkError(i18n.T("betAmountRequired")) }

func msgInvalidBetAmountPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidBetAmount", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgInvalidBetAmount(val string) string {
	return i18n.MarkError(i18n.Tf("invalidBetAmount", "val", val))
}

// msgKey renders any rejection the controllers raise through invalidArg.
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
// assertions that looked for the English word "Invalid": those pass in Japanese
// mode only while the reply is still English, so they test the bug.
func msgRejected(out string) bool {
	_, isErr := i18n.StripErrorPrefix(out)
	return isErr
}
