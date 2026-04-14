package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// LetItRideCuiPresenter レット・イット・ライドCUIプレゼンタークラス
type LetItRideCuiPresenter struct {
}

// Output ゲーム状態を出力
func (lp *LetItRideCuiPresenter) Output(lir interfaces.LetItRideGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", lir.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", lp.phaseStr(lir.GetPhase()))

	if lir.GetBetAmount() > 0 {
		fmt.Fprintf(&sb, "bet: %d x %d active\n", lir.GetBetAmount(), lp.activeBetCount(lir))
	}

	playerHand := lir.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold("PLAYER") + " ---\n")
		if lir.GetPhase() == domain.LetItRidePhaseEnd {
			rank := lir.GetHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "hand: %s\n", domain.PokerHandNames[rank])
			}
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	communityCards := lir.GetCommunityCards()
	if len(communityCards) > 0 {
		sb.WriteString("--- " + color.Bold("COMMUNITY") + " ---\n")
		parts := make([]string, len(communityCards))
		switch lir.GetPhase() {
		case domain.LetItRidePhaseBet, domain.LetItRidePhaseFirstDecision:
			for i := range communityCards {
				parts[i] = "??"
			}
		case domain.LetItRidePhaseSecondDecision:
			parts[0] = cuiCardStr(communityCards[0])
			for i := 1; i < len(communityCards); i++ {
				parts[i] = "??"
			}
		default:
			for i, card := range communityCards {
				parts[i] = cuiCardStr(card)
			}
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	if lir.GetGameEndFlag() {
		fmt.Fprintf(&sb, "bet1: %s  bet2: %s  bet3: %s\n",
			lp.betStatusStr(lir.GetBet1Active()),
			lp.betStatusStr(lir.GetBet2Active()),
			lp.betStatusStr(lir.GetBet3Active()))
		switch lir.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red("Player loses.") + "\n")
		default:
		}
		fmt.Fprintf(&sb, "total payout: %d\n", lir.GetTotalPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (lp *LetItRideCuiPresenter) ActionLogOutput(lir interfaces.LetItRideGame) string {
	return actionLogOutputText(lir)
}

// phaseStr フェーズ文字列
func (lp *LetItRideCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.LetItRidePhaseBet:
		return "BET"
	case domain.LetItRidePhaseFirstDecision:
		return "FIRST DECISION"
	case domain.LetItRidePhaseSecondDecision:
		return "SECOND DECISION"
	case domain.LetItRidePhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}

// betStatusStr ベット状態文字列
func (lp *LetItRideCuiPresenter) betStatusStr(active bool) string {
	if active {
		return color.Green("RIDE")
	}
	return color.Yellow("PULL")
}

// activeBetCount アクティブベット数
func (lp *LetItRideCuiPresenter) activeBetCount(lir interfaces.LetItRideGame) int {
	count := 0
	if lir.GetBet1Active() {
		count++
	}
	if lir.GetBet2Active() {
		count++
	}
	if lir.GetBet3Active() {
		count++
	}
	return count
}
