package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// heartsPlayerStr returns the display string for a single Hearts player.
func heartsPlayerStr(player *domain.HeartsPlayer, i int) string {
	var b strings.Builder
	name := cuiPlayerName(player, i)
	fmt.Fprintf(&b, "%s: 累積%d点 ラウンド%d点 %d枚 %dトリック\n",
		name,
		player.GetCumulativeScore(),
		player.GetRoundScore(),
		player.GetCardsSize(),
		player.GetTrickCount(),
	)
	if player.GetIsHuman() && player.GetCardsSize() > 0 {
		b.WriteString(cuiIndexedCardListStr(player))
		b.WriteString("\n")
	}
	return b.String()
}

// HeartsCuiPresenter ハーツCUIプレゼンタークラス
type HeartsCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *HeartsCuiPresenter) Output(h interfaces.HeartsGame, lastErr error) string {
	return buildCuiOutput("Hearts (ハーツ)", func(b *strings.Builder) {
		fmt.Fprintf(b, "ラウンド: %d  トリック: %d\n", h.GetRoundNumber(), h.GetTrickNumber())

		if h.GetHeartsBroken() {
			b.WriteString("ハートブレイク: あり\n")
		} else {
			b.WriteString("ハートブレイク: なし\n")
		}

		// プレイヤー情報
		for i := 0; i < h.GetPlayerCnt(); i++ {
			b.WriteString(heartsPlayerStr(h.GetPlayer(i), i))
		}

		b.WriteString("----------\n")

		// 現在のトリック
		trick := h.GetCurrentTrick()
		if len(trick) > 0 {
			parts := make([]string, len(trick))
			for i, tc := range trick {
				player := h.GetPlayer(tc.PlayerIdx)
				parts[i] = fmt.Sprintf("%s=%s", cuiPlayerName(player, tc.PlayerIdx), cuiCardStr(tc.Card))
			}
			fmt.Fprintf(b, "トリック: %s\n", strings.Join(parts, ", "))
		}

		// エラーメッセージ
		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		// ゲーム状態
		if h.GetGameEndFlag() {
			winnerIdx := h.GetWinnerIdx()
			player := h.GetPlayer(winnerIdx)
			fmt.Fprintf(b, "ゲーム終了！ %s\n", color.Green(cuiPlayerName(player, winnerIdx)+"の勝利です！"))
		} else {
			phase := h.GetPhase()
			switch phase {
			case domain.HeartsPhasePass:
				dir := h.GetPassDirection()
				fmt.Fprintf(b, "パスフェーズ: %s\n", cuiPassDirectionStr(dir))
				b.WriteString("pass <idx> <idx> <idx>・・・3枚のカードを選択\n")
			case domain.HeartsPhasePlay:
				currentIdx := h.GetCurrentPlayerIdx()
				player := h.GetPlayer(currentIdx)
				fmt.Fprintf(b, "手番: %s\n", cuiPlayerName(player, currentIdx))
				b.WriteString("play <idx>・・・カードを出す\n")
			case domain.HeartsPhaseTrickEnd:
				b.WriteString("トリック終了\n")
				b.WriteString("next・・・次のトリックへ\n")
			case domain.HeartsPhaseRoundEnd:
				b.WriteString("ラウンド終了\n")
				b.WriteString("nr / nextround・・・次のラウンドへ\n")
			}
		}
	})
}

// HintOutput ヒント情報を出力する
func (p *HeartsCuiPresenter) HintOutput(h interfaces.HeartsGame) string {
	hint := h.GetHint()
	if hint == nil {
		return "ヒントはありません。\n"
	}
	player := h.GetPlayer(0)
	cards := make([]string, len(hint.CardIndices))
	for i, idx := range hint.CardIndices {
		cards[i] = fmt.Sprintf("[%d]%s", idx, cuiCardStr(player.GetCard(idx)))
	}
	return fmt.Sprintf("%s\n", color.Yellow(fmt.Sprintf("[HINT: %s (%s)]", strings.Join(cards, ", "), heartsHintReasonStr(hint.Reason))))
}

// heartsHintReasonStr ヒント理由を日本語に変換する
func heartsHintReasonStr(reason string) string {
	switch reason {
	case "pass_high_risk_cards":
		return "リスクの高いカードを渡す"
	case "lead_low":
		return "低いカードでリード"
	case "follow_suit":
		return "リードスートに追随"
	case "discard_queen_spades":
		return "Q♠を捨てるチャンス"
	case "discard_hearts":
		return "ハートを捨てる"
	case "discard_high":
		return "高いカードを捨てる"
	default:
		return reason
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (p *HeartsCuiPresenter) ActionLogOutput(h interfaces.HeartsGame) string {
	return actionLogOutputText(h)
}

// cuiPassDirectionStr パス方向の日本語表示
func cuiPassDirectionStr(dir domain.HeartsPassDirection) string {
	switch dir {
	case domain.HeartsPassLeft:
		return "左へ渡す"
	case domain.HeartsPassRight:
		return "右へ渡す"
	case domain.HeartsPassAcross:
		return "向かいへ渡す"
	case domain.HeartsPassNone:
		return "交換なし"
	default:
		return "不明"
	}
}
