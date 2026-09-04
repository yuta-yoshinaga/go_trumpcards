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

// BristolCuiPresenter renders the Bristol Solitaire CUI view.
type BristolCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (p *BristolCuiPresenter) Output(b interfaces.BristolGame, lastErr error) string {
	return buildCuiOutput(i18n.T("bristol.helpTitle"), func(sb *strings.Builder) {
		// Foundations (top card of each)
		foundation := b.GetFoundation()
		for i := range domain.BristolFoundationCnt {
			sb.WriteString(i18n.Tf("bristol.foundationLabel", "idx", strconv.Itoa(i)))
			pile := foundation[i]
			if len(pile) == 0 {
				sb.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				sb.WriteString(" " + cuiCardStr(pile[len(pile)-1]) +
					i18n.Tf("bristol.pileCount", "count", strconv.Itoa(len(pile))))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("----------\n")

		// Tableau columns
		sb.WriteString(i18n.Tf("bristol.tableauRule", "count", strconv.Itoa(domain.BristolTableauCnt)) + "\n")
		tableau := b.GetTableau()
		for i := range domain.BristolTableauCnt {
			sb.WriteString(i18n.Tf("bristol.tableauLabel", "idx", strconv.Itoa(i)))
			col := tableau[i]
			if len(col) == 0 {
				sb.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				sb.WriteString(" " + formatCardSlice(col, cuiCardStr, ","))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("----------\n")

		// Fans (top card of each)
		fan := b.GetFan()
		for i := range domain.BristolFanCnt {
			sb.WriteString(i18n.Tf("bristol.fanLabel", "idx", strconv.Itoa(i)))
			pile := fan[i]
			if len(pile) == 0 {
				sb.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				sb.WriteString(" " + cuiCardStr(pile[len(pile)-1]) +
					i18n.Tf("bristol.pileCount", "count", strconv.Itoa(len(pile))))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("----------\n")

		// Stock
		sb.WriteString(i18n.Tf("bristol.stockLine", "count", strconv.Itoa(b.GetStockCount())))
		sb.WriteString("\n----------\n")

		cuiErrorBlock(sb, lastErr)

		switch b.GetPhase() {
		case domain.BristolPhasePlaying:
			// **ストックは作り直せない。**配り切ると手詰まりに普通に到達するので、
			// 他のソリティアと同じように知らせる (#5631)。
			if b.IsStalemate() {
				sb.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				if n := b.UndoToEscape(); n > 0 {
					sb.WriteString(color.Yellow(i18n.Tf("bristol.undoToEscape", "count", strconv.Itoa(n))) + "\n")
				}
			}
			sb.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(b.GetMoveCount())) +
				cuiSolitaireUndoHint(b.CanUndo()) + "\n")
		case domain.BristolPhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(b.GetMoveCount())) + "\n")
		case domain.BristolPhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Bristol hint.
func (p *BristolCuiPresenter) HintOutput(b interfaces.BristolGame) string {
	hint := b.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	var from string
	if hint.FromZone == "fan" {
		from = i18n.Tf("bristol.hintFromFan", "col", strconv.Itoa(hint.FromCol))
	} else {
		from = i18n.Tf("bristol.hintFromTableau", "col", strconv.Itoa(hint.FromCol))
	}
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.Tf("bristol.hintToFoundation", "idx", strconv.Itoa(hint.ToCol))
	} else {
		to = i18n.Tf("bristol.hintToTableau", "idx", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("bristol.hintLine", "from", from, "to", to) + "\n"
}

// TargetsOutput は移動元ゾーン zone の列 col の札を置ける先を一覧する。
//
// **8 列 + 3 ファン + 4 組札は押して試すには広すぎる。**Web は選択の瞬間に置ける先を
// リングで示しているのに、CUI は打ってサーバに弾かれるまで分からなかった
// (#6427)。置ける先が無いときは黙らず、無いと言う ── 空行だと、コマンドが
// 効いていないのか置けないのか区別が付かない。
func (p *BristolCuiPresenter) TargetsOutput(b interfaces.BristolGame, zone string, col int) string {
	if zone != "tableau" && zone != "fan" {
		return i18n.MarkError(i18n.Tf("bristol.invalidFromZone", "val", zone)) + "\n"
	}
	if col < 0 || (zone == "tableau" && col >= domain.BristolTableauCnt) || (zone == "fan" && col >= domain.BristolFanCnt) {
		return i18n.MarkError(i18n.Tf("invalidColumn", "val", strconv.Itoa(col))) + "\n"
	}
	var from string
	if zone == "tableau" {
		from = i18n.Tf("bristol.fromTableau", "col", strconv.Itoa(col))
	} else {
		from = i18n.Tf("bristol.fromFan", "idx", strconv.Itoa(col))
	}
	tableau, foundation := b.LegalTargets(zone, col)
	if len(tableau) == 0 && len(foundation) == 0 {
		return i18n.Tf("bristol.targetsNone", "from", from) + "\n"
	}
	parts := make([]string, 0, len(tableau)+len(foundation))
	for _, t := range tableau {
		parts = append(parts, i18n.Tf("bristol.targetTableau", "col", strconv.Itoa(t)))
	}
	for _, f := range foundation {
		parts = append(parts, i18n.Tf("bristol.targetFoundation", "idx", strconv.Itoa(f)))
	}
	return i18n.Tf("bristol.targetsLine",
		"from", from,
		"targets", strings.Join(parts, " / ")) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BristolCuiPresenter) ActionLogOutput(b interfaces.BristolGame) string {
	if b.GetPhase() == domain.BristolPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(b.GetActionLog())
}
