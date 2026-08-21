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

// CurdsAndWheyCuiPresenter カーズ・アンド・ホエイのCUIプレゼンタークラス。
type CurdsAndWheyCuiPresenter struct{}

// ssColumnStrCW renders a column as its cards (or the shared empty marker).
func ssColumnStrCW(pile []*domain.Card) string {
	if len(pile) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	// **まとめて動かせるのは末尾から続く run だけ** (#5679)。Web はその起点に
	// リングを付けているのに、CUI は平らに並べるだけで毎回目視させていた。
	// 列全体が run なら区切る意味がないので入れない。
	from := domain.CurdsAndWheyMovableFrom(pile)
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
func (p *CurdsAndWheyCuiPresenter) Output(g interfaces.CurdsAndWheyGame, lastErr error) string {
	return buildCuiOutput(i18n.T("curdsandwhey.helpTitle"), func(sb *strings.Builder) {
		cols := g.GetColumns()
		for i := 0; i < domain.CurdsAndWheyColCnt; i++ {
			sb.WriteString(i18n.Tf("curdsandwhey.colLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + ssColumnStrCW(cols[i]) + "\n")
		}
		sb.WriteString(i18n.Tf("curdsandwhey.completedLine", "count", strconv.Itoa(g.GetCompletedSuits())) + "\n")

		cuiErrorBlock(sb, lastErr)

		switch g.GetPhase() {
		case domain.CurdsAndWheyPhasePlaying:
			sb.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) +
				cuiSolitaireUndoHint(g.CanUndo()) + "\n")
		case domain.CurdsAndWheyPhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.CurdsAndWheyPhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *CurdsAndWheyCuiPresenter) HintOutput(g interfaces.CurdsAndWheyGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return color.Yellow(i18n.Tf("curdsandwhey.hintLine",
		"from", strconv.Itoa(hint.FromCol),
		"idx", strconv.Itoa(hint.CardIndex),
		"to", strconv.Itoa(hint.ToCol))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *CurdsAndWheyCuiPresenter) ActionLogOutput(g interfaces.CurdsAndWheyGame) string {
	return actionLogOutputText(g)
}
