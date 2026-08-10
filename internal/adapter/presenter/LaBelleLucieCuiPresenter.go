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

// LaBelleLucieCuiPresenter ラ・ベル・ルーシーのCUIプレゼンタークラス。
type LaBelleLucieCuiPresenter struct{}

// llPileStr renders a pile as its cards (or the shared empty marker).
func llPileStr(pile []*domain.Card) string {
	if len(pile) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	parts := make([]string, len(pile))
	for i, c := range pile {
		parts[i] = cuiCardStr(c)
	}
	return strings.Join(parts, " ")
}

// llFoundationStr renders a foundation compactly as its top card plus a count,
// since late-game foundations grow to 10+ cards and full listing is noisy.
func llFoundationStr(pile []*domain.Card) string {
	if len(pile) == 0 {
		return i18n.T("cuiEmptyCol")
	}
	return "[" + cuiCardStr(pile[len(pile)-1]) + "] " +
		i18n.Tf("labellelucie.foundationCount", "count", strconv.Itoa(len(pile)))
}

// Output renders the current game state.
func (p *LaBelleLucieCuiPresenter) Output(g interfaces.LaBelleLucieGame, lastErr error) string {
	return buildCuiOutput(i18n.T("labellelucie.helpTitle"), func(sb *strings.Builder) {
		foundation := g.GetFoundation()
		for i := 0; i < domain.LaBelleLucieFoundationCnt; i++ {
			sb.WriteString(i18n.Tf("labellelucie.foundationLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + llFoundationStr(foundation[i]) + "\n")
		}
		for i, fan := range g.GetFans() {
			sb.WriteString(i18n.Tf("labellelucie.fanLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + llPileStr(fan) + "\n")
		}
		sb.WriteString(i18n.Tf("labellelucie.redealsLine", "count", strconv.Itoa(g.GetRedealsLeft())) + "\n")

		cuiErrorBlock(sb, lastErr)

		switch g.GetPhase() {
		case domain.LaBelleLuciePhasePlaying:
			sb.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
			// No legal move but redeals remain -> recommend a redeal, mirroring the
			// web stuck banner.
			if !g.HasAnyLegalMove() {
				// **再配札が尽きた真の手詰まりは別物 (#4769)。**Web は
				// ll-deadlock-banner を出して giveup を点滅させるのに、CUI は
				// 何も言わず、合法手が無いまま延々と手を探させていた。
				if g.GetRedealsLeft() > 0 {
					sb.WriteString(color.Yellow(i18n.T("labellelucie.redealRecommended")) + "\n")
				} else {
					sb.WriteString(color.Red(i18n.T("labellelucie.stuckDeadlock")) + "\n")
				}
			}
		case domain.LaBelleLuciePhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.LaBelleLuciePhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *LaBelleLucieCuiPresenter) HintOutput(g interfaces.LaBelleLucieGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	to := i18n.Tf("labellelucie.hintToFan", "idx", strconv.Itoa(hint.ToFan))
	if hint.ToFoundation {
		to = i18n.T("labellelucie.hintToFoundation")
	}
	return color.Yellow(i18n.Tf("labellelucie.hintLine",
		"from", strconv.Itoa(hint.FromFan), "to", to)) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *LaBelleLucieCuiPresenter) ActionLogOutput(g interfaces.LaBelleLucieGame) string {
	return actionLogOutputText(g)
}
