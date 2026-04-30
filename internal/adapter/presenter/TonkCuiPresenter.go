package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// tonkPlayerStr returns the display string for a single Tonk player.
func tonkPlayerStr(player *domain.TonkPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	fmt.Fprintf(&b, "%s: 累積%d点 ラウンド%d点 %d枚\n",
		name,
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// TonkCuiPresenter Tonk CUIプレゼンタークラス
type TonkCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *TonkCuiPresenter) Output(g interfaces.TonkGame, lastErr error) string {
	return buildCuiOutput("Tonk (トンク)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  山札: %d枚\n", g.GetRoundNumber(), g.GetDrawPileCount())

		top := g.GetDiscardTop()
		if top != nil {
			fmt.Fprintf(b, "捨て札: %s\n", cuiCardStr(top))
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(tonkPlayerStr(g.GetPlayer(i), i))
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
			case domain.TonkPhaseDraw:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s (ドローフェーズ)\n", cuiPlayerName(player, currentIdx))
				b.WriteString("ds・・・山札から引く\n")
				b.WriteString("dd・・・捨て札から引く\n")
			case domain.TonkPhaseDiscard:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s (ディスカードフェーズ)\n", cuiPlayerName(player, currentIdx))
				b.WriteString("d <idx>・・・カードを捨てる\n")
				b.WriteString("k <idx>・・・ノック (デッドウッド5点以下で宣言)\n")
			case domain.TonkPhaseRoundEnd:
				if g.GetIsTonk() {
					b.WriteString("配牌Tonk成立！\n")
				}
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *TonkCuiPresenter) ActionLogOutput(g interfaces.TonkGame) string {
	return actionLogOutputText(g)
}
