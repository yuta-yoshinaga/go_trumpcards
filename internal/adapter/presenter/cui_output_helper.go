package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
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
	fmt.Fprintf(b, "トリック: %s\n", strings.Join(parts, ", "))
}

// cuiErrorBlock writes the error line if lastErr is non-nil.
func cuiErrorBlock(b *strings.Builder, lastErr error) {
	if lastErr != nil {
		fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
	}
}

// sharedHintReasons holds translations common across trick-taking games.
var sharedHintReasons = map[string]string{
	"follow_suit":   "リードスートに追随",
	"lead_low":      "低いカードでリード",
	"lead_strong":   "強いカードでリード",
	"discard_high":  "高いカードを捨てる",
	"trump_cut":     "切り札でカット",
	"strategic_bid": "戦略的なビッド",
	"strong_hand":   "強い手札",
	"weak_hand":     "弱い手札",
}

// lookupHintReason looks up a hint reason string from game-specific map, then shared map.
func lookupHintReason(reason string, gameReasons map[string]string) string {
	if s, ok := gameReasons[reason]; ok {
		return s
	}
	if s, ok := sharedHintReasons[reason]; ok {
		return s
	}
	return reason
}
