package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ohHellPlayerStr returns the display string for a single Oh Hell player.
func ohHellPlayerStr(player *domain.OhHellPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	bidStr := "未ビッド"
	if player.GetBid() >= 0 {
		bidStr = fmt.Sprintf("%d", player.GetBid())
	}
	fmt.Fprintf(&b, "%s: ビッド=%s 獲得%dトリック 累積%d点 ラウンド%d点 %d枚\n",
		name,
		bidStr,
		player.GetTrickCount(),
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

// OhHellCuiPresenter オー・ヘルCUIプレゼンタークラス
type OhHellCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *OhHellCuiPresenter) Output(o interfaces.OhHellGame, lastErr error) string {
	return buildCuiOutput("Oh Hell (オー・ヘル)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d/%d  手札枚数: %d  トリック: %d\n",
			o.GetRoundNumber(), o.GetTotalRounds(), o.GetHandSize(), o.GetTrickNumber())

		// 切り札表示
		trumpCard := o.GetTrumpCard()
		if trumpCard != nil {
			fmt.Fprintf(b, "切り札: %s\n", cuiCardStr(trumpCard))
		} else {
			b.WriteString("切り札: なし\n")
		}

		// ディーラー表示
		dealerIdx := o.GetDealerIdx()
		dealer := o.GetPlayer(dealerIdx)
		fmt.Fprintf(b, "ディーラー: %s\n", cuiPlayerName(dealer, dealerIdx))

		// プレイヤー情報
		for i := 0; i < o.GetPlayerCnt(); i++ {
			b.WriteString(ohHellPlayerStr(o.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// 現在のトリック
		trick := o.GetCurrentTrick()
		cuiTrickBlock(b, trick,
			func(tc *domain.OhHellTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.OhHellTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(o.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

		// ゲーム状態
		if o.GetGameEndFlag() {
			winnerIdx := o.GetWinnerIdx()
			player := o.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(cuiPlayerName(player, winnerIdx)+"の勝利です！"))
		} else {
			phase := o.GetPhase()
			switch phase {
			case domain.OhHellPhaseBid:
				bidIdx := o.GetBidPlayerIdx()
				player := o.GetPlayer(bidIdx)
				msg := fmt.Sprintf("ビッドフェーズ: %sの番", cuiPlayerName(player, bidIdx))
				restricted := o.GetRestrictedBid()
				if restricted >= 0 {
					msg += fmt.Sprintf(" (ビッド%dは不可)", restricted)
				}
				fmt.Fprintf(b, "%s\n", msg)
				b.WriteString("b <n>・・・ビッドを宣言 (0-手札枚数)\n")
			case domain.OhHellPhasePlay:
				currentIdx := o.GetCurrentPlayerIdx()
				player := o.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
			case domain.OhHellPhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("next・・・次のトリックへ\n")
			case domain.OhHellPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *OhHellCuiPresenter) HintOutput(o interfaces.OhHellGame) string {
	hint := o.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	if hint.Bid != nil {
		return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: ビッド %d を推奨 (%s)]", *hint.Bid, ohHellHintReasonStr(hint.Reason))))
	}
	if hint.CardIndex == nil {
		return "ヒントはありません。\n"
	}
	humanIdx := -1
	for i := 0; i < o.GetPlayerCnt(); i++ {
		if o.GetPlayer(i).GetIsHuman() {
			humanIdx = i
			break
		}
	}
	if humanIdx < 0 {
		return "ヒントはありません。\n"
	}
	player := o.GetPlayer(humanIdx)
	card := player.GetCard(*hint.CardIndex)
	return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: [%d]%s (%s)]", *hint.CardIndex, cuiCardStr(card), ohHellHintReasonStr(hint.Reason))))
}

// ohHellHintReasonStr ヒント理由を日本語に変換する
func ohHellHintReasonStr(reason string) string {
	return lookupHintReason(reason, nil)
}

// ActionLogOutput 棋譜をテキスト出力
func (p *OhHellCuiPresenter) ActionLogOutput(o interfaces.OhHellGame) string {
	return actionLogOutputText(o)
}
