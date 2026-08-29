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
				"score", strconv.Itoa(p.RowScore(r))))
			if rank := p.PartialRowRank(r); rank >= 0 {
				b.WriteString(i18n.Tf("pokersquares.rowPartialHand", "hand", pokerHandName(rank)))
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")

		colParts := make([]string, domain.PokerSquaresGridSize)
		for i := range domain.PokerSquaresGridSize {
			colScoreStr := i18n.Tf("pokersquares.colScore",
				"idx", strconv.Itoa(i),
				"score", strconv.Itoa(p.ColScore(i)))
			if rank := p.PartialColRank(i); rank >= 0 {
				colScoreStr += i18n.Tf("pokersquares.colPartialHand", "hand", pokerHandName(rank))
			}
			colParts[i] = colScoreStr
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
		// Web は Undo ボタンの disabled で押せないことが見えるが、CUI は `u` を
		// 打ってエラーが返って初めて分かる状態だった (#5538)。戻せるときは
		// 何も足さない -- 毎手「戻せます」と言われても情報にならない。
		if !p.CanUndo() {
			b.WriteString(i18n.T("pokersquares.undoUnavailable") + "\n")
		}

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

// HintOutput emits the best-placement hint for the current card. It suggests
// the empty cell whose row/column offer the strongest poker-hand synergy, or an
// explanatory line when no hint is available (game over or no current card).
func (pr *PokerSquaresCuiPresenter) HintOutput(p interfaces.PokerSquaresGame) string {
	hint := p.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	reason := i18n.T("pokersquares.hintAny")
	if hint.Synergy {
		reason = i18n.T("pokersquares.hintSynergy")
	}
	card := i18n.T("pokersquares.currentCardNone")
	if cc := p.GetCurrentCard(); cc != nil {
		card = cuiCardStr(cc)
	}
	return i18n.Tf("pokersquares.hintLine",
		"card", card,
		"r", strconv.Itoa(hint.Row),
		"c", strconv.Itoa(hint.Col),
		"reason", reason) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *PokerSquaresCuiPresenter) ActionLogOutput(p interfaces.PokerSquaresGame) string {
	if p.GetPhase() == domain.PokerSquaresPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(p.GetActionLog())
}
