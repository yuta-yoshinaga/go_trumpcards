package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TrashCuiPresenter トラッシュCUIプレゼンタークラス
type TrashCuiPresenter struct{}

// Output ゲーム状態を文字列出力
func (p *TrashCuiPresenter) Output(t interfaces.TrashGame, lastErr error) string {
	return buildCuiOutput("Trash (トラッシュ)", func(b *strings.Builder) {
		// Opponent first, then player — mirrors the physical table view.
		for _, idx := range [...]int{domain.TrashCpuIdx, domain.TrashHumanIdx} {
			label := "あなた"
			if t.IsCpuPlayer(idx) {
				label = "CPU"
			}
			if idx == t.GetCurrent() && t.GetPhase() != domain.TrashPhaseGameOver {
				label += " (ターン中)"
			}
			fmt.Fprintf(b, "%s:\n  ", label)
			for j, s := range t.GetPlayerSlots(idx) {
				if s.FaceUp && s.Card != nil {
					fmt.Fprintf(b, "[%02d %s] ", j+1, cuiCardStr(s.Card))
				} else {
					fmt.Fprintf(b, "[%02d  ? ] ", j+1)
				}
				if (j+1)%5 == 0 && j+1 < domain.TrashSlotCnt {
					b.WriteString("\n  ")
				}
			}
			b.WriteString("\n")
		}

		b.WriteString("----------\n")
		fmt.Fprintf(b, "山札: %d   捨て札: %d", t.GetStockSize(), t.GetDiscardSize())
		if top := t.GetDiscardTop(); top != nil {
			fmt.Fprintf(b, " (top: %s)", cuiCardStr(top))
		}
		b.WriteString("\n")
		if pending := t.GetPending(); pending != nil {
			fmt.Fprintf(b, "ペンディング: %s\n", cuiCardStr(pending))
		}
		b.WriteString("----------\n")

		if lastErr != nil {
			fmt.Fprintf(b, "%s\n", color.Red(lastErr.Error()))
		}

		switch t.GetPhase() {
		case domain.TrashPhasePlayerTurn:
			fmt.Fprintf(b, "手数: %d\n", t.GetMoveCount())
		case domain.TrashPhaseAwaitWild:
			fmt.Fprintf(b, "%s 手数: %d\n", color.Green("ワイルドを配置してください (p <位置>)"), t.GetMoveCount())
		case domain.TrashPhaseGameOver:
			if t.GetWinner() == domain.TrashHumanIdx {
				fmt.Fprintf(b, "%s 手数: %d\n", color.Green("あなたの勝ち！"), t.GetMoveCount())
			} else {
				fmt.Fprintf(b, "%s 手数: %d\n", color.Red("CPUの勝ち"), t.GetMoveCount())
			}
		}
	})
}

// ActionLogOutput 棋譜をテキスト出力
func (p *TrashCuiPresenter) ActionLogOutput(t interfaces.TrashGame) string {
	if t.GetPhase() != domain.TrashPhaseGameOver {
		return actionLogToText(nil)
	}
	return actionLogToText(t.GetActionLog())
}
