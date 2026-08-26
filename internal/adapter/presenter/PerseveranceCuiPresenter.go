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

// perseveranceColumnStr returns the display string for a Perseverance tableau column.
func perseveranceColumnStr(colCards []*domain.PerseveranceTableauCard) string {
	parts := make([]string, len(colCards))
	for j, tc := range colCards {
		parts[j] = fmt.Sprintf(" [%d]%s", j, cuiCardStr(tc.Card))
	}
	return strings.Join(parts, " ")
}

// PerseveranceCuiPresenter renders the Perseverance Solitaire CUI view.
type PerseveranceCuiPresenter struct{}

// Output renders the current game state for the active locale (#1699).
func (p *PerseveranceCuiPresenter) Output(bd interfaces.PerseveranceGame, lastErr error) string {
	return buildCuiOutput(i18n.T("perseverance.helpTitle"), func(b *strings.Builder) {
		// Foundation
		b.WriteString(i18n.T("perseverance.foundationHeader"))
		foundation := bd.GetFoundation()
		for i := range domain.PerseveranceFoundationCnt {
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

		b.WriteString("----------\n")

		// Tableau
		tableau := bd.GetTableau()
		for col := range domain.PerseveranceTableauCnt {
			colCards := tableau[col]
			b.WriteString(i18n.Tf("perseverance.columnLabel", "col", strconv.Itoa(col)))
			if len(colCards) == 0 {
				b.WriteString(" " + i18n.T("cuiEmptyCol"))
			} else {
				b.WriteString(perseveranceColumnStr(colCards))
				// A column down to its last card is at risk of emptying, and empty
				// columns can never be refilled — flag it so the risk is visible.
				if len(colCards) == 1 {
					b.WriteString(" " + color.Yellow(i18n.T("perseverance.oneCardWarning")))
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch bd.GetPhase() {
		case domain.PerseverancePhasePlaying:
			if bd.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := bd.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
				// **手詰まりの出口は Undo だけではない。**残っていれば配り直せる
				// ことを、その場で言う。Web は redeal ボタンを点滅させている。
				if bd.GetRedealsLeft() > 0 {
					b.WriteString(color.Yellow(i18n.T("perseverance.helpRedeal")) + "\n")
				}
			}
			b.WriteString(i18n.T("perseverance.emptyColNote") + "\n")
			b.WriteString(i18n.Tf("perseverance.redealsLine",
				"count", strconv.Itoa(bd.GetRedealsLeft())) + "\n")
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(bd.GetMoveCount())) + "\n")
		case domain.PerseverancePhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(bd.GetMoveCount())) + "\n")
		case domain.PerseverancePhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Perseverance hint.
func (p *PerseveranceCuiPresenter) HintOutput(bd interfaces.PerseveranceGame) string {
	hint := bd.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	from := i18n.Tf("perseverance.hintFrom",
		"col", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex))
	var to string
	if hint.ToZone == "foundation" {
		to = i18n.T("perseverance.hintToFoundation")
	} else {
		to = i18n.Tf("perseverance.hintToTableau", "col", strconv.Itoa(hint.ToCol))
	}
	return i18n.Tf("perseverance.hintLine", "from", from, "to", to) + "\n"
}

// TargetsOutput は列 col の一番下の札を置ける先を一覧する。
//
// **13 列 + 4 組札は押して試すには広すぎる。**Web は選択の瞬間に置ける先を
// リングで示しているのに、CUI は打ってサーバに弾かれるまで分からなかった
// (#5581)。置ける先が無いときは黙らず、無いと言う ── 空行だと、コマンドが
// 効いていないのか置けないのか区別が付かない。
func (p *PerseveranceCuiPresenter) TargetsOutput(bd interfaces.PerseveranceGame, col int) string {
	if col < 0 || col >= domain.PerseveranceTableauCnt {
		return i18n.MarkError(i18n.Tf("invalidColumn", "val", strconv.Itoa(col))) + "\n"
	}
	tableau, foundation := bd.LegalTargets(col)
	if len(tableau) == 0 && len(foundation) == 0 {
		return i18n.Tf("perseverance.targetsNone", "col", strconv.Itoa(col)) + "\n"
	}
	parts := make([]string, 0, len(tableau)+len(foundation))
	for _, t := range tableau {
		parts = append(parts, i18n.Tf("perseverance.targetTableau", "col", strconv.Itoa(t)))
	}
	for _, f := range foundation {
		parts = append(parts, i18n.Tf("perseverance.targetFoundation", "idx", strconv.Itoa(f)))
	}
	return i18n.Tf("perseverance.targetsLine",
		"col", strconv.Itoa(col),
		"targets", strings.Join(parts, " / ")) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *PerseveranceCuiPresenter) ActionLogOutput(bd interfaces.PerseveranceGame) string {
	return actionLogOutputText(bd)
}
