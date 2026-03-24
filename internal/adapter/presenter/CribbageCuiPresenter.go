package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// cribbagePlayerStr returns the display string for a single Cribbage player.
func cribbagePlayerStr(player *domain.CribbagePlayer, i int, dealerIdx int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	dealerMark := ""
	if i == dealerIdx {
		dealerMark = " [D]"
	}
	fmt.Fprintf(&b, "%s%s: 累積%d点 ラウンド%d点 %d枚\n",
		name,
		dealerMark,
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

// CribbageCuiPresenter クリベッジCUIプレゼンタークラス
type CribbageCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *CribbageCuiPresenter) Output(g interfaces.CribbageGame, lastErr error) string {
	return buildCuiOutput("Cribbage (クリベッジ)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  ディーラー: Player%d\n", g.GetRoundNumber(), g.GetDealerIdx())

		// スターターカード
		starter := g.GetStarter()
		if starter != nil {
			fmt.Fprintf(b, "スターター: %s\n", cuiCardStr(starter))
		}

		// ペギング情報
		phase := g.GetPhase()
		if phase == domain.CribbagePhasePegging {
			fmt.Fprintf(b, "ペギング合計: %d/31\n", g.GetPegCount())
			pegCards := g.GetPegPlayedCards()
			if len(pegCards) > 0 {
				b.WriteString("出されたカード: ")
				for i, c := range pegCards {
					if i > 0 {
						b.WriteString(", ")
					}
					b.WriteString(cuiCardStr(c))
				}
				b.WriteString("\n")
			}
		}

		// プレイヤー情報
		for i := range domain.CribbagePlayerCnt {
			player := g.GetPlayer(i)
			if player != nil {
				b.WriteString(cribbagePlayerStr(player, i, g.GetDealerIdx()))
			}
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
			case domain.CribbagePhaseDiscard:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s (ディスカードフェーズ)\n", cuiPlayerName(player, currentIdx))
				b.WriteString("d <idx,idx>・・・クリブに2枚捨てる\n")
			case domain.CribbagePhasePegging:
				currentIdx := g.GetCurrentPlayerIdx()
				player := g.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s (ペギングフェーズ)\n", cuiPlayerName(player, currentIdx))
				b.WriteString("p <idx>・・・カードを出す\n")
				b.WriteString("go・・・Goを宣言する\n")
			case domain.CribbagePhaseShow:
				b.WriteString("ショーフェーズ\n")
				p.writeShowDetails(b, g)
				b.WriteString("sn / shownext・・・次のスコア計算\n")
			case domain.CribbagePhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				p.writeShowDetails(b, g)
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// writeShowDetails ショーフェーズのスコア詳細を表示
func (p *CribbageCuiPresenter) writeShowDetails(b *strings.Builder, g interfaces.CribbageGame) {
	details := g.GetHandScoreDetails()
	labels := [3]string{"非ディーラー手札", "ディーラー手札", "クリブ"}
	for i, d := range details {
		if d != nil {
			fmt.Fprintf(b, "  %s: %d点 (15s=%d, ペア=%d, ラン=%d, フラッシュ=%d, ノブ=%d)\n",
				labels[i], d.Total, d.Fifteens, d.Pairs, d.Runs, d.Flush, d.Nobs)
		}
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (p *CribbageCuiPresenter) ActionLogOutput(g interfaces.CribbageGame) string {
	return actionLogOutputText(g)
}
