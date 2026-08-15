//go:build test

package controller_test

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// The same helpers as cui_msg_test.go, for the test files in the external
// package. See there for why the assertions go through i18n.
func msgCardIndexRequired() string { return i18n.T("cardIndexRequired") }

func msgInvalidCardIndex(val string) string { return i18n.Tf("invalidCardIndex", "val", val) }

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
func msgCardIndexRequiredField() string { return i18n.T("cardIndexRequiredField") }

func msgCardIndexRequiredCapture() string { return i18n.T("cardIndexRequiredCapture") }

// msgInvalidDiscardIndexPrefix is msgInvalidCardIndexPrefix for the discard pile.
func msgInvalidDiscardIndexPrefix() string {
	stem := strings.SplitN(i18n.Tf("invalidDiscardIndex", "val", "\x00"), "\x00", 2)[0]
	return strings.TrimRight(stem, ":. ")
}
