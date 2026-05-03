package presenter

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
)

// CasinoWarCuiPresenter カジノウォーCUIプレゼンタークラス
type CasinoWarCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *CasinoWarCuiPresenter) Output(cw interfaces.CasinoWarGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "chips: %d\n", cw.GetChips())
	fmt.Fprintf(&sb, "phase: %s\n", cp.phaseStr(cw.GetPhase()))
	if cw.GetAnte() > 0 {
		fmt.Fprintf(&sb, "ante: %d", cw.GetAnte())
		if cw.GetWarBet() > 0 {
			fmt.Fprintf(&sb, "  warBet: %d", cw.GetWarBet())
		}
		sb.WriteString("\n")
	}

	if pc, dc := cw.GetPlayerCard(), cw.GetDealerCard(); pc != nil || dc != nil {
		sb.WriteString("--- " + color.Bold("INITIAL") + " ---\n")
		if pc != nil {
			fmt.Fprintf(&sb, "player: %s\n", cuiCardStr(pc))
		}
		if dc != nil {
			fmt.Fprintf(&sb, "dealer: %s\n", cuiCardStr(dc))
		}
	}

	if burn := cw.GetBurnCards(); len(burn) > 0 {
		sb.WriteString("--- " + color.Bold("BURN") + " ---\n")
		parts := make([]string, len(burn))
		for i, c := range burn {
			parts[i] = cuiCardStr(c)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if pw, dw := cw.GetPlayerWarCard(), cw.GetDealerWarCard(); pw != nil || dw != nil {
		sb.WriteString("--- " + color.Bold("WAR") + " ---\n")
		if pw != nil {
			fmt.Fprintf(&sb, "player: %s\n", cuiCardStr(pw))
		}
		if dw != nil {
			fmt.Fprintf(&sb, "dealer: %s\n", cuiCardStr(dw))
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	if cw.GetGameEndFlag() {
		switch cw.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green("Player wins!") + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red("Player loses.") + "\n")
		default:
			sb.WriteString(color.Yellow("Push.") + "\n")
		}
		fmt.Fprintf(&sb, "total payout: %d\n", cw.GetTotalPayout())
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *CasinoWarCuiPresenter) ActionLogOutput(cw interfaces.CasinoWarGame) string {
	return actionLogOutputText(cw)
}

// phaseStr フェーズ文字列
func (cp *CasinoWarCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.CasinoWarPhaseBet:
		return "BET"
	case domain.CasinoWarPhaseInitialDealt:
		return "INITIAL DEALT"
	case domain.CasinoWarPhaseTieDecision:
		return "TIE DECISION"
	case domain.CasinoWarPhaseWarDealt:
		return "WAR DEALT"
	case domain.CasinoWarPhaseEnd:
		return "END"
	default:
		return "UNKNOWN"
	}
}
