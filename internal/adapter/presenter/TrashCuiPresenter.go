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
			b.WriteString(i18n.Tf("trash.pending", "card", cuiCardStr(pending)) +
				trashPendingDestination(t, pending) + "\n")
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

// trashHumanFaceDownSlots returns the 1-based positions of the human's slots
// that are still unfilled (face down) — the valid targets for a wild card.
func trashHumanFaceDownSlots(t interfaces.TrashGame) []string {
	var out []string
	for i, s := range t.GetPlayerSlots(domain.TrashHumanIdx) {
		if !s.FaceUp {
			out = append(out, strconv.Itoa(i+1))
		}
	}
	return out
}

// trashPendingDestination annotates the pending card with where it can go on
// the board of the player holding it. The Web GUI narrates exactly this in
// `pendingAnnounce` and pulses the target slot; the CUI printed the bare card
// and left the player to work the destination out (#4867).
func trashPendingDestination(t interfaces.TrashGame, pending *domain.Card) string {
	slots := t.GetPlayerSlots(t.GetCurrent())
	if trashIsWild(pending) {
		var open []string
		for i, s := range slots {
			if !s.FaceUp {
				open = append(open, strconv.Itoa(i+1))
			}
		}
		if len(open) == 0 {
			return i18n.T("trash.pendingDead")
		}
		// AwaitWild のプロンプト (「p <位置>」) と矛盾しないよう、選べる位置を
		// 並べるだけにする。どこを推すかは HintOutput の役目。
		return i18n.Tf("trash.pendingWild", "slots", strings.Join(open, ", "))
	}
	slot := trashSlotFor(pending)
	if slot == 0 || slot > len(slots) || slots[slot-1].FaceUp {
		return i18n.T("trash.pendingDead")
	}
	return i18n.Tf("trash.pendingSlot", "slot", strconv.Itoa(slot))
}

// trashIsWild reports whether a card is a Trash wild (King or Joker).
func trashIsWild(c *domain.Card) bool {
	return c != nil && (c.GetDesign() == domain.CardDesignJoker || c.GetValue() == 13)
}

// trashSlotFor returns the 1-based slot a rank-value card fills, or 0 for
// end-turn/wild cards that have no fixed slot.
func trashSlotFor(c *domain.Card) int {
	if c == nil || c.GetDesign() == domain.CardDesignJoker {
		return 0
	}
	if v := c.GetValue(); v >= 1 && v <= domain.TrashSlotCnt {
		return v
	}
	return 0
}

// HintOutput suggests where the drawn/wild card should go, or whether the
// face-up discard is worth taking on the player's turn.
func (p *TrashCuiPresenter) HintOutput(t interfaces.TrashGame) string {
	// Advice is always for the human's board, so it only makes sense on the
	// human's turn — the same phases apply to the CPU mid-resolution.
	if t.GetPhase() != domain.TrashPhaseGameOver && t.GetCurrent() != domain.TrashHumanIdx {
		return i18n.T("trash.hintNotYourTurn") + "\n"
	}
	switch t.GetPhase() {
	case domain.TrashPhaseAwaitWild:
		slots := trashHumanFaceDownSlots(t)
		if len(slots) == 0 {
			return i18n.T("trash.hintGameOver") + "\n"
		}
		// Filling the highest open position first keeps the low ranks — which
		// you are equally likely to draw — available to place themselves.
		rec := slots[len(slots)-1]
		return i18n.Tf("trash.hintWild", "slots", strings.Join(slots, ", "), "rec", rec) + "\n"
	case domain.TrashPhasePlayerTurn:
		top := t.GetDiscardTop()
		if trashIsWild(top) && len(trashHumanFaceDownSlots(t)) > 0 {
			return i18n.T("trash.hintTakeDiscardWild") + "\n"
		}
		if slot := trashSlotFor(top); slot > 0 && !t.GetPlayerSlots(domain.TrashHumanIdx)[slot-1].FaceUp {
			return i18n.Tf("trash.hintTakeDiscard", "slot", strconv.Itoa(slot)) + "\n"
		}
		return i18n.T("trash.hintDrawStock") + "\n"
	default:
		return i18n.T("trash.hintGameOver") + "\n"
	}
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *TrashCuiPresenter) ActionLogOutput(t interfaces.TrashGame) string {
	if t.GetPhase() != domain.TrashPhaseGameOver {
		return actionLogToText(nil)
	}
	return actionLogToText(t.GetActionLog())
}
