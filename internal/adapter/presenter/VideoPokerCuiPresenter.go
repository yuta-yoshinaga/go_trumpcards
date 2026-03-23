package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// VideoPokerCuiPresenter ビデオポーカーCUIプレゼンタークラス
type VideoPokerCuiPresenter struct {
}

// Output ゲーム状態を出力
func (vpp *VideoPokerCuiPresenter) Output(vp interfaces.VideoPokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", vp.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", vpp.phaseStr(vp.GetPhase()))

	// ハンド表示
	hand := vp.GetHand()
	if len(hand) > 0 {
		sb.WriteString("--- HAND ---\n")
		held := vp.GetHeldIndices()
		parts := make([]string, len(hand))
		for i, card := range hand {
			s := cuiCardStr(card)
			if held[i] {
				s += " [HOLD]"
			}
			parts[i] = s
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	// ゲーム結果
	if vp.GetGameEndFlag() {
		fmt.Fprintf(&sb, "bet: %d coin(s)\n", vp.GetBetAmount())
		if vp.GetResult() == domain.GameResultWin {
			fmt.Fprintf(&sb, "%s\n", color.Green(vp.GetHandName()+"! You win!"))
		} else {
			sb.WriteString(color.Red("No winning hand.") + "\n")
		}
		fmt.Fprintf(&sb, "payout: %d\n", vp.GetPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (vpp *VideoPokerCuiPresenter) ActionLogOutput(vp interfaces.VideoPokerGame) string {
	return actionLogOutputText(vp)
}

// phaseStr フェーズ文字列
func (vpp *VideoPokerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.VideoPokerPhaseBet:
		return "BET"
	case domain.VideoPokerPhaseDraw:
		return "DRAW"
	case domain.VideoPokerPhaseResult:
		return "RESULT"
	default:
		return "UNKNOWN"
	}
}
