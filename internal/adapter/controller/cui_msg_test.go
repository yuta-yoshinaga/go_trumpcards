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
func msgCardIndexRequired() string { return i18n.T("cardIndexRequired") }

func msgInvalidCardIndex(val string) string { return i18n.Tf("invalidCardIndex", "val", val) }

func msgInvalidCpuDifficultyPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidCpuDifficulty", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}

func msgCpuDifficultyRequired() string { return i18n.T("cpuDifficultyRequired") }

func msgBetAmountRequired() string { return i18n.T("betAmountRequired") }

func msgInvalidBetAmountPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidBetAmount", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}
