package presenter

import (
	"math"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// cuiCountPileCards sums the cards held across every given pile.
//
// The solitaire domains expose their foundations as fixed-size arrays whose
// length differs per game ([4][]*Card, [8][]*Card, ...). Those are distinct Go
// types, so callers slice the array (f[:]...) and the counting itself lives
// here once instead of being re-spelled in every presenter.
func cuiCountPileCards(piles ...[]*domain.Card) int {
	n := 0
	for _, p := range piles {
		n += len(p)
	}
	return n
}

// cuiSolitaireProgressPercent converts a foundation count into the whole-number
// percentage the web pages show, matching their Math.round semantics. A total of
// zero or less yields 0 rather than dividing by zero.
func cuiSolitaireProgressPercent(count, total int) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(float64(count) / float64(total) * 100))
}

// cuiSolitaireGameOverSummary renders the "reached N of total cards (P%)" line
// that the web pages display at game over, so a CUI player learns how close the
// deal came instead of only that it ended.
func cuiSolitaireGameOverSummary(count, total int) string {
	return i18n.Tf("cuiSolitaireGameOverSummary",
		"count", strconv.Itoa(count),
		"total", strconv.Itoa(total),
		"percent", strconv.Itoa(cuiSolitaireProgressPercent(count, total)))
}

// cuiSolitaireGameOverFaces is the Grandfather's Clock variant of
// cuiSolitaireGameOverSummary: that game's web page reports completed clock
// faces rather than a card count, so a percentage would be meaningless.
func cuiSolitaireGameOverFaces(count, total int) string {
	return i18n.Tf("cuiSolitaireGameOverFaces",
		"count", strconv.Itoa(count),
		"total", strconv.Itoa(total))
}
