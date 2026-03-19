package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// crazyEightsPlayerStr returns the display string for a single CrazyEights player.
func crazyEightsPlayerStr(player *domain.CrazyEightsPlayer, i int) string {
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

// CrazyEightsCuiPresenter クレイジーエイトCUIプレゼンタークラス
type CrazyEightsCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *CrazyEightsCuiPresenter) Output(g interfaces.CrazyEightsGame, lastErr error) string {
	return buildCuiOutput("Crazy Eights (クレイジーエイト)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  山札: %d枚\n", g.GetRoundNumber(), g.GetDrawPileCount())

		// 捨て札トップ
		top := g.GetDiscardTop()
		if top != nil {
			fmt.Fprintf(b, "捨て札: %s", cuiCardStr(top))
			if g.GetChosenSuit() > 0 {
				fmt.Fprintf(b, " (指定スート: %s)", suitDisplayName(g.GetChosenSuit()))
			}
			b.WriteString("\n")
		}

		// プレイヤー情報
		for i := 0; i < g.GetPlayerCnt(); i++ {
			b.WriteString(crazyEightsPlayerStr(g.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		if g.GetGameEndFlag() {
			winnerIdx := g.GetWinnerIdx()
			player := g.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(cuiPlayerName(player, winnerIdx)+"の勝利です！"))
		} else {
			phase := g.GetPhase()
			switch phase {
			case domain.CrazyEightsPhasePlay:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
				b.WriteString("draw・・・カードを引く\n")
			case domain.CrazyEightsPhaseChooseSuit:
				b.WriteString("スート選択フェーズ\n")
				b.WriteString("suit <1-4>・・・スートを選択 (1=♠, 2=♣, 3=♥, 4=♦)\n")
			case domain.CrazyEightsPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *CrazyEightsCuiPresenter) ActionLogOutput(g interfaces.CrazyEightsGame) string {
	return actionLogOutputText(g)
}

// suitDisplayName スート表示名を返す
func suitDisplayName(suit int) string {
	switch suit {
	case domain.CardDesignSpade:
		return "♠"
	case domain.CardDesignClover:
		return "♣"
	case domain.CardDesignHeart:
		return "♥"
	case domain.CardDesignDiamond:
		return "♦"
	default:
		return "?"
	}
}
