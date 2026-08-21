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

// ShamrocksCuiPresenter シャムロックスのCUIプレゼンタークラス。
type ShamrocksCuiPresenter struct{}

// llPileStrSH renders a pile as its cards (or the shared empty marker).
func llPileStrSH(pile []*domain.Card) string {
	if len(pile) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	parts := make([]string, len(pile))
	for i, c := range pile {
		parts[i] = cuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// llFoundationStrSH renders a foundation compactly as its top card plus a count,
// since late-game foundations grow to 10+ cards and full listing is noisy.
func llFoundationStrSH(pile []*domain.Card) string {
	if len(pile) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	return "[" + cuiCardStr(pile[len(pile)-1]) + "] " +
		i18n.Tf("shamrocks.foundationCount", "count", strconv.Itoa(len(pile)))
}

// Output renders the current game state.
func (p *ShamrocksCuiPresenter) Output(g interfaces.ShamrocksGame, lastErr error) string {
	return buildCuiOutput(i18n.T("shamrocks.helpTitle"), func(sb *strings.Builder) {
		foundation := g.GetFoundation()
		for i := 0; i < domain.ShamrocksFoundationCnt; i++ {
			sb.WriteString(i18n.Tf("shamrocks.foundationLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + llFoundationStrSH(foundation[i]) + "\n")
		}
		for i, fan := range g.GetFans() {
			sb.WriteString(i18n.Tf("shamrocks.fanLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + llPileStrSH(fan) + "\n")
		}

		cuiErrorBlock(sb, lastErr)

		switch g.GetPhase() {
		case domain.ShamrocksPhasePlaying:
			sb.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) +
				cuiSolitaireUndoHint(g.CanUndo()) + "\n")
			// **Shamrocks に再配りは無い**ので、合法手が尽きたらそれで終わり。
			// La Belle Lucie はここで「残り再配り n 回」を見て再配りを勧めるが、
			// この分岐はこのゲームでは常に「真の手詰まり」側になる (#4769 の
			// 「何も言わずに探させ続ける」問題は同じなので、表示自体は残す)。
			if !g.HasAnyLegalMove() {
				sb.WriteString(color.Red(i18n.T("shamrocks.stuckDeadlock")) + "\n")
			}
		case domain.ShamrocksPhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.ShamrocksPhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *ShamrocksCuiPresenter) HintOutput(g interfaces.ShamrocksGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	to := i18n.Tf("shamrocks.hintToFan", "idx", strconv.Itoa(hint.ToFan))
	if hint.ToFoundation {
		to = i18n.T("shamrocks.hintToFoundation")
	}
	return color.Yellow(i18n.Tf("shamrocks.hintLine",
		"from", strconv.Itoa(hint.FromFan), "to", to)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *ShamrocksCuiPresenter) ActionLogOutput(g interfaces.ShamrocksGame) string {
	return actionLogOutputText(g)
}
