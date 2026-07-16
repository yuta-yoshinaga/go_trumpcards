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

// HighCardFlushCuiPresenter ハイカードフラッシュCUIプレゼンタークラス
type HighCardFlushCuiPresenter struct {
}

// Output ゲーム状態を出力
func (hp *HighCardFlushCuiPresenter) Output(hcf interfaces.HighCardFlushGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("highcardflush.chipsLine", "chips", strconv.Itoa(hcf.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("highcardflush.phaseLine", "phase", hp.phaseStr(hcf.GetPhase())) + "\n")

	playerHand := hcf.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("highcardflush.playerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("highcardflush.flushLine", "len", strconv.Itoa(hcf.GetPlayerFlushLen())) + "\n")
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	dealerHand := hcf.GetDealerHand()
	if len(dealerHand) > 0 && hcf.GetPhase() == domain.HighCardFlushPhaseEnd {
		sb.WriteString("--- " + color.Bold(i18n.T("highcardflush.dealerHeader")) + " ---\n")
		sb.WriteString(i18n.Tf("highcardflush.flushLine", "len", strconv.Itoa(hcf.GetDealerFlushLen())) + "\n")
		if hcf.GetDealerQualified() {
			sb.WriteString(i18n.T("highcardflush.qualified") + "\n")
		} else {
			sb.WriteString(i18n.T("highcardflush.notQualified") + "\n")
		}
		parts := make([]string, len(dealerHand))
		for i, card := range dealerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	sb.WriteString("----------\n")

	if lastErr != nil {
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if hcf.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("highcardflush.anteLine", "ante", strconv.Itoa(hcf.GetAnteBet())) + "\n")
		if hcf.GetRaiseBet() > 0 {
			sb.WriteString(i18n.Tf("highcardflush.raiseLine", "raise", strconv.Itoa(hcf.GetRaiseBet())) + "\n")
		}
		switch hcf.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("highcardflush.playerWins")) + "\n")
		case domain.GameResultLose:
			if hcf.GetRaiseBet() == 0 {
				sb.WriteString(color.Red(i18n.T("highcardflush.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("highcardflush.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("highcardflush.push")) + "\n")
		default:
		}
		sb.WriteString(i18n.Tf("highcardflush.totalPayoutLine", "payout", strconv.Itoa(hcf.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (hp *HighCardFlushCuiPresenter) ActionLogOutput(hcf interfaces.HighCardFlushGame) string {
	return actionLogOutputText(hcf)
}

// phaseStr フェーズ文字列
func (hp *HighCardFlushCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.HighCardFlushPhaseBet:
		return i18n.T("highcardflush.phaseBet")
	case domain.HighCardFlushPhaseAction:
		return i18n.T("highcardflush.phaseAction")
	case domain.HighCardFlushPhaseEnd:
		return i18n.T("highcardflush.phaseEnd")
	default:
		return i18n.T("highcardflush.phaseUnknown")
	}
}

// HintOutput emits a raise/fold recommendation during the action phase. Basic
// strategy: raise once the player's longest flush reaches the dealer's
// qualifying length (3); otherwise fold. Other phases have no decision to advise.
func (hp *HighCardFlushCuiPresenter) HintOutput(hcf interfaces.HighCardFlushGame) string {
	if hcf.GetPhase() != domain.HighCardFlushPhaseAction {
		return i18n.T("highcardflush.hintNone") + "\n"
	}
	if hcf.GetPlayerFlushLen() >= domain.HighCardFlushDealerMinFlushLen {
		return color.Yellow(i18n.T("highcardflush.hintRaise")) + "\n"
	}
	return color.Yellow(i18n.T("highcardflush.hintFold")) + "\n"
}
