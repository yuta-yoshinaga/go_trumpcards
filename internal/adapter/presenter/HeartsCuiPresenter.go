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
		cuiTrickBlock(b, trick,
			func(tc *domain.HeartsTrickCard) int { return tc.PlayerIdx },
			func(tc *domain.HeartsTrickCard) string { return cuiCardStr(tc.Card) },
			func(idx int) string { return cuiPlayerName(h.GetPlayer(idx), idx) },
		)

		cuiErrorBlock(b, lastErr)

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

// heartsHintReasons はHearts固有のヒント理由翻訳
var heartsHintReasons = map[string]string{
	"pass_high_risk_cards": "リスクの高いカードを渡す",
	"discard_queen_spades": "Q♠を捨てるチャンス",
	"discard_hearts":       "ハートを捨てる",
}

// heartsHintReasonStr ヒント理由を日本語に変換する
func heartsHintReasonStr(reason string) string {
	return lookupHintReason(reason, heartsHintReasons)
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
