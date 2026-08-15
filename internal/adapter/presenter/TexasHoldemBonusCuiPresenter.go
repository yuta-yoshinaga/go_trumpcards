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

// TexasHoldemBonusCuiPresenter テキサスホールデムボーナスポーカーCUIプレゼンタークラス
type TexasHoldemBonusCuiPresenter struct{}

// Output ゲーム状態を出力
func (tp *TexasHoldemBonusCuiPresenter) Output(g interfaces.TexasHoldemBonusGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.chipsLine", "chips", strconv.Itoa(g.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.phaseLine", "phase", tp.phaseStr(g.GetPhase())))

	// **アクション中はアンテ額も実コストも画面に無かった (#4698)。**Web は
	// Play/Raise ボタンに ante×倍率をラベル表示している。CUI 側はアンテを
	// 暗記して暗算するしかなかった。END の anteLine は結果表示なので別。
	if cost := g.GetNextBetCost(); cost > 0 {
		action := i18n.T("texasholdembonus.costRaise")
		if g.GetPhase() == domain.TexasHoldemBonusPhasePreFlop {
			action = i18n.T("texasholdembonus.costPlay")
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.betCostLine",
			"ante", strconv.Itoa(g.GetAnteBet()),
			"action", action,
			"cost", strconv.Itoa(cost)))
	}

	if community := g.GetCommunity(); len(community) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("texasholdembonus.boardHeader")) + " ---\n")
		parts := make([]string, len(community))
		for i, card := range community {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if playerHand := g.GetPlayerHand(); len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("texasholdembonus.playerHeader")) + " ---\n")
		if g.GetPhase() == domain.TexasHoldemBonusPhaseEnd {
			rank := g.GetPlayerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.handLine", "hand", domain.PokerHandNames[rank]))
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
		sb.WriteString("--- " + color.Bold(i18n.T("texasholdembonus.dealerHeader")) + " ---\n")
		if g.GetPhase() == domain.TexasHoldemBonusPhaseEnd {
			rank := g.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.handLine", "hand", domain.PokerHandNames[rank]))
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
		fmt.Fprintf(&sb, "%s\n", i18n.MarkErrorLine(color.Red(lastErr.Error())))
	}

	if g.GetGameEndFlag() {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.anteLine", "ante", strconv.Itoa(g.GetAnteBet())))
		if g.GetBonusBet() > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.bonusLine", "bonus", strconv.Itoa(g.GetBonusBet())))
		}
		if play := g.GetTotalPlayBet(); play > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.playBetsLine", "play", strconv.Itoa(play)))
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("texasholdembonus.playerWins")) + "\n")
		case domain.GameResultLose:
			if g.GetTotalPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("texasholdembonus.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("texasholdembonus.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("texasholdembonus.push")) + "\n")
		default:
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("texasholdembonus.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())))
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
		return i18n.T("texasholdembonus.phaseBet")
	case domain.TexasHoldemBonusPhasePreFlop:
		return i18n.T("texasholdembonus.phasePreFlop")
	case domain.TexasHoldemBonusPhaseFlop:
		return i18n.T("texasholdembonus.phaseFlop")
	case domain.TexasHoldemBonusPhaseTurn:
		return i18n.T("texasholdembonus.phaseTurn")
	case domain.TexasHoldemBonusPhaseEnd:
		return i18n.T("texasholdembonus.phaseEnd")
	default:
		return i18n.T("texasholdembonus.phaseUnknown")
	}
}
