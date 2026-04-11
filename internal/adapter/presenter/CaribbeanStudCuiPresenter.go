package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CaribbeanStudCuiPresenter カリビアンスタッドポーカーCUIプレゼンタークラス
type CaribbeanStudCuiPresenter struct {
}

// Output ゲーム状態を出力
func (cp *CaribbeanStudCuiPresenter) Output(cs interfaces.CaribbeanStudGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", cs.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", cp.phaseStr(cs.GetPhase()))

	playerHand := cs.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold("PLAYER") + " ---\n")
		rank := cs.GetPlayerHandRank()
		if rank >= 0 && rank < len(domain.PokerHandNames) {
			fmt.Fprintf(&sb, "hand: %s\n", domain.PokerHandNames[rank])
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	dealerHand := cs.GetDealerHand()
	if len(dealerHand) > 0 {
		sb.WriteString("--- " + color.Bold("DEALER") + " ---\n")
		if cs.GetPhase() == domain.CaribbeanStudPhaseEnd {
			rank := cs.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "hand: %s\n", domain.PokerHandNames[rank])
			}
			if cs.GetDealerQualified() {
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
		} else {
			// Action phase: show only the first card; hide the rest
			parts := make([]string, len(dealerHand))
			parts[0] = cuiCardStr(dealerHand[0])
			for i := 1; i < len(dealerHand); i++ {
				parts[i] = "??"
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	if cs.GetGameEndFlag() {
		fmt.Fprintf(&sb, "ante: %d\n", cs.GetAnteBet())
		if cs.GetPlayBet() > 0 {
			fmt.Fprintf(&sb, "play: %d\n", cs.GetPlayBet())
		}
		switch cs.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
		case domain.GameResultLose:
			if cs.GetPlayBet() == 0 {
				sb.WriteString(color.Red("Player folded.") + "\n")
			} else {
				sb.WriteString(color.Red("Dealer wins!") + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow("Push!") + "\n")
		default:
		}
		fmt.Fprintf(&sb, "total payout: %d\n", cs.GetTotalPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *CaribbeanStudCuiPresenter) ActionLogOutput(cs interfaces.CaribbeanStudGame) string {
	return actionLogOutputText(cs)
}

// phaseStr フェーズ文字列
func (cp *CaribbeanStudCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.CaribbeanStudPhaseBet:
		return "BET"
	case domain.CaribbeanStudPhaseAction:
		return "ACTION"
	case domain.CaribbeanStudPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}
