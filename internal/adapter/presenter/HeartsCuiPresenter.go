package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// HeartsCuiPresenter ハーツCUIプレゼンタークラス
type HeartsCuiPresenter struct{}

// NewHeartsCuiPresenter コンストラクタ
func NewHeartsCuiPresenter() *HeartsCuiPresenter {
	return &HeartsCuiPresenter{}
}

// Output ゲーム状態を文字列出力
func (p *HeartsCuiPresenter) Output(h interfaces.HeartsGame, lastErr error) string {
	var b strings.Builder

	b.WriteString("==========\n")
	b.WriteString("Hearts (ハーツ)\n")
	b.WriteString("==========\n")

	fmt.Fprintf(&b, "ラウンド: %d  トリック: %d\n", h.GetRoundNumber(), h.GetTrickNumber())

	if h.GetHeartsBroken() {
		b.WriteString("ハートブレイク: あり\n")
	} else {
		b.WriteString("ハートブレイク: なし\n")
	}

	// プレイヤー情報
	for i := 0; i < h.GetPlayerCnt(); i++ {
		player := h.GetPlayer(i)
		name := cuiPlayerName(player, i)
		fmt.Fprintf(&b, "%s: 累積%d点 ラウンド%d点 %d枚 %dトリック\n",
			name,
			player.GetCumulativeScore(),
			player.GetRoundScore(),
			player.GetCardsSize(),
			player.GetTrickCount(),
		)
		if player.GetIsHuman() {
			for j := 0; j < player.GetCardsSize(); j++ {
				if j != 0 {
					b.WriteString("  ")
				}
				fmt.Fprintf(&b, "[%d]%s", j, cuiCardStr(player.GetCard(j)))
			}
			if player.GetCardsSize() > 0 {
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("----------\n")

	// 現在のトリック
	trick := h.GetCurrentTrick()
	if len(trick) > 0 {
		b.WriteString("トリック: ")
		for i, tc := range trick {
			if i != 0 {
				b.WriteString(", ")
			}
			player := h.GetPlayer(tc.PlayerIdx)
			fmt.Fprintf(&b, "%s=%s", cuiPlayerName(player, tc.PlayerIdx), cuiCardStr(tc.Card))
		}
		b.WriteString("\n")
	}

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&b, "%s\n", lastErr.Error())
	}

	// ゲーム状態
	if h.GetGameEndFlag() {
		winnerIdx := h.GetWinnerIdx()
		player := h.GetPlayer(winnerIdx)
		fmt.Fprintf(&b, "ゲーム終了！ %sの勝利です！\n", cuiPlayerName(player, winnerIdx))
	} else {
		phase := h.GetPhase()
		switch phase {
		case domain.HeartsPhasePass:
			dir := h.GetPassDirection()
			fmt.Fprintf(&b, "パスフェーズ: %s\n", cuiPassDirectionStr(dir))
			b.WriteString("pass <idx> <idx> <idx>・・・3枚のカードを選択\n")
		case domain.HeartsPhasePlay:
			currentIdx := h.GetCurrentPlayerIdx()
			player := h.GetPlayer(currentIdx)
			fmt.Fprintf(&b, "手番: %s\n", cuiPlayerName(player, currentIdx))
			b.WriteString("play <idx>・・・カードを出す\n")
		case domain.HeartsPhaseTrickEnd:
			b.WriteString("トリック終了\n")
			b.WriteString("next・・・次のトリックへ\n")
		case domain.HeartsPhaseRoundEnd:
			b.WriteString("ラウンド終了\n")
			b.WriteString("next・・・次のラウンドへ\n")
		}
	}

	b.WriteString("==========\n")
	return b.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (p *HeartsCuiPresenter) ActionLogOutput(h interfaces.HeartsGame) string {
	if !h.GetGameEndFlag() {
		return actionLogToText(nil)
	}
	return actionLogToText(h.GetActionLog())
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
