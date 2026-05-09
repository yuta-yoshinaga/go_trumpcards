package presenter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// CasinoWarCuiPresenter カジノウォーCUIプレゼンタークラス
type CasinoWarCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *CasinoWarCuiPresenter) Output(cw interfaces.CasinoWarGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinowar.chipsLine", "chips", strconv.Itoa(cw.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinowar.phaseLine", "phase", cp.phaseStr(cw.GetPhase())))
	if cw.GetAnte() > 0 {
		sb.WriteString(i18n.Tf("casinowar.anteLine", "ante", strconv.Itoa(cw.GetAnte())))
		if cw.GetWarBet() > 0 {
			sb.WriteString(i18n.Tf("casinowar.warBetInline", "warBet", strconv.Itoa(cw.GetWarBet())))
		}
		sb.WriteString("\n")
	}

	if pc, dc := cw.GetPlayerCard(), cw.GetDealerCard(); pc != nil || dc != nil {
		sb.WriteString("--- " + color.Bold(i18n.T("casinowar.initialHeader")) + " ---\n")
		if pc != nil {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinowar.playerLine", "card", cuiCardStr(pc)))
		}
		if dc != nil {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinowar.dealerLine", "card", cuiCardStr(dc)))
		}
	}

	if burn := cw.GetBurnCards(); len(burn) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("casinowar.burnHeader")) + " ---\n")
		parts := make([]string, len(burn))
		for i, c := range burn {
			parts[i] = cuiCardStr(c)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if pw, dw := cw.GetPlayerWarCard(), cw.GetDealerWarCard(); pw != nil || dw != nil {
		sb.WriteString("--- " + color.Bold(i18n.T("casinowar.warHeader")) + " ---\n")
		if pw != nil {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinowar.playerLine", "card", cuiCardStr(pw)))
		}
		if dw != nil {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinowar.dealerLine", "card", cuiCardStr(dw)))
		}
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	if cw.GetGameEndFlag() {
		switch cw.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("casinowar.playerWins")) + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red(i18n.T("casinowar.playerLoses")) + "\n")
		default:
			sb.WriteString(color.Yellow(i18n.T("casinowar.push")) + "\n")
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinowar.totalPayoutLine", "payout", strconv.Itoa(cw.GetTotalPayout())))
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
		return i18n.T("casinowar.phaseBet")
	case domain.CasinoWarPhaseInitialDealt:
		return i18n.T("casinowar.phaseInitialDealt")
	case domain.CasinoWarPhaseTieDecision:
		return i18n.T("casinowar.phaseTieDecision")
	case domain.CasinoWarPhaseWarDealt:
		return i18n.T("casinowar.phaseWarDealt")
	case domain.CasinoWarPhaseEnd:
		return i18n.T("casinowar.phaseEnd")
	default:
		return i18n.T("casinowar.phaseUnknown")
	}
}
