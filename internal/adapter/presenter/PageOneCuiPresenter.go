package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// pageOnePlayerStr returns the display string for a single PageOne player.
func pageOnePlayerStr(player *domain.PageOnePlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	declared := ""
	if player.GetHasDeclared() {
		declared = " " + i18n.T("pageone.declaredBadge")
	}
	// Warn when a player sits on their last card without declaring "Page One",
	// mirroring the web cpuAtOne badge (mutually exclusive with declaredBadge).
	warning := ""
	if !player.GetHasDeclared() && player.GetCardsSize() == 1 {
		warning = " " + color.Red(i18n.T("pageone.lastCardWarning"))
	}
	fmt.Fprintf(&b, "%s\n",
		i18n.Tf("pageone.playerLine",
			"name", name,
			"cumulative", fmt.Sprintf("%d", player.GetCumulativeScore()),
			"round", fmt.Sprintf("%d", player.GetRoundScore()),
			"cards", fmt.Sprintf("%d", player.GetCardsSize()),
			"declared", declared,
			"warning", warning,
		),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// PageOneCuiPresenter ページワンCUIプレゼンタークラス
type PageOneCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *PageOneCuiPresenter) Output(g interfaces.PageOneGame, lastErr error) string {
	return buildCuiOutput(i18n.T("pageone.helpTitle"), func(b *strings.Builder) {
		fmt.Fprintf(b, "%s\n", i18n.Tf("pageone.header",
			"round", fmt.Sprintf("%d", g.GetRoundNumber()),
			"drawPile", fmt.Sprintf("%d", g.GetDrawPileCount()),
		))

		top := g.GetDiscardTop()
		if top != nil {
			fmt.Fprintf(b, "%s\n", i18n.Tf("pageone.discardLine", "card", cuiCardStr(top)))
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(pageOnePlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			player := g.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "%s\n", color.Green(i18n.Tf("pageone.gameEnd", "name", cuiPlayerName(player, winnerIdx))))
		} else {
			phase := g.GetPhase()
			switch phase {
			case domain.PageOnePhasePlay:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "%s\n", i18n.Tf("pageone.turnLine", "name", cuiPlayerName(player, currentIdx)))
				if top := g.GetDiscardTop(); top != nil {
					b.WriteString(i18n.Tf("pageone.cuiPlayCondition", "card", cuiCardStr(top)) + "\n")
				}
				b.WriteString(i18n.T("pageone.cmdPlay") + "\n")
				b.WriteString(i18n.T("pageone.cmdDraw") + "\n")
			case domain.PageOnePhaseMustDeclare:
				b.WriteString(color.Yellow(i18n.T("pageone.cuiMustDeclare")) + "\n")
				b.WriteString(i18n.T("pageone.declarePhase") + "\n")
				b.WriteString(i18n.T("pageone.cmdDeclare") + "\n")
				b.WriteString(i18n.T("pageone.cmdSkip") + "\n")
			case domain.PageOnePhaseRoundEnd:
				b.WriteString(i18n.T("pageone.roundEnd") + "\n")
				b.WriteString(i18n.T("pageone.cmdNextRound") + "\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *PageOneCuiPresenter) ActionLogOutput(g interfaces.PageOneGame) string {
	return actionLogOutputText(g)
}
