//go:build !js || !wasm || casino

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

// UltimateTexasHoldemCuiPresenter アルティメット・テキサスホールデムCUIプレゼンタークラス
type UltimateTexasHoldemCuiPresenter struct{}

// Output ゲーム状態を出力
func (up *UltimateTexasHoldemCuiPresenter) Output(g interfaces.UltimateTexasHoldemGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.chipsLine", "chips", strconv.Itoa(g.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.phaseLine", "phase", up.phaseStr(g.GetPhase())))

	// During play (not the final result block), surface the current bets so the
	// player can size the play-bet multiple; omitted before the ante is placed.
	if !g.GetGameEndFlag() && g.GetAnteBet() > 0 {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.anteLine", "ante", strconv.Itoa(g.GetAnteBet())))
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.blindLine", "blind", strconv.Itoa(g.GetBlindBet())))
		if g.GetTripsBet() > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.tripsLine", "trips", strconv.Itoa(g.GetTripsBet())))
		}
		if play := g.GetPlayBet(); play > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.playBetLine", "play", strconv.Itoa(play)))
		}
	}

	if community := g.GetCommunity(); len(community) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("ultimatetexasholdem.boardHeader")) + " ---\n")
		parts := make([]string, len(community))
		for i, card := range community {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if playerHand := g.GetPlayerHand(); len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("ultimatetexasholdem.playerHeader")) + " ---\n")
		if g.GetPhase() == domain.UltimateTexasHoldemPhaseEnd {
			rank := g.GetPlayerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.handLine", "hand", cuiPokerHandName(rank)))
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
		sb.WriteString("--- " + color.Bold(i18n.T("ultimatetexasholdem.dealerHeader")) + " ---\n")
		if g.GetPhase() == domain.UltimateTexasHoldemPhaseEnd {
			rank := g.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.handLine", "hand", cuiPokerHandName(rank)))
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
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.anteLine", "ante", strconv.Itoa(g.GetAnteBet())))
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.blindLine", "blind", strconv.Itoa(g.GetBlindBet())))
		if g.GetTripsBet() > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.tripsLine", "trips", strconv.Itoa(g.GetTripsBet())))
		}
		if play := g.GetPlayBet(); play > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.playBetLine", "play", strconv.Itoa(play)))
		}
		if g.GetDealerQualified() {
			sb.WriteString(i18n.T("ultimatetexasholdem.dealerQualified") + "\n")
		} else {
			sb.WriteString(i18n.T("ultimatetexasholdem.dealerNotQualified") + "\n")
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("ultimatetexasholdem.playerWins")) + "\n")
		case domain.GameResultLose:
			if g.GetFolded() {
				sb.WriteString(color.Red(i18n.T("ultimatetexasholdem.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("ultimatetexasholdem.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("ultimatetexasholdem.push")) + "\n")
		default:
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("ultimatetexasholdem.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())))
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (up *UltimateTexasHoldemCuiPresenter) ActionLogOutput(g interfaces.UltimateTexasHoldemGame) string {
	return actionLogOutputText(g)
}

// phaseStr フェーズ文字列
func (up *UltimateTexasHoldemCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.UltimateTexasHoldemPhaseBet:
		return i18n.T("ultimatetexasholdem.phaseBet")
	case domain.UltimateTexasHoldemPhasePreFlop:
		return i18n.T("ultimatetexasholdem.phasePreFlop")
	case domain.UltimateTexasHoldemPhaseFlop:
		return i18n.T("ultimatetexasholdem.phaseFlop")
	case domain.UltimateTexasHoldemPhaseRiver:
		return i18n.T("ultimatetexasholdem.phaseRiver")
	case domain.UltimateTexasHoldemPhaseEnd:
		return i18n.T("ultimatetexasholdem.phaseEnd")
	default:
		return i18n.T("ultimatetexasholdem.phaseUnknown")
	}
}
