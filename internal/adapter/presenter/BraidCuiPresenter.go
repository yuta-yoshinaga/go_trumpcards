//go:build !js || !wasm || extra2

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// braidSlotLine renders one row of single-card slots as "label0:CARD label1:[空]".
func braidSlotLine(label string, cards []*domain.Card) string {
	parts := make([]string, len(cards))
	for i, card := range cards {
		text := i18n.T("cuiEmptyCol")
		if card != nil {
			text = cuiCardStr(card)
		}
		parts[i] = i18n.Tf("braid.slotLabel", "label", label, "idx", strconv.Itoa(i)) + text
	}
	return strings.Join(parts, " ")
}

// BraidCuiPresenter renders the Braid CUI view.
type BraidCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BraidCuiPresenter) Output(b interfaces.BraidGame, lastErr error) string {
	return buildCuiOutput(i18n.T("braid.helpTitle"), func(sb *strings.Builder) {
		// 向きが未選択のうちは、その 1 手が最重要なので先頭に出す。
		if b.IsAwaitingDirection() {
			sb.WriteString(color.Yellow(i18n.Tf("braid.awaitingDirection",
				"rank", strconv.Itoa(b.GetBaseRank()))) + "\n")
		} else {
			dirKey := "braid.directionDescending"
			if b.GetDirection() == domain.BraidDirectionAscending {
				dirKey = "braid.directionAscending"
			}
			sb.WriteString(i18n.Tf("braid.baseRankLine",
				"rank", strconv.Itoa(b.GetBaseRank()),
				"direction", i18n.T(dirKey)) + "\n")
		}

		sb.WriteString(i18n.T("braid.foundationHeader"))
		foundation := b.GetFoundation()
		for i := range domain.BraidFoundationCnt {
			if i != 0 {
				sb.WriteString(" | ")
			}
			pile := foundation[i]
			if len(pile) == 0 {
				sb.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				sb.WriteString(cuiCardStr(pile[len(pile)-1]))
			}
		}
		sb.WriteString("\n")

		// ブレイドは末尾 1 枚しか使えない。残り枚数がそのまま重みになる。
		braid := b.GetBraid()
		if len(braid) == 0 {
			sb.WriteString(i18n.T("braid.braidEmpty") + "\n")
		} else {
			sb.WriteString(i18n.Tf("braid.braidLine",
				"card", cuiCardStr(braid[len(braid)-1]),
				"count", strconv.Itoa(len(braid))) + "\n")
		}

		fields := b.GetFields()
		sb.WriteString(braidSlotLine(i18n.T("braid.fieldLabel"), fields[:]) + "\n")
		helpers := b.GetHelpers()
		sb.WriteString(braidSlotLine(i18n.T("braid.helperLabel"), helpers[:]) + "\n")

		sb.WriteString(i18n.Tf("braid.stockLine",
			"count", strconv.Itoa(b.GetStockCount()),
			"redeals", strconv.Itoa(domain.BraidMaxPasses-1-b.GetPassesUsed())))
		waste := b.GetWaste()
		if len(waste) == 0 {
			sb.WriteString(" " + i18n.T("braid.wasteEmpty"))
		} else {
			sb.WriteString(" " + i18n.Tf("braid.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		sb.WriteString("\n")

		sb.WriteString("----------\n")

		cuiErrorBlock(sb, lastErr)

		switch b.GetPhase() {
		case domain.BraidPhasePlaying:
			if b.IsStalemate() {
				sb.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := b.UndoToEscape(); n > 0 {
					sb.WriteString(color.Yellow(i18n.Tf("braid.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			sb.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(b.GetMoveCount())) + "\n")
		case domain.BraidPhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(b.GetMoveCount())) + "\n")
		case domain.BraidPhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Braid hint.
func (p *BraidCuiPresenter) HintOutput(b interfaces.BraidGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	// 向きの選択だけは移動ではないので、from/to の形に載せず専用の一文にする。
	if hint.FromZone == "direction" {
		return i18n.T("braid.hintChooseDirection") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "braid":
		from = i18n.T("braid.hintFromBraid")
	case "field":
		from = i18n.Tf("braid.hintFromField", "idx", strconv.Itoa(hint.FromIdx))
	case "helper":
		from = i18n.Tf("braid.hintFromHelper", "idx", strconv.Itoa(hint.FromIdx))
	case "stock":
		from = i18n.T("braid.hintFromStock")
	default:
		from = i18n.T("braid.hintFromWaste")
	}
	var to string
	switch hint.ToZone {
	case "foundation":
		to = i18n.Tf("braid.hintToFoundation", "idx", strconv.Itoa(hint.ToIdx))
	case "helper":
		to = i18n.Tf("braid.hintToHelper", "idx", strconv.Itoa(hint.ToIdx))
	default:
		to = i18n.T("braid.hintToWaste")
	}
	return i18n.Tf("braid.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BraidCuiPresenter) ActionLogOutput(b interfaces.BraidGame) string {
	if b.GetPhase() == domain.BraidPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(b.GetActionLog())
}
