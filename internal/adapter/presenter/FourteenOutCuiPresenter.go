//go:build !js || !wasm || extra4

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// FourteenOutCuiPresenter renders the Fourteen Out Solitaire CUI view.
type FourteenOutCuiPresenter struct{}

// Output renders the current game state for the active locale.
func (pr *FourteenOutCuiPresenter) Output(g interfaces.FourteenOutGame, lastErr error) string {
	return buildCuiOutput(i18n.T("fourteenout.helpTitle"), func(b *strings.Builder) {
		b.WriteString(i18n.Tf("fourteenout.header",
			"removed", strconv.Itoa(g.GetRemovedCount())) + "\n")
		b.WriteString("----------\n")

		// **末尾だけが動かせる札。**列の中身も見せないと次の一手が読めないが、
		// 動かせるのがどれかは一目で分かる必要があるので、末尾に印を付ける。
		for c, col := range g.GetColumns() {
			cs := strconv.Itoa(c)
			if len(col) == 0 {
				b.WriteString(i18n.Tf("fourteenout.columnEmpty", "col", cs) + "\n")
				continue
			}
			cards := make([]string, len(col))
			for i, card := range col {
				cards[i] = cuiCardStr(card)
			}
			b.WriteString(i18n.Tf("fourteenout.columnLine",
				"col", cs,
				"cards", strings.Join(cards, " "),
				"tail", cuiCardStr(col[len(col)-1])) + "\n")
		}
		b.WriteString("----------\n")

		cuiErrorBlock(b, lastErr)

		switch g.GetPhase() {
		case domain.FourteenOutPhasePlaying:
			// **これがこのゲームの判断材料そのもの。**Web は常時カウンタとして
			// 出しているのに、CUI は列を目で走査させていた (#5587)。
			// **0 はそのまま敗北を意味する** ── 補充で救われるクローン元と違い、
			// Fourteen Out に山札は無い。
			pairs := g.CountRemovablePairs()
			line := i18n.Tf("fourteenout.removablePairs", "count", strconv.Itoa(pairs))
			if pairs == 0 {
				line = color.Yellow(line)
			}
			b.WriteString(line + "\n")
			if g.IsStalemate() {
				b.WriteString(color.Red(i18n.T("fourteenout.stalemate")) + "\n")
			}
		case domain.FourteenOutPhaseGameClear:
			b.WriteString(color.Green(i18n.T("fourteenout.gameClear")) + "\n")
		case domain.FourteenOutPhaseGameOver:
			b.WriteString(color.Red(i18n.T("fourteenout.gameOver")) + "\n")
		}
	})
}

// HintOutput emits the current Fourteen Out hint.
func (pr *FourteenOutCuiPresenter) HintOutput(g interfaces.FourteenOutGame) string {
	hint := g.Hint()
	if hint == nil {
		return i18n.T("fourteenout.noHint") + "\n"
	}
	// **"deal" のヒントは無い。**山札が無いので、組が見つからなければ nil で返る。
	cols := g.GetColumns()
	c1 := fourteenOutTail(cols, hint.FromCol)
	c2 := fourteenOutTail(cols, hint.ToCol)
	if c1 != nil && c2 != nil {
		return i18n.Tf("fourteenout.hintLineRemoveCard",
			"col1", strconv.Itoa(hint.FromCol), "card1", cuiCardStr(c1),
			"col2", strconv.Itoa(hint.ToCol), "card2", cuiCardStr(c2)) + "\n"
	}
	// Fallback: column numbers only if a tail is unreadable (nil-guard).
	return i18n.Tf("fourteenout.hintLineRemove",
		"col1", strconv.Itoa(hint.FromCol),
		"col2", strconv.Itoa(hint.ToCol)) + "\n"
}

// fourteenOutTail safely reads a column's exposed card, returning nil for an
// out-of-range or empty column so the hint never panics.
func fourteenOutTail(cols [][]*domain.Card, c int) *domain.Card {
	if c < 0 || c >= len(cols) || len(cols[c]) == 0 {
		return nil
	}
	return cols[c][len(cols[c])-1]
}

// ActionLogOutput emits the action-log transcript as plain text.
func (pr *FourteenOutCuiPresenter) ActionLogOutput(g interfaces.FourteenOutGame) string {
	if g.GetPhase() == domain.FourteenOutPhasePlaying {
		return actionLogToText(nil)
	}
	return actionLogToText(g.GetActionLog())
}
