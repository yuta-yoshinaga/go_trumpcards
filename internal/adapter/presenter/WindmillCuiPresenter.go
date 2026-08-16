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

// WindmillCuiPresenter renders the Windmill CUI view.
type WindmillCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *WindmillCuiPresenter) Output(w interfaces.WindmillGame, lastErr error) string {
	return buildCuiOutput(i18n.T("windmill.helpTitle"), func(b *strings.Builder) {
		// 中央基礎札。52 枚まで積むので、一番上と残り枚数の両方を出す。
		center := w.GetCenter()
		b.WriteString(i18n.T("windmill.centerHeader"))
		if len(center) == 0 {
			b.WriteString(i18n.T("cuiEmptyCol"))
		} else {
			b.WriteString(i18n.Tf("windmill.pileEntry",
				"card", cuiCardStr(center[len(center)-1]),
				"count", strconv.Itoa(len(center)),
				"total", strconv.Itoa(domain.WindmillCenterTarget)))
		}
		b.WriteString("\n")

		// 四隅の K 基礎札
		b.WriteString(i18n.T("windmill.cornerHeader"))
		corners := w.GetCorners()
		for i := range domain.WindmillCornerCnt {
			if i != 0 {
				b.WriteString(" | ")
			}
			pile := corners[i]
			b.WriteString("#" + strconv.Itoa(i) + " ")
			if len(pile) == 0 {
				b.WriteString(i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(i18n.Tf("windmill.pileEntry",
					"card", cuiCardStr(pile[len(pile)-1]),
					"count", strconv.Itoa(len(pile)),
					"total", strconv.Itoa(domain.WindmillCornerTarget)))
			}
		}
		b.WriteString("\n")

		// 山札と捨て札
		b.WriteString(i18n.Tf("windmill.stockLine", "count", strconv.Itoa(w.GetStockCount())))
		waste := w.GetWaste()
		if len(waste) == 0 {
			b.WriteString(" " + i18n.T("windmill.wasteEmpty"))
		} else {
			b.WriteString(" " + i18n.Tf("windmill.wasteTop",
				"card", cuiCardStr(waste[len(waste)-1]),
				"count", strconv.Itoa(len(waste))))
		}
		b.WriteString("\n")

		b.WriteString("----------\n")

		// 十字（帆）8 枠。補充が尽きた枠は空のまま残る。
		sails := w.GetSails()
		for i := range domain.WindmillSailCnt {
			b.WriteString(i18n.Tf("windmill.sailLabel", "idx", strconv.Itoa(i)))
			if sails[i] == nil {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(" " + cuiCardStr(sails[i]))
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch w.GetPhase() {
		case domain.WindmillPhasePlaying:
			// 引き戻し直後は救済手が使えないので、その理由を明示する。
			if w.IsTransferBlocked() {
				b.WriteString(color.Yellow(i18n.T("windmill.transferBlocked")) + "\n")
			}
			if w.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := w.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("windmill.undoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(w.GetMoveCount())) + "\n")
		case domain.WindmillPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(w.GetMoveCount())) + "\n")
		case domain.WindmillPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
			corners := w.GetCorners()
			b.WriteString(color.Yellow(cuiSolitaireGameOverSummary(
				len(w.GetCenter())+cuiCountPileCards(corners[:]...),
				domain.WindmillTotalCards)) + "\n")
		}
	})
}

// HintOutput emits the current Windmill hint.
func (p *WindmillCuiPresenter) HintOutput(w interfaces.WindmillGame) string {
	hint := w.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	switch hint.FromZone {
	case "sail":
		from = i18n.Tf("windmill.hintFromSail", "idx", strconv.Itoa(hint.FromIdx))
	case "waste":
		from = i18n.T("windmill.hintFromWaste")
	case "corner":
		from = i18n.Tf("windmill.hintFromCorner", "idx", strconv.Itoa(hint.FromIdx))
	default:
		from = i18n.T("windmill.hintFromStock")
	}
	var to string
	switch hint.ToZone {
	case "center":
		to = i18n.T("windmill.hintToCenter")
	case "corner":
		to = i18n.Tf("windmill.hintToCorner", "idx", strconv.Itoa(hint.ToIdx))
	default:
		to = i18n.T("windmill.hintToWaste")
	}
	return i18n.Tf("windmill.hintLine", "from", from, "to", to) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *WindmillCuiPresenter) ActionLogOutput(w interfaces.WindmillGame) string {
	if w.GetPhase() == domain.WindmillPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(w.GetActionLog())
}
