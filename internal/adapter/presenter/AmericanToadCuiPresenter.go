//go:build !js || !wasm || extra2

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

// americanToadColumnStr returns the display string for one tableau column.
func americanToadColumnStr(colCards []*domain.AmericanToadTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// AmericanToadCuiPresenter renders the American Toad CUI view.
type AmericanToadCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *AmericanToadCuiPresenter) Output(at interfaces.AmericanToadGame, lastErr error) string {
	return buildCuiOutput(i18n.T("americantoad.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("americantoad.baseRankLine",
			"rank", strconv.Itoa(at.GetBaseRank())) + "\n")

		// 基礎札は 8 つある。1 行に並べると長いので、山ごとに枚数も出す。
		b.WriteString(i18n.T("americantoad.foundationHeader"))
		foundation := at.GetFoundation()
		for i := range domain.AmericanToadFoundationCnt {
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

		// リザーブは一番上だけが使える。残り枚数がそのまま難度に効く。
		reserve := at.GetReserve()
		if len(reserve) == 0 {
			b.WriteString(i18n.T("americantoad.reserveEmpty") + "\n")
		} else {
			b.WriteString(i18n.Tf("americantoad.reserveLine",
				"card", cuiCardStr(reserve[len(reserve)-1]),
				"count", strconv.Itoa(len(reserve))) + "\n")
		}

		b.WriteString(i18n.Tf("americantoad.stockLine", "count", strconv.Itoa(at.GetStockCount())))
		waste := at.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("americantoad.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("americantoad.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		tableau := at.GetTableau()
		for col := range domain.AmericanToadTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("americantoad.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(americanToadColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch at.GetPhase() {
		case domain.AmericanToadPhasePlaying:
			// めくり直しは 1 回だけなので、使えるうちは目立たせる。
			if at.CanRedeal() {
				b.WriteString(color.Yellow(i18n.T("americantoad.redealAvailable")) + "\n")
			}
			if at.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := at.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("americantoad.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(at.GetMoveCount())) + "\n")
		case domain.AmericanToadPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(at.GetMoveCount())) + "\n")
		case domain.AmericanToadPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current American Toad hint.
func (p *AmericanToadCuiPresenter) HintOutput(at interfaces.AmericanToadGame) string {
	hint := at.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "reserve":
		from = i18n.T("americantoad.hintFromReserve")
	case "waste":
		from = i18n.T("americantoad.hintFromWaste")
	case "stock":
		from = i18n.T("americantoad.hintFromStock")
	default:
		from = i18n.Tf("americantoad.hintFromTableau",
			"col", strconv.Itoa(hint.FromIdx),
			"idx", strconv.Itoa(hint.CardIndex))
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("americantoad.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "waste":
		to = i18n.T("americantoad.hintToWaste")
	default:
		to = i18n.Tf("americantoad.hintToTableau", "col", strconv.Itoa(hint.ToIdx))
	}
	return i18n.Tf("americantoad.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *AmericanToadCuiPresenter) ActionLogOutput(at interfaces.AmericanToadGame) string {
	if at.GetPhase() == domain.AmericanToadPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(at.GetActionLog())
}
