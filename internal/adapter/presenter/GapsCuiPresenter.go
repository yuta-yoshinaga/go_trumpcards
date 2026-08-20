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

// GapsCuiPresenter はGapsゲームのCUIプレゼンター。
type GapsCuiPresenter struct{}

// Output は現在のゲーム状態をテキストで描画する。
func (pr *GapsCuiPresenter) Output(g interfaces.GapsGame, lastErr error) string {
	return buildCuiOutput(i18n.T("gaps.helpTitle"), func(b *strings.Builder) {
		grid := g.GetGrid()
		locks := g.GetLockedPrefixLengths()
		for r := 0; r < domain.GapsRowCnt; r++ {
			for c := 0; c < domain.GapsColCnt; c++ {
				if c > 0 {
					b.WriteString(" ")
				}
				cell := grid[r][c]
				if cell == nil {
					// **どのカードが入るかも詰みかも分からなかった (#4800)。**
					// Web はゴーストカードと 🚫 で常時プレビューしている。
					b.WriteString(gapsGapCell(g, r, c))
				} else {
					b.WriteString(i18n.Tf("gaps.gridCard",
						"row", strconv.Itoa(r),
						"col", strconv.Itoa(c),
						"card", cuiCardStr(cell)))
					// Mark cards in the row's locked prefix — those kept on a redeal.
					if c < locks[r] {
						b.WriteString(i18n.T("gaps.lockedMark"))
					}
				}
			}
			b.WriteString("\n")
		}
		b.WriteString("----------\n")
		b.WriteString(i18n.Tf("gaps.redealsLine",
			"used", strconv.Itoa(g.GetRedealsUsed()),
			"remaining", strconv.Itoa(g.GetRedealsRemaining())))
		b.WriteString("\n")
		b.WriteString(i18n.T("gaps.lockedLegend") + "\n")
		b.WriteString("----------\n")
		cuiErrorBlock(b, lastErr)
		switch g.GetPhase() {
		case domain.GapsPhasePlaying:
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("cuiSolitaireStalemate")) + "\n")
				// Tell the player how many undos escape the dead end, matching the
				// web StalemateEscapeButton.
				if n := g.UndoToEscape(); n > 0 {
					b.WriteString(color.Yellow(i18n.Tf("cuiSolitaireUndoToEscape",
						"count", strconv.Itoa(n))) + "\n")
				}
			}
			b.WriteString(i18n.Tf("cuiSolitaireMoves",
				"count", strconv.Itoa(g.GetMoveCount())) +
				cuiSolitaireUndoHint(g.CanUndo()) + "\n")
		case domain.GapsPhaseGameClear:
			b.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.GapsPhaseGameOver:
			b.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput はヒントをテキストで返す。
func (pr *GapsCuiPresenter) HintOutput(g interfaces.GapsGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return i18n.Tf("gaps.hintMove",
		"fromRow", strconv.Itoa(hint.FromRow),
		"fromCol", strconv.Itoa(hint.FromCol),
		"toRow", strconv.Itoa(hint.ToRow),
		"toCol", strconv.Itoa(hint.ToCol)) + "\n"
}

// ActionLogOutput は棋譜をテキストで返す。
func (pr *GapsCuiPresenter) ActionLogOutput(g interfaces.GapsGame) string {
	if g.GetPhase() == domain.GapsPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}

// gapsGapCell は空きマスの表示を返す。何が入るか決まらないときは従来どおり
// [ . ] のまま。**決まらないものを決まったように見せない。**
func gapsGapCell(g interfaces.GapsGame, row, col int) string {
	need := g.GetGapNeed(row, col)
	if need == nil {
		return "[ . ]"
	}
	switch need.Kind {
	case domain.GapsNeedBlocked:
		return i18n.T("gaps.gapBlocked")
	case domain.GapsNeedAnySuit:
		return i18n.Tf("gaps.gapAnySuit", "rank", strconv.Itoa(need.Value))
	case domain.GapsNeedCard:
		return i18n.Tf("gaps.gapNeeded",
			"card", cuiCardStr(domain.NewCard(need.Design, need.Value, false)))
	}
	return "[ . ]"
}
