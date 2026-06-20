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

// Output renders the current game state.
func (p *LaBelleLucieCuiPresenter) Output(g interfaces.LaBelleLucieGame, lastErr error) string {
	return buildCuiOutput(i18n.T("labellelucie.helpTitle"), func(sb *strings.Builder) {
		foundation := g.GetFoundation()
		for i := 0; i < domain.LaBelleLucieFoundationCnt; i++ {
			sb.WriteString(i18n.Tf("labellelucie.foundationLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + llPileStr(foundation[i]) + "\n")
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
