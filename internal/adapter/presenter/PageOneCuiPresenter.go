package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// pageOnePlayerStr returns the display string for a single PageOne player.
func pageOnePlayerStr(player *domain.PageOnePlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	declared := ""
	if player.GetHasDeclared() {
		declared = " [PAGE ONE!]"
	}
	fmt.Fprintf(&b, "%s: 累積%d点 ラウンド%d点 %d枚%s\n",
		name,
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
		declared,
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
	return buildCuiOutput("Page One (ページワン)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  山札: %d枚\n", g.GetRoundNumber(), g.GetDrawPileCount())

		top := g.GetDiscardTop()
		if top != nil {
			fmt.Fprintf(b, "捨て札: %s\n", cuiCardStr(top))
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
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(cuiPlayerName(player, winnerIdx)+"の勝利です！"))
		} else {
			phase := g.GetPhase()
			switch phase {
			case domain.PageOnePhasePlay:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
				b.WriteString("draw・・・カードを引く\n")
			case domain.PageOnePhaseMustDeclare:
				b.WriteString("宣言フェーズ: 手札が残り1枚です！\n")
				b.WriteString("declare・・・「ページワン！」と宣言する\n")
				b.WriteString("skip・・・宣言をスキップ（ペナルティ: 2枚引く）\n")
			case domain.PageOnePhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *PageOneCuiPresenter) ActionLogOutput(g interfaces.PageOneGame) string {
	return actionLogOutputText(g)
}
