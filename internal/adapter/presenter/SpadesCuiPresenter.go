package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// spadesPlayerStr returns the display string for a single Spades player.
func spadesPlayerStr(player *domain.SpadesPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	bidStr := "未ビッド"
	if player.GetBid() >= 0 {
		bidStr = fmt.Sprintf("%d", player.GetBid())
	}
	fmt.Fprintf(&b, "%s: ビッド=%s 獲得%dトリック バッグ%d 累積%d点 ラウンド%d点 %d枚\n",
		name,
		bidStr,
		player.GetTrickCount(),
		player.GetBags(),
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

// SpadesCuiPresenter スペードCUIプレゼンタークラス
type SpadesCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *SpadesCuiPresenter) Output(s interfaces.SpadesGame, lastErr error) string {
	return buildCuiOutput("Spades (スペード)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d\n", s.GetRoundNumber(), s.GetTrickNumber())

		if s.GetSpadesBroken() {
			b.WriteString("スペードブレイク: あり\n")
		} else {
			b.WriteString("スペードブレイク: なし\n")
		}

		// プレイヤー情報
		for i := 0; i < s.GetPlayerCnt(); i++ {
			b.WriteString(spadesPlayerStr(s.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// 現在のトリック
		trick := s.GetCurrentTrick()
		if len(trick) > 0 {
			parts := make([]string, len(trick))
			for i, tc := range trick {
				player := s.GetPlayer(tc.PlayerIdx)
				parts[i] = fmt.Sprintf("%s=%s", cuiPlayerName(player, tc.PlayerIdx), cuiCardStr(tc.Card))
			}
			fmt.Fprintf(b, "トリック: %s\n", strings.Join(parts, ", "))
		}

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		if s.GetGameEndFlag() {
			winnerIdx := s.GetWinnerIdx()
			player := s.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(cuiPlayerName(player, winnerIdx)+"の勝利です！"))
		} else {
			phase := s.GetPhase()
			switch phase {
			case domain.SpadesPhaseBid:
				bidIdx := s.GetBidPlayerIdx()
				player := s.GetPlayer(bidIdx)
				fmt.Fprintf(b, "ビッドフェーズ: %sの番\n", cuiPlayerName(player, bidIdx))
				b.WriteString("b <n>・・・ビッドを宣言 (0=ニル, 1-13)\n")
			case domain.SpadesPhasePlay:
				currentIdx := s.GetCurrentPlayerIdx()
				player := s.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
			case domain.SpadesPhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("next・・・次のトリックへ\n")
			case domain.SpadesPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("next・・・次のラウンドへ\n")
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *SpadesCuiPresenter) ActionLogOutput(s interfaces.SpadesGame) string {
	return actionLogOutputText(s)
}
