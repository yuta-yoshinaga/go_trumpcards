package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// sevenBridgePlayerStr returns the display string for a single SevenBridge player.
func sevenBridgePlayerStr(player *domain.SevenBridgePlayer, i int) string {
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
	if player.GetMeldCount() > 0 {
		b.WriteString("  場: ")
		for mi := 0; mi < player.GetMeldCount(); mi++ {
			meld := player.GetMeld(mi)
			if mi > 0 {
				b.WriteString(" | ")
			}
			for ci, c := range meld {
				if ci > 0 {
					b.WriteString(" ")
				}
				b.WriteString(cuiCardStr(c))
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// SevenBridgeCuiPresenter セブンブリッジ CUI プレゼンター
type SevenBridgeCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *SevenBridgeCuiPresenter) Output(g interfaces.SevenBridgeGame, lastErr error) string {
	return buildCuiOutput("Seven Bridge (セブンブリッジ)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  山札: %d枚\n", g.GetRoundNumber(), g.GetDrawPileCount())

		top := g.GetDiscardTop()
		if top != nil {
			fmt.Fprintf(b, "捨て札: %s\n", cuiCardStr(top))
		}

		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(sevenBridgePlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			player := g.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(cuiPlayerName(player, winnerIdx)+"の勝利です！"))
			return
		}

		phase := g.GetPhase()
		switch phase {
		case domain.SevenBridgePhaseDraw:
			currentIdx := g.GetCurrentPlayerIdx()
			player := g.GetPlayer(currentIdx)
			fmt.Fprintf(b, "手番: %s (ドローフェーズ)\n", cuiPlayerName(player, currentIdx))
			b.WriteString("ds・・・山札から引く\n")
			b.WriteString("pon <idx,idx>・・・ポンで捨て札を取得\n")
			b.WriteString("chi <idx,idx>・・・チーで捨て札を取得\n")
		case domain.SevenBridgePhasePlay:
			currentIdx := g.GetCurrentPlayerIdx()
			player := g.GetPlayer(currentIdx)
			fmt.Fprintf(b, "手番: %s (プレイフェーズ)\n", cuiPlayerName(player, currentIdx))
			b.WriteString("m <idx,idx,idx[,...]>・・・メルドを場に出す\n")
			b.WriteString("lo <pIdx> <mIdx> <cIdx>・・・既存メルドにカード追加\n")
			b.WriteString("d <idx>・・・カードを捨てる\n")
		case domain.SevenBridgePhaseRoundEnd:
			b.WriteString("ラウンド終了\n")
			b.WriteString("nr / nextround・・・次のラウンドへ\n")
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *SevenBridgeCuiPresenter) ActionLogOutput(g interfaces.SevenBridgeGame) string {
	return actionLogOutputText(g)
}
