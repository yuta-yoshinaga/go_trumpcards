package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// canastaPlayerStr returns the display string for a single Canasta player.
func canastaPlayerStr(player *domain.CanastaPlayer, i int, showCards bool) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	fmt.Fprintf(&b, "%s: 累積%d点 ラウンド%d点 %d枚",
		name,
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
	)
	if len(player.GetRed3s()) > 0 {
		fmt.Fprintf(&b, " 赤3: %d枚", len(player.GetRed3s()))
	}
	if player.HasCanasta() {
		b.WriteString(" ★カナスタ")
	}
	b.WriteString("\n")

	// メルド表示
	for _, m := range player.GetMelds() {
		meldType := "ミックス"
		if m.IsNatural {
			meldType = "ナチュラル"
		}
		if m.IsCanasta() {
			meldType += "カナスタ"
		}
		fmt.Fprintf(&b, "  メルド[%s]: ", meldType)
		for j, c := range m.Cards {
			if j > 0 {
				b.WriteString(", ")
			}
			b.WriteString(cuiCardStr(c))
		}
		b.WriteString("\n")
	}

	if showCards && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// CanastaCuiPresenter カナスタCUIプレゼンタークラス
type CanastaCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *CanastaCuiPresenter) Output(g interfaces.CanastaGame, lastErr error) string {
	return buildCuiOutput("Canasta (カナスタ)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  山札: %d枚  捨て札: %d枚", g.GetRoundNumber(), g.GetDrawPileCount(), g.GetDiscardPileCount())
		if g.GetIsFrozen() {
			b.WriteString(" [フリーズ]")
		}
		b.WriteString("\n")

		// 捨て札トップ
		top := g.GetDiscardTop()
		if top != nil {
			fmt.Fprintf(b, "捨て札: %s\n", cuiCardStr(top))
		}

		// プレイヤー情報
		phase := g.GetPhase()
		showAllCards := phase == domain.CanastaPhaseRoundEnd || phase == domain.CanastaPhaseGameEnd
		for i := 0; i < g.GetPlayerCnt(); i++ {
			player := g.GetPlayer(i)
			showCards := player.GetIsHuman() || showAllCards
			b.WriteString(canastaPlayerStr(player, i, showCards))
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
			switch phase {
			case domain.CanastaPhaseDraw:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s (ドローフェーズ)\n", cuiPlayerName(player, currentIdx))
				b.WriteString("ds・・・山札から引く\n")
				b.WriteString("dd <idx,idx>・・・捨て札の山を取る (ナチュラルペアのインデックス)\n")
			case domain.CanastaPhaseMeld:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s (メルドフェーズ)\n", cuiPlayerName(player, currentIdx))
				b.WriteString("m <idx,idx,idx;idx,idx,idx>・・・メルドを出す (;でグループ区切り)\n")
				b.WriteString("sm・・・メルドせずにスキップ\n")
			case domain.CanastaPhaseDiscard:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s (ディスカードフェーズ)\n", cuiPlayerName(player, currentIdx))
				b.WriteString("d <idx>・・・カードを捨てる\n")
				b.WriteString("go・・・上がる (カナスタが必要)\n")
			case domain.CanastaPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *CanastaCuiPresenter) ActionLogOutput(g interfaces.CanastaGame) string {
	return actionLogOutputText(g)
}
