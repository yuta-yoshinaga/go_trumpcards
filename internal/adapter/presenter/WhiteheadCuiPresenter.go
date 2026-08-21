//go:build !js || !wasm || solo

package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// whiteheadColumnStr returns the display string for a Whitehead tableau column.
func whiteheadColumnStr(colCards []*domain.WhiteheadTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		if tc.FaceUp {
			parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
		} else {
			parts[j] = fmt.Sprintf(" [%d]??", j)
		}
	}
	return strings.Join(parts, " ")
}

// whiteheadScoringModeLabel returns the localized scoring-mode name.
func whiteheadScoringModeLabel(mode domain.WhiteheadScoringMode) string {
	if mode == domain.WhiteheadScoringVegas {
		return i18n.T("whitehead.scoringVegas")
	}
	return i18n.T("whitehead.scoringNone")
}

// WhiteheadCuiPresenter renders the Whitehead Solitaire CUI view.
type WhiteheadCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *WhiteheadCuiPresenter) Output(k interfaces.WhiteheadGame, lastErr error) string {
	return buildCuiOutput(i18n.T("whitehead.helpTitle"), func(b *strings.Builder) {
		// Header: draw mode, scoring mode, and the running Vegas score, matching
		// the web header (none of which the CUI surfaced before).
		b.WriteString(i18n.Tf("whitehead.settingsLine",
			"draw", strconv.Itoa(k.GetDrawCount()),
			"mode", whiteheadScoringModeLabel(k.GetScoringMode()),
			"score", strconv.Itoa(k.GetScore())) + "\n")

		// Foundation
		b.WriteString(i18n.T("whitehead.foundationHeader"))
		foundation := k.GetFoundation()
		for i := 0; i < domain.WhiteheadFoundationCnt; i++ {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		b.WriteString("\n")

		// Stock + waste
		b.WriteString(i18n.Tf("whitehead.stockLine",
			"count", strconv.Itoa(k.GetStockCount())))
		waste := k.GetWaste()
		switch {
		case len(waste) == 0:
			b.WriteString(i18n.T("whitehead.wasteEmpty"))
		case k.GetDrawCount() == 3 && len(waste) > 1:
			// Three-draw: show the last up-to-3 cards as a fan; only the last
			// (top) card can be played, so mark it and note the restriction.
			start := len(waste) - 3
			if start < 0 {
				start = 0
			}
			shown := waste[start:]
			parts := make([]string, len(shown))
			for i, c := range shown {
				parts[i] = cuiCardStr(c)
			}
			parts[len(parts)-1] += "*"
			b.WriteString(i18n.Tf("whitehead.wasteFan", "cards", strings.Join(parts, " ")))
		default:
			b.WriteString(i18n.Tf("whitehead.wasteCard",
				"card", cuiCardStr(waste[len(waste)-1])))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := k.GetTableau()
		for col := 0; col < domain.WhiteheadTableauCnt; col++ {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("whitehead.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(whiteheadColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch k.GetPhase() {
		case domain.WhiteheadPhasePlaying:
			if k.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := k.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			// **もう自動で揃えられることを能動的に知らせる (#4776)。**Web は
			// 条件が揃うとボタンを光らせバッジも出すのに、CUI は ac コマンドが
			// あること自体も、いま使えるかも出していなかった。タブロー全体を
			// 目で確認して自分で判断するしかない。
			if k.CanAutoComplete() {
				b.WriteString(color.Green(i18n.T("whitehead.autoCompleteReady")) + "\n")
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(k.GetMoveCount())) + "\n")
		case domain.WhiteheadPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(k.GetMoveCount())))
			if k.GetScoringMode() == domain.WhiteheadScoringVegas {
				b.WriteString(" " + i18n.Tf("whitehead.clearScore", "score", strconv.Itoa(k.GetScore())))
			}
			b.WriteString("\n")
		case domain.WhiteheadPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Whitehead hint.
func (p *WhiteheadCuiPresenter) HintOutput(k interfaces.WhiteheadGame) string {
	hint := k.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "tableau" {
		from = i18n.Tf("whitehead.hintFromTableau",
			"col", strconv.Itoa(hint.FromCol),
			"idx", strconv.Itoa(hint.CardIndex))
	} else {
		from = i18n.T("whitehead.hintFromWaste")
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("whitehead.hintToFoundation")
	} else {
		to = i18n.Tf("whitehead.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("whitehead.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WhiteheadCuiPresenter) ActionLogOutput(k interfaces.WhiteheadGame) string {
	if k.GetPhase() == domain.WhiteheadPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(k.GetActionLog())
}
