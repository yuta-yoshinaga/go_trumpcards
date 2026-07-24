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

// triPeaksAdjacentRank reports whether two ranks differ by one, treating
// King(13) and Ace(1) as adjacent (mirrors the domain's isAdjacentRank, which
// is unexported, and the frontend's isTriPeaksAdjacent).
func triPeaksAdjacentRank(v1, v2 int) bool {
	diff := v1 - v2
	if diff < 0 {
		diff = -diff
	}
	return diff == 1 || diff == 12
}

// TriPeaksCuiPresenter renders the TriPeaks Solitaire CUI view.
type TriPeaksCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (pr *TriPeaksCuiPresenter) Output(t interfaces.TriPeaksGame, lastErr error) string {
	return buildCuiOutput(i18n.T("tripeaks.helpTitle"), func(b *strings.Builder) {
		layout := t.GetLayout()

		// Waste top drives the ±1 playable check for exposed tableau cards.
		waste := t.GetWaste()
		var wasteTop *domain.Card
		if len(waste) > 0 {
			wasteTop = waste[len(waste)-1]
		}

		// Count exposed tableau cards playable onto the current waste top so the
		// summary line below can spare the player from scanning all four rows.
		playableCount := 0

		// Tableau (with row indent)
		for row := range domain.TriPeaksRowCnt {
			indent := strings.Repeat("  ", domain.TriPeaksRowCnt-1-row)
			b.WriteString(indent)
			first := true
			for col := range domain.TriPeaksColCnt {
				tc := layout[row][col]
				if tc == nil {
					continue
				}
				if !first {
					b.WriteString("  ")
				}
				first = false
				switch {
				case tc.Removed:
					b.WriteString("    ")
				case !t.IsExposed(row, col):
					// Blocked cards hide coordinates: they cannot be taken yet.
					b.WriteString(i18n.Tf("tripeaks.blockedCard", "card", cuiCardStr(tc.Card)))
				case wasteTop != nil && triPeaksAdjacentRank(tc.Card.GetValue(), wasteTop.GetValue()):
					// Exposed and ±1 from the waste top: playable now (trailing *).
					playableCount++
					b.WriteString(i18n.Tf("tripeaks.playableCard",
						"row", strconv.Itoa(row),
						"col", strconv.Itoa(col),
						"card", cuiCardStr(tc.Card)))
				default:
					// Exposed but not adjacent: shown with coordinates, no marker.
					b.WriteString(i18n.Tf("tripeaks.tableauCard",
						"row", strconv.Itoa(row),
						"col", strconv.Itoa(col),
						"card", cuiCardStr(tc.Card)))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		// Stock + waste
		b.WriteString(i18n.Tf("tripeaks.stockLine",
			"count", strconv.Itoa(t.GetStockCount())))
		if len(waste) > 0 {
			b.WriteString(i18n.Tf("tripeaks.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		} else {
			b.WriteString(i18n.T("tripeaks.wasteEmpty"))
		}
		b.WriteString("\n")

		// Playable-now summary: how many cards can be taken onto the waste top,
		// plus a pre-emptive draw hint when none can (and the stock isn't empty).
		if t.GetPhase() == domain.TriPeaksPhasePlaying {
			b.WriteString(i18n.Tf("tripeaks.playableSummary",
				"count", strconv.Itoa(playableCount)) + "\n")
			if playableCount == 0 && t.GetStockCount() > 0 {
				b.WriteString(color.Yellow(i18n.T("tripeaks.drawRecommended")) + "\n")
			}
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch t.GetPhase() {
		case domain.TriPeaksPhasePlaying:
			if t.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(t.GetMoveCount())) + "\n")
		case domain.TriPeaksPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(t.GetMoveCount())) + "\n")
		case domain.TriPeaksPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current TriPeaks hint.
func (pr *TriPeaksCuiPresenter) HintOutput(t interfaces.TriPeaksGame) string {
	hint := t.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	switch hint.Type {
	case "remove":
		return i18n.Tf("tripeaks.hintRemove",
			"row", strconv.Itoa(hint.Row),
			"col", strconv.Itoa(hint.Col)) + "\n"
	case "draw":
		return i18n.T("tripeaks.hintDraw") + "\n"
	default:
		return i18n.T("tripeaks.hintUnknown") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *TriPeaksCuiPresenter) ActionLogOutput(t interfaces.TriPeaksGame) string {
	if t.GetPhase() == domain.TriPeaksPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(t.GetActionLog())
}
