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

// BlackHoleCuiPresenter ブラックホールのCUIプレゼンタークラス。
type BlackHoleCuiPresenter struct{}

// bhPileStr renders a pile as its cards (or the shared empty marker).
func bhPileStr(pile []*domain.Card) string {
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
func (p *BlackHoleCuiPresenter) Output(g interfaces.BlackHoleGame, lastErr error) string {
	return buildCuiOutput(i18n.T("blackhole.helpTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.T("blackhole.blackHoleLabel") + " " + bhPileStr(g.GetBlackHole()) + "\n")
		// **17 個の扇を掘り進める長いゲーム**なのに、あと何枚で終わるかがどこにも
		// 出ていなかった (#5681)。勝利条件は 52 枚すべてを吸い込むこと。
		sb.WriteString(i18n.Tf("blackhole.progress",
			"count", strconv.Itoa(len(g.GetBlackHole())),
			"total", strconv.Itoa(domain.BlackHoleTotalCards)) + "\n")
		for i, fan := range g.GetFans() {
			sb.WriteString(i18n.Tf("blackhole.fanLabel", "idx", strconv.Itoa(i)))
			sb.WriteString(" " + bhPileStr(fan) + "\n")
		}

		// **±1 を暗算させない。**Web は「出せるランク」と「残り合法手」を常時
		// 出しているのに、CUI は穴のトップと扇一覧しか出していなかった (#4818)。
		if g.GetPhase() == domain.BlackHolePhasePlaying {
			ranks := g.AcceptableRanks()
			if len(ranks) > 0 {
				labels := make([]string, 0, len(ranks))
				for _, r := range ranks {
					labels = append(labels, cuiRankLabel(r))
				}
				sb.WriteString(i18n.Tf("blackhole.acceptableRanks",
					"ranks", strings.Join(labels, " / ")) + "\n")
			} else {
				sb.WriteString(i18n.T("blackhole.acceptableRanksNone") + "\n")
			}
			playable := g.PlayableFans()
			line := i18n.Tf("blackhole.legalMoveCount", "count", strconv.Itoa(len(playable)))
			if len(playable) == 0 {
				line = color.Red(line)
			}
			sb.WriteString(line + "\n")
		}

		cuiErrorBlock(sb, lastErr)

		switch g.GetPhase() {
		case domain.BlackHolePhasePlaying:
			sb.WriteString(i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) +
				cuiSolitaireUndoHint(g.CanUndo()) + "\n")
		case domain.BlackHolePhaseGameClear:
			sb.WriteString(color.Green(i18n.T("cuiSolitaireGameClear")) + " " +
				i18n.Tf("cuiSolitaireMoves", "count", strconv.Itoa(g.GetMoveCount())) + "\n")
		case domain.BlackHolePhaseGameOver:
			sb.WriteString(color.Red(i18n.T("cuiSolitaireGameOver")) + "\n")
		}
	})
}

// HintOutput emits the current hint.
func (p *BlackHoleCuiPresenter) HintOutput(g interfaces.BlackHoleGame) string {
	hint := g.GetHint()
	if hint == nil {
		return i18n.T("cuiHintNone") + "\n"
	}
	return color.Yellow(i18n.Tf("blackhole.hintLine", "fan", strconv.Itoa(hint.Fan))) + "\n"
}

// ActionLogOutput emits the action-log transcript as plain text.
func (p *BlackHoleCuiPresenter) ActionLogOutput(g interfaces.BlackHoleGame) string {
	return actionLogOutputText(g)
}
