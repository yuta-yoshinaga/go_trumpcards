package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// SpeedCuiPresenter スピードCUIプレゼンタークラス
type SpeedCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *SpeedCuiPresenter) Output(s interfaces.SpeedGame, lastErr error) string {
	return buildCuiOutput("Speed (スピード)", func(b *strings.Builder) {
		// CPU情報
		cpu := s.GetPlayer(1)
		fmt.Fprintf(b, "CPU: 手札%d枚 / 山札%d枚\n", cpu.GetCardsSize(), cpu.GetDrawPileSize())

		// 台札
		b.WriteString("----------\n")
		b.WriteString(color.Bold("[台札]") + " ")
		for i := range 2 {
			c := s.GetCenterPile(i)
			if c != nil {
				fmt.Fprintf(b, "[%d] %s ", i, cuiCardStr(c))
			}
		}
		b.WriteString("\n")
		b.WriteString("----------\n")

		// 人間プレイヤー情報
		human := s.GetPlayer(0)
		fmt.Fprintf(b, "あなた: 手札%d枚 / 山札%d枚\n", human.GetCardsSize(), human.GetDrawPileSize())
		b.WriteString(cuiIndexedCardListStr(human))
		b.WriteString("\n")

		// ヒント
		ci, pi, found := s.GetHint()
		if found {
			fmt.Fprintf(b, "%s カード[%d]を台札[%d]に出せます\n", color.Bold("[ヒント]"), ci, pi)
		}

		// フェーズ状態
		switch s.GetPhase() {
		case 1: // Stuck
			b.WriteString(color.Yellow("膠着状態です。flip コマンドでカードをめくってください。\n"))
		}

		// 勝敗
		if s.GetGameEndFlag() {
			if s.GetWinnerIdx() == 0 {
				b.WriteString(color.Green("あなたの勝ちです！\n"))
			} else {
				b.WriteString(color.Red("CPUの勝ちです...\n"))
			}
		}

		// エラー
		if lastErr != nil {
			fmt.Fprintf(b, "%s %s\n", color.Red("[エラー]"), lastErr.Error())
		}
	})
}

// ActionLogOutput 棋譜を文字列出力
func (p *SpeedCuiPresenter) ActionLogOutput(s interfaces.SpeedGame) string {
	return actionLogOutputText(s)
}
