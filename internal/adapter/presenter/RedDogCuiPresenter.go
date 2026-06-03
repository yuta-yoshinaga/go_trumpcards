//go:build !js || !wasm || casino

package presenter

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/color"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
)

// RedDogCuiPresenter レッドドッグCUIプレゼンタークラス
type RedDogCuiPresenter struct{}

// Output ゲーム状態を出力
func (rp *RedDogCuiPresenter) Output(rd interfaces.RedDogGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("reddog.chipsLine", "chips", strconv.Itoa(rd.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("reddog.phaseLine", "phase", rp.phaseStr(rd.GetPhase())) + "\n")
	if rd.GetAnte() > 0 {
		sb.WriteString(i18n.Tf("reddog.anteLine", "ante", strconv.Itoa(rd.GetAnte())))
		if rd.GetRaise() > 0 {
			sb.WriteString(i18n.Tf("reddog.raiseInline", "raise", strconv.Itoa(rd.GetRaise())))
		}
		sb.WriteString("\n")
	}
	if rd.GetPhase() == domain.RedDogPhaseSpreadDecision || rd.GetPhase() == domain.RedDogPhaseEnd {
		sb.WriteString(i18n.Tf("reddog.spreadLine", "spread", strconv.Itoa(rd.GetSpread())) + "\n")
	}

	initial := rd.GetInitialCards()
	if len(initial) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("reddog.initialHeader")) + " ---\n")
		parts := make([]string, len(initial))
		for i, c := range initial {
			parts[i] = cuiCardStr(c)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if rd.GetThirdCard() != nil {
		sb.WriteString("--- " + color.Bold(i18n.T("reddog.thirdHeader")) + " ---\n")
		sb.WriteString(cuiCardStr(rd.GetThirdCard()))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if rd.GetGameEndFlag() {
		switch rd.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("reddog.playerWins")) + "\n")
		case domain.GameResultLose:
			sb.WriteString(color.Red(i18n.T("reddog.playerLoses")) + "\n")
		default:
			sb.WriteString(color.Yellow(i18n.T("reddog.push")) + "\n")
		}
		sb.WriteString(i18n.Tf("reddog.totalPayoutLine", "payout", strconv.Itoa(rd.GetTotalPayout())) + "\n")
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
		return i18n.T("reddog.phaseBet")
	case domain.RedDogPhaseInitialDealt:
		return i18n.T("reddog.phaseInitialDealt")
	case domain.RedDogPhaseSpreadDecision:
		return i18n.T("reddog.phaseSpreadDecision")
	case domain.RedDogPhasePairThird:
		return i18n.T("reddog.phasePairThird")
	case domain.RedDogPhaseEnd:
		return i18n.T("reddog.phaseEnd")
	default:
		return i18n.T("reddog.phaseUnknown")
	}
}
