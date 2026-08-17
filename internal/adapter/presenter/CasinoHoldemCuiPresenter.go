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

// CasinoHoldemCuiPresenter カジノホールデムCUIプレゼンタークラス
type CasinoHoldemCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *CasinoHoldemCuiPresenter) Output(g interfaces.CasinoHoldemGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinoholdem.chipsLine", "chips", strconv.Itoa(g.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinoholdem.phaseLine", "phase", cp.phaseStr(g.GetPhase())))

	if community := g.GetCommunity(); len(community) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("casinoholdem.boardHeader")) + " ---\n")
		parts := make([]string, len(community))
		for i, card := range community {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	if playerHand := g.GetPlayerHand(); len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("casinoholdem.playerHeader")) + " ---\n")
		// **フロップでも現在の役を出す。**Call/Fold を決めるのはこの時点なので、
		// 役が END にしか出ないと CUI だけ判断材料を欠く (#5604)。ドメインは
		// dealFlop で playerHandRank をホール+フロップの 5 枚から更新済み。
		phase := g.GetPhase()
		if phase == domain.CasinoHoldemPhaseEnd || phase == domain.CasinoHoldemPhaseFlop {
			rank := g.GetPlayerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				key := "casinoholdem.handLine"
				if phase == domain.CasinoHoldemPhaseFlop {
					key = "casinoholdem.currentHandLine"
				}
				fmt.Fprintf(&sb, "%s\n", i18n.Tf(key, "hand", domain.PokerHandNames[rank]))
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
		sb.WriteString("--- " + color.Bold(i18n.T("casinoholdem.dealerHeader")) + " ---\n")
		if g.GetPhase() == domain.CasinoHoldemPhaseEnd {
			rank := g.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinoholdem.handLine", "hand", domain.PokerHandNames[rank]))
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
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinoholdem.anteLine", "ante", strconv.Itoa(g.GetAnteBet())))
		if g.GetBonusBet() > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinoholdem.bonusLine", "bonus", strconv.Itoa(g.GetBonusBet())))
		}
		if call := g.GetCallBet(); call > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinoholdem.callBetLine", "call", strconv.Itoa(call)))
		}
		// ディーラークオリファイ表示（フォールド時を除く＝ショーダウン後のみ意味がある）
		if g.GetCallBet() > 0 {
			if g.GetDealerQualify() {
				sb.WriteString(i18n.T("casinoholdem.dealerQualified") + "\n")
			} else {
				sb.WriteString(i18n.T("casinoholdem.dealerDoesNotQualify") + "\n")
			}
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("casinoholdem.playerWins")) + "\n")
		case domain.GameResultLose:
			if g.GetCallBet() == 0 {
				sb.WriteString(color.Red(i18n.T("casinoholdem.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("casinoholdem.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("casinoholdem.push")) + "\n")
		default:
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("casinoholdem.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())))
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *CasinoHoldemCuiPresenter) ActionLogOutput(g interfaces.CasinoHoldemGame) string {
	return actionLogOutputText(g)
}

// phaseStr フェーズ文字列
func (cp *CasinoHoldemCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.CasinoHoldemPhaseBet:
		return i18n.T("casinoholdem.phaseBet")
	case domain.CasinoHoldemPhaseFlop:
		return i18n.T("casinoholdem.phaseFlop")
	case domain.CasinoHoldemPhaseEnd:
		return i18n.T("casinoholdem.phaseEnd")
	default:
		return i18n.T("casinoholdem.phaseUnknown")
	}
}

// HintOutput advises call/fold after the flop using basic strategy (call with a
// pair or better, or an Ace/King; otherwise fold). Other phases get no hint.
func (p *CasinoHoldemCuiPresenter) HintOutput(g interfaces.CasinoHoldemGame) string {
	if g.GetPhase() != domain.CasinoHoldemPhaseFlop {
		return i18n.T("casinoholdem.hintNone") + "\n"
	}
	if g.RecommendCall() {
		return color.Yellow(i18n.T("casinoholdem.hintCall")) + "\n"
	}
	return color.Yellow(i18n.T("casinoholdem.hintFold")) + "\n"
}
