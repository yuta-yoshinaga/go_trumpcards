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

// PokerSquaresCuiPresenter renders the Poker Squares CUI view.
type PokerSquaresCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (pr *PokerSquaresCuiPresenter) Output(p interfaces.PokerSquaresGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pokersquares.helpTitle"), func(b *strings.Builder) {
		board := p.GetBoard()
		for r := range domain.PokerSquaresGridSize {
			for c := range domain.PokerSquaresGridSize {
				if c > 0 {
					b.WriteString(" | ")
				}
				rs := strconv.Itoa(r)
				cs := strconv.Itoa(c)
				if card := board[r][c]; card == nil {
					b.WriteString(i18n.Tf("pokersquares.cellEmpty", "r", rs, "c", cs))
				} else {
					b.WriteString(i18n.Tf("pokersquares.cellCard",
						"r", rs, "c", cs, "card", cuiCardStr(card)))
				}
			}
			b.WriteString(i18n.Tf("pokersquares.rowScore",
				"score", strconv.Itoa(p.RowScore(r))) + "\n")
		}
		b.WriteString("----------\n")

		colParts := make([]string, domain.PokerSquaresGridSize)
		for i := range domain.PokerSquaresGridSize {
			colParts[i] = i18n.Tf("pokersquares.colScore",
				"idx", strconv.Itoa(i),
				"score", strconv.Itoa(p.ColScore(i)))
		}
		b.WriteString(strings.Join(colParts, " ") + "\n")
		b.WriteString("----------\n")

		if cc := p.GetCurrentCard(); cc != nil {
			b.WriteString(i18n.Tf("pokersquares.currentCard", "card", cuiCardStr(cc)) + "\n")
		} else {
			b.WriteString(i18n.T("pokersquares.currentCardNone") + "\n")
		}
		b.WriteString(i18n.Tf("pokersquares.placedLine",
			"placed", strconv.Itoa(p.GetPlacedCount()),
			"total", strconv.Itoa(domain.PokerSquaresTotalCells),
			"score", strconv.Itoa(p.TotalScore())) + "\n")

		if p.GetPhase() != domain.PokerSquaresPhaseComplete {
			b.WriteString(i18n.T("pokersquares.cuiPlaceHint") + "\n")
		}

		cuiErrorBlock(b, lastErr)

		if p.GetPhase() == domain.PokerSquaresPhaseComplete {
			b.WriteString(color.Green(i18n.Tf("pokersquares.gameComplete",
				"score", strconv.Itoa(p.TotalScore()))) + "\n")
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *PokerSquaresCuiPresenter) ActionLogOutput(p interfaces.PokerSquaresGame) string {
	if p.GetPhase() == domain.PokerSquaresPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(p.GetActionLog())
}
