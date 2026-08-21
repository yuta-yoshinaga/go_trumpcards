//go:build !js || !wasm || extra4

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

// mrsMopColumnStr returns the display string for a MrsMop tableau column.
func mrsMopColumnStr(colCards []*domain.MrsMopTableauCard) string {
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

// MrsMopCuiPresenter renders the MrsMop Solitaire CUI view.
type MrsMopCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *MrsMopCuiPresenter) Output(s interfaces.MrsMopGame, lastErr error) string {
	return buildCuiOutput(i18n.T("mrsmop.helpTitle"), func(b *strings.Builder) {
		// **山札も残りディールも出さない。**Mrs. Mop は 104 枚を配り切るので
		// どちらも存在しない。常に 0 を出すと「まだ配れる」と読めてしまう。
		b.WriteString(i18n.Tf("mrsmop.header",
			"completed", strconv.Itoa(s.GetCompletedSuits()),
			"total", strconv.Itoa(domain.MrsMopFoundationCnt),
			"score", strconv.Itoa(s.GetScore())) + "\n")
		b.WriteString(i18n.Tf("mrsmop.difficultyLine",
			"suits", strconv.Itoa(int(s.GetDifficulty()))) + "\n")

		b.WriteString("----------\n")

		// Tableau
		tableau := s.GetTableau()
		emptyColumn := false
		for col := range domain.MrsMopTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("mrsmop.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				emptyColumn = true
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(mrsMopColumnStr(colCards))
			}
			b.WriteString("\n")
		}

		// **空列の警告は要らない。**配る操作が無いので、空列は「どの札でも置ける
		// 枠」でしかない。クローン元の Spider では配りを塞ぐ障害だった。
		_ = emptyColumn

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch s.GetPhase() {
		case domain.MrsMopPhasePlaying:
			if s.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := s.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(s.GetMoveCount())) + "\n")
		case domain.MrsMopPhaseGameClear:
			b.WriteString(color.Green(i18n.Tf("mrsmop.gameClearLine",
				"moves", strconv.Itoa(s.GetMoveCount()),
				"score", strconv.Itoa(s.GetScore()))) + "\n")
		case domain.MrsMopPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current MrsMop hint.
func (p *MrsMopCuiPresenter) HintOutput(s interfaces.MrsMopGame) string {
	hint := s.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return i18n.Tf("mrsmop.hintLine",
		"fromCol", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"toCol", strconv.Itoa(hint.ToCol)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *MrsMopCuiPresenter) ActionLogOutput(s interfaces.MrsMopGame) string {
	if s.GetPhase() == domain.MrsMopPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(s.GetActionLog())
}
