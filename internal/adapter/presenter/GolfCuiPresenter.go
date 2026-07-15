//go:build !js || !wasm || solo

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// golfAdjacentRank reports whether two ranks differ by one, treating King(13)
// and Ace(1) as adjacent — matching the domain's isAdjacentRank and the
// frontend's isGolfAdjacent (K-A wrap included).
func golfAdjacentRank(v1, v2 int) bool {
	diff := v1 - v2
	if diff < 0 {
		diff = -diff
	}
	return diff == 1 || diff == 12
}

// GolfCuiPresenter renders the Golf Solitaire CUI view.
type GolfCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (pr *GolfCuiPresenter) Output(g interfaces.GolfGame, lastErr error) string {
	return buildCuiOutput(i18n.T("golf.helpTitle"), func(b *strings.Builder) {
		layout := g.GetLayout()

		// Waste top drives the ±1 playable check for exposed tableau cards.
		waste := g.GetWaste()
		var wasteTop *domain.Card
		if len(waste) > 0 {
			wasteTop = waste[len(waste)-1]
		}

		// Tableau (each row prints 7 columns)
		for row := range domain.GolfRowCnt {
			for col := range domain.GolfColCnt {
				if col > 0 {
					b.WriteString("  ")
				}
				gc := layout[col][row]
				switch {
				case gc == nil || gc.Removed:
					b.WriteString("    ")
				case g.IsExposed(col, row) && wasteTop != nil && golfAdjacentRank(gc.Card.GetValue(), wasteTop.GetValue()):
					// Exposed and ±1 from the waste top: playable now (trailing *).
					b.WriteString(i18n.Tf("golf.playableCard",
						"col", strconv.Itoa(col),
						"card", cuiCardStr(gc.Card)))
				case g.IsExposed(col, row):
					b.WriteString(i18n.Tf("golf.exposedCard",
						"col", strconv.Itoa(col),
						"card", cuiCardStr(gc.Card)))
				default:
					b.WriteString("   " + cuiCardStr(gc.Card))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// Stock + waste
		b.WriteString(i18n.Tf("golf.stockLine",
			"count", strconv.Itoa(g.GetStockCount())))
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("golf.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("golf.wasteEmpty"))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.GolfPhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.GolfPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.GolfPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Golf hint.
func (pr *GolfCuiPresenter) HintOutput(g interfaces.GolfGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.Type {
	case "remove":
		return i18n.Tf("golf.hintRemove", "col", strconv.Itoa(hint.Col)) + "\n"
	case "draw":
		return i18n.T("golf.hintDraw") + "\n"
	default:
		return i18n.T("golf.hintUnknown") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *GolfCuiPresenter) ActionLogOutput(g interfaces.GolfGame) string {
	if g.GetPhase() == domain.GolfPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
