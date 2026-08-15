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

// RussianPokerCuiPresenter ロシアンポーカーCUIプレゼンタークラス
type RussianPokerCuiPresenter struct {
}

// Output ゲーム状態を出力
func (rp *RussianPokerCuiPresenter) Output(g interfaces.RussianPokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("russianpoker.chipsLine", "chips", strconv.Itoa(g.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("russianpoker.phaseLine", "phase", rp.phaseStr(g.GetPhase())) + "\n")

	playerHand := g.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("russianpoker.playerHeader")) + " ---\n")
		if g.GetPhase() == domain.RussianPokerPhaseEnd {
			rank := g.GetPlayerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				sb.WriteString(i18n.Tf("russianpoker.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
			}
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	dealerHand := g.GetDealerHand()
	if len(dealerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("russianpoker.dealerHeader")) + " ---\n")
		if g.GetPhase() == domain.RussianPokerPhaseEnd {
			rank := g.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				sb.WriteString(i18n.Tf("russianpoker.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
			}
			if g.GetDealerQualified() {
				sb.WriteString(i18n.T("russianpoker.qualified") + "\n")
			} else {
				sb.WriteString(i18n.T("russianpoker.notQualified") + "\n")
			}
			parts := make([]string, len(dealerHand))
			for i, card := range dealerHand {
				parts[i] = cuiCardStr(card)
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
		} else {
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
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
	}

	if g.GetExchangeCount() > 0 {
		sb.WriteString(i18n.Tf("russianpoker.exchangeLine",
			"count", strconv.Itoa(g.GetExchangeCount()),
			"fee", strconv.Itoa(g.GetExchangeFee())) + "\n")
	}
	if g.GetBought6th() {
		sb.WriteString(i18n.Tf("russianpoker.buy6thLine",
			"fee", strconv.Itoa(g.GetBuy6thFee())) + "\n")
	}

	if g.GetPhase() == domain.RussianPokerPhaseForceQualify {
		sb.WriteString(color.Yellow(i18n.T("russianpoker.forceQualifyGuide")) + "\n")
	}

	if g.GetPhase() == domain.RussianPokerPhaseSelect {
		sb.WriteString(color.Yellow(i18n.T("russianpoker.selectGuide")) + "\n")
	}

	if g.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("russianpoker.anteLine", "ante", strconv.Itoa(g.GetAnteBet())) + "\n")
		if g.GetPlayBet() > 0 {
			sb.WriteString(i18n.Tf("russianpoker.playLine", "play", strconv.Itoa(g.GetPlayBet())) + "\n")
		}
		if g.GetForceExchanged() {
			sb.WriteString(i18n.Tf("russianpoker.forceExchangeLine",
				"fee", strconv.Itoa(g.GetForceExchangeFee())) + "\n")
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("russianpoker.playerWins")) + "\n")
		case domain.GameResultLose:
			if g.GetPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("russianpoker.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("russianpoker.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("russianpoker.push")) + "\n")
		default:
		}
		// Per-bet payout breakdown (omitted when zero) so the player sees which
		// bet paid, matching the web payout-breakdown.
		if ante := g.GetAntePayout(); ante != 0 {
			sb.WriteString(i18n.Tf("russianpoker.antePayoutLine", "payout", strconv.Itoa(ante)) + "\n")
		}
		if play := g.GetPlayPayout(); play != 0 {
			sb.WriteString(i18n.Tf("russianpoker.playPayoutLine", "payout", strconv.Itoa(play)) + "\n")
		}
		sb.WriteString(i18n.Tf("russianpoker.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (rp *RussianPokerCuiPresenter) ActionLogOutput(g interfaces.RussianPokerGame) string {
	return actionLogOutputText(g)
}

// phaseStr フェーズ文字列
func (rp *RussianPokerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.RussianPokerPhaseBet:
		return i18n.T("russianpoker.phaseBet")
	case domain.RussianPokerPhaseAction:
		return i18n.T("russianpoker.phaseAction")
	case domain.RussianPokerPhaseSelect:
		return i18n.T("russianpoker.phaseSelect")
	case domain.RussianPokerPhasePostAction:
		return i18n.T("russianpoker.phasePostAction")
	case domain.RussianPokerPhaseForceQualify:
		return i18n.T("russianpoker.phaseForceQualify")
	case domain.RussianPokerPhaseEnd:
		return i18n.T("russianpoker.phaseEnd")
	default:
		return i18n.T("russianpoker.phaseUnknown")
	}
}
