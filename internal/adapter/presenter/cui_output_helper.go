package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// buildCuiOutput constructs a CUI output string with a standard "==========" header block and footer.
// The output format is:
//
//	==========
//	<title>
//	==========
//	<content from buildContent>
//	==========
func buildCuiOutput(title string, buildContent func(b *strings.Builder)) string {
	var b strings.Builder
	b.WriteString("==========\n")
	b.WriteString(title + "\n")
	b.WriteString("==========\n")
	buildContent(&b)
	b.WriteString("==========\n")
	return b.String()
}

// cuiTrickBlock writes the current trick display block.
// playerIdx extracts the player index from a trick card, cardStr converts the trick card to display string,
// getPlayerName returns the display name for a given player index.
func cuiTrickBlock[TC any](b *strings.Builder, trick []TC, playerIdx func(TC) int, cardStr func(TC) string, getPlayerName func(int) string) {
	if len(trick) == 0 {
		return
	}
	parts := make([]string, len(trick))
	for i, tc := range trick {
		pidx := playerIdx(tc)
		parts[i] = fmt.Sprintf("%s=%s", getPlayerName(pidx), cardStr(tc))
	}
	fmt.Fprintln(b, i18n.Tf("cuiTrickLine", "cards", strings.Join(parts, ", ")))
}

// cuiErrorBlock writes the error line if lastErr is non-nil.
func cuiErrorBlock(b *strings.Builder, lastErr error) {
	if lastErr != nil {
		// Marked per line, not per reply: the reply is a board and belongs on
		// stdout. i18n.StripErrorLines takes the marker off before display.
		fmt.Fprintf(b, "%s\n", i18n.MarkErrorLine(color.Red(lastErr.Error())))
	}
}

// sharedHintReasonKeys maps a hint-reason identifier to the i18n key
// shared across trick-taking games. The displayed string follows the
// active locale via i18n.T (issue #1699 Phase 1).
var sharedHintReasonKeys = map[string]string{
	"follow_suit":    "cuiHintFollowSuit",
	"lead_low":       "cuiHintLeadLow",
	"lead_strong":    "cuiHintLeadStrong",
	"discard_high":   "cuiHintDiscardHigh",
	"trump_cut":      "cuiHintTrumpCut",
	"strategic_bid":  "cuiHintStrategicBid",
	"strategic_bury": "cuiHintStrategicBury",
	"strong_hand":    "cuiHintStrongHand",
	"weak_hand":      "cuiHintWeakHand",
}

// hintReasonStr resolves a hint reason via a game-specific key map first
// (reason → i18n key, translated through i18n.T), then falls back to the
// shared sharedHintReasonKeys, and finally returns the raw reason. It is the
// common implementation behind every per-game *HintReasonStr helper: a game
// passes its xHintReasonKeys map (or nil when it only needs the shared
// fallback). Unlike lookupHintReason, the map values are treated as i18n keys,
// not pre-resolved strings.
func hintReasonStr(reason string, gameKeys map[string]string) string {
	if key, ok := gameKeys[reason]; ok {
		return i18n.T(key)
	}
	if key, ok := sharedHintReasonKeys[reason]; ok {
		return i18n.T(key)
	}
	return reason
}
