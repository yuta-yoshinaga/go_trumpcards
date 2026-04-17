package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// RedDogCuiPresenter レッドドッグCUIプレゼンタークラス
type RedDogCuiPresenter struct{}

// Output ゲーム状態を出力
func (rp *RedDogCuiPresenter) Output(rd interfaces.RedDogGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", rd.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", rp.phaseStr(rd.GetPhase()))
	if rd.GetAnte() > 0 {
		fmt.Fprintf(&sb, "ante: %d", rd.GetAnte())
		if rd.GetRaise() > 0 {
			fmt.Fprintf(&sb, "  raise: %d", rd.GetRaise())
		}
		sb.WriteString("\n")
	}
	if rd.GetPhase() == domain.RedDogPhaseSpreadDecision || rd.GetPhase() == domain.RedDogPhaseEnd {
		fmt.Fprintf(&sb, "spread: %d\n", rd.GetSpread())
	}

	initial := rd.GetInitialCards()
	if len(initial) > 0 {
		sb.WriteString("--- " + color.Bold("INITIAL") + " ---\n")
		parts := make([]string, len(initial))
		for i, c := range initial {
			parts[i] = cuiCardStr(c)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if rd.GetThirdCard() != nil {
		sb.WriteString("--- " + color.Bold("THIRD") + " ---\n")
		sb.WriteString(cuiCardStr(rd.GetThirdCard()))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	if rd.GetGameEndFlag() {
		switch rd.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red("Player loses.") + "\n")
		default:
			sb.WriteString(color.Yellow("Push.") + "\n")
		}
		fmt.Fprintf(&sb, "total payout: %d\n", rd.GetTotalPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (rp *RedDogCuiPresenter) ActionLogOutput(rd interfaces.RedDogGame) string {
	return actionLogOutputText(rd)
}

// phaseStr フェーズ文字列
func (rp *RedDogCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.RedDogPhaseBet:
		return "BET"
	case domain.RedDogPhaseInitialDealt:
		return "INITIAL DEALT"
	case domain.RedDogPhaseSpreadDecision:
		return "SPREAD DECISION"
	case domain.RedDogPhasePairThird:
		return "PAIR THIRD"
	case domain.RedDogPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}
