package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// ThreeCardCuiPresenter スリーカードポーカーCUIプレゼンタークラス
type ThreeCardCuiPresenter struct {
}

// Output ゲーム状態を出力
func (tp *ThreeCardCuiPresenter) Output(tc interfaces.ThreeCardGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", tc.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", tp.phaseStr(tc.GetPhase()))

	// プレイヤーハンド
	playerHand := tc.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold("PLAYER") + " ---\n")
		rank := tc.GetPlayerHandRank()
		if rank > 0 && rank < len(domain.ThreeCardHandNames) {
			fmt.Fprintf(&sb, "hand: %s\n", domain.ThreeCardHandNames[rank])
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	// ディーラーハンド（ENDフェーズのみ表示）
	dealerHand := tc.GetDealerHand()
	if len(dealerHand) > 0 && tc.GetPhase() == domain.ThreeCardPhaseEnd {
		sb.WriteString("--- " + color.Bold("DEALER") + " ---\n")
		rank := tc.GetDealerHandRank()
		if rank > 0 && rank < len(domain.ThreeCardHandNames) {
			fmt.Fprintf(&sb, "hand: %s\n", domain.ThreeCardHandNames[rank])
		}
		if tc.GetDealerQualified() {
			sb.WriteString("(Qualified)\n")
		} else {
			sb.WriteString("(Not Qualified)\n")
		}
		parts := make([]string, len(dealerHand))
		for i, card := range dealerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	// エラーメッセージ
	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	// ゲーム結果
	if tc.GetGameEndFlag() {
		fmt.Fprintf(&sb, "ante: %d\n", tc.GetAnteBet())
		if tc.GetPlayBet() > 0 {
			fmt.Fprintf(&sb, "play: %d\n", tc.GetPlayBet())
		}
		switch tc.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
		case domain.GameResultLose:
			if tc.GetPlayBet() == 0 {
				sb.WriteString(color.Red("Player folded.") + "\n")
			} else {
				sb.WriteString(color.Red("Dealer wins!") + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow("Push!") + "\n")
		default:
		}
		fmt.Fprintf(&sb, "total payout: %d\n", tc.GetTotalPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (tp *ThreeCardCuiPresenter) ActionLogOutput(tc interfaces.ThreeCardGame) string {
	return actionLogOutputText(tc)
}

// phaseStr フェーズ文字列
func (tp *ThreeCardCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.ThreeCardPhaseBet:
		return "BET"
	case domain.ThreeCardPhaseAction:
		return "ACTION"
	case domain.ThreeCardPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}
