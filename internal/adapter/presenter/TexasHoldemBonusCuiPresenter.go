package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// TexasHoldemBonusCuiPresenter テキサスホールデムボーナスポーカーCUIプレゼンタークラス
type TexasHoldemBonusCuiPresenter struct{}

// Output ゲーム状態を出力
func (tp *TexasHoldemBonusCuiPresenter) Output(g interfaces.TexasHoldemBonusGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", g.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", tp.phaseStr(g.GetPhase()))

	if community := g.GetCommunity(); len(community) > 0 {
		sb.WriteString("--- " + color.Bold("BOARD") + " ---\n")
		parts := make([]string, len(community))
		for i, card := range community {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if playerHand := g.GetPlayerHand(); len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold("PLAYER") + " ---\n")
		if g.GetPhase() == domain.TexasHoldemBonusPhaseEnd {
			rank := g.GetPlayerHandRank()
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

	if dealerHand := g.GetDealerHand(); len(dealerHand) > 0 {
		sb.WriteString("--- " + color.Bold("DEALER") + " ---\n")
		if g.GetPhase() == domain.TexasHoldemBonusPhaseEnd {
			rank := g.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "hand: %s\n", domain.PokerHandNames[rank])
			}
			parts := make([]string, len(dealerHand))
			for i, card := range dealerHand {
				parts[i] = cuiCardStr(card)
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
		} else {
			parts := make([]string, len(dealerHand))
			for i := range dealerHand {
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

	if g.GetGameEndFlag() {
		fmt.Fprintf(&sb, "ante: %d\n", g.GetAnteBet())
		if g.GetBonusBet() > 0 {
			fmt.Fprintf(&sb, "bonus: %d\n", g.GetBonusBet())
		}
		if play := g.GetTotalPlayBet(); play > 0 {
			fmt.Fprintf(&sb, "play bets: %d\n", play)
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
		case domain.GameResultLose:
			if g.GetTotalPlayBet() == 0 {
				sb.WriteString(color.Red("Player folded.") + "\n")
			} else {
				sb.WriteString(color.Red("Dealer wins!") + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow("Push!") + "\n")
		default:
		}
		fmt.Fprintf(&sb, "total payout: %d\n", g.GetTotalPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (tp *TexasHoldemBonusCuiPresenter) ActionLogOutput(g interfaces.TexasHoldemBonusGame) string {
	return actionLogOutputText(g)
}

// phaseStr フェーズ文字列
func (tp *TexasHoldemBonusCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.TexasHoldemBonusPhaseBet:
		return "BET"
	case domain.TexasHoldemBonusPhasePreFlop:
		return "PRE-FLOP"
	case domain.TexasHoldemBonusPhaseFlop:
		return "FLOP"
	case domain.TexasHoldemBonusPhaseTurn:
		return "TURN"
	case domain.TexasHoldemBonusPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}
