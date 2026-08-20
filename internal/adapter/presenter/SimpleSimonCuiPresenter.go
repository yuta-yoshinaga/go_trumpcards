//go:build !js || !wasm || classic

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// SimpleSimonCuiPresenter シンプル・サイモンのCUIプレゼンタークラス。
type SimpleSimonCuiPresenter struct{}

// ssColumnStr renders a column as its cards (or the shared empty marker).
func ssColumnStr(pile []*domain.Card) string {
	if len(pile) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	// **まとめて動かせるのは末尾から続く run だけ** (#5679)。Web はその起点に
	// リングを付けているのに、CUI は平らに並べるだけで毎回目視させていた。
	// 列全体が run なら区切る意味がないので入れない。
	from := domain.SimpleSimonMovableFrom(pile)
	parts := make([]string, 0, len(pile)+1)
	for i, c := range pile {
		if i == from && from > 0 {
			parts = append(parts, CuiRunMark)
		}
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}

// Output renders the current game state.
func (p *SimpleSimonCuiPresenter) Output(g interfaces.SimpleSimonGame, lastErr error) string {
	return buildCuiOutput(i18n.T("simplesimon.helpTitle"), func(sb *strings.Builder) {
		cols := g.GetColumns()
		for i := 0; i < domain.SimpleSimonColCnt; i++ {
			sb.WriteString(i18n.Tf("simplesimon.colLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + ssColumnStr(cols[i]) + "\n")
		}
		sb.WriteString(i18n.Tf("simplesimon.completedLine", "count", strconv.Itoa(g.GetCompletedSuits())) + "\n")

		cuiErrorBlock(sb, lastErr)

		switch g.GetPhase() {
		case domain.SimpleSimonPhasePlaying:
			sb.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) +
				cuiSolitaireUndoHint(g.CanUndo()) + "\n")
		case domain.SimpleSimonPhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.SimpleSimonPhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *SimpleSimonCuiPresenter) HintOutput(g interfaces.SimpleSimonGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return color.Yellow(i18n.Tf("simplesimon.hintLine",
		"from", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"to", strconv.Itoa(hint.ToCol))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *SimpleSimonCuiPresenter) ActionLogOutput(g interfaces.SimpleSimonGame) string {
	return actionLogOutputText(g)
}
