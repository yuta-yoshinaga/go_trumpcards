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

// TrashCuiPresenter renders the Trash CUI view.
type TrashCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *TrashCuiPresenter) Output(t interfaces.TrashGame, lastErr error) string {
	return buildCuiOutput(i18n.T("trash.helpTitle"), func(b *strings.Builder) {
		// Opponent first, then player — mirrors the physical table view.
		for _, idx := range [...]int{domain.TrashCpuIdx, domain.TrashHumanIdx} {
			label := i18n.T("trash.playerLabelHuman")
			if t.IsCpuPlayer(idx) {
				label = i18n.T("trash.playerLabelCpu")
			}
			if idx == t.GetCurrent() && t.GetPhase() != domain.TrashPhaseGameOver {
				label += i18n.T("trash.playerLabelTurnSuffix")
			}
			b.WriteString(i18n.Tf("trash.playerLineHeader", "label", label) + "\n  ")
			for j, s := range t.GetPlayerSlots(idx) {
				idxStr := fmt.Sprintf("%02d", j+1)
				if s.FaceUp && s.Card != nil {
					b.WriteString(i18n.Tf("trash.slotFaceUp",
						"idx", idxStr, "card", cuiCardStr(s.Card)))
				} else {
					b.WriteString(i18n.Tf("trash.slotFaceDown", "idx", idxStr))
				}
				if (j+1)%5 == 0 && j+1 < domain.TrashSlotCnt {
					b.WriteString("\n  ")
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("trash.stockLine",
			"stock", strconv.Itoa(t.GetStockSize()),
			"discard", strconv.Itoa(t.GetDiscardSize())))
		if top := t.GetDiscardTop(); top != nil {
			b.WriteString(i18n.Tf("trash.stockTop", "card", cuiCardStr(top)))
		}
		b.WriteString("\n")
		if pending := t.GetPending(); pending != nil {
			b.WriteString(i18n.Tf("trash.pending", "card", cuiCardStr(pending)) + "\n")
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		movesLine := i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(t.GetMoveCount()))
		switch t.GetPhase() {
		case domain.TrashPhasePlayerTurn:
			b.WriteString(movesLine + "\n")
		case domain.TrashPhaseAwaitWild:
			b.WriteString(color.Green(i18n.T("trash.promptAwaitWild")) + " " + movesLine + "\n")
		case domain.TrashPhaseGameOver:
			if t.GetWinner() == domain.TrashHumanIdx {
				b.WriteString(color.Green(i18n.T("trash.winHuman")) + " " + movesLine + "\n")
			} else {
				b.WriteString(color.Red(i18n.T("trash.winCpu")) + " " + movesLine + "\n")
			}
		}
	})
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TrashCuiPresenter) ActionLogOutput(t interfaces.TrashGame) string {
	if t.GetPhase() != domain.TrashPhaseGameOver {
		return actionLogToText(nil)
	}
	return actionLogToText(t.GetActionLog())
}
