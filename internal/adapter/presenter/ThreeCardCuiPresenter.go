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

// ThreeCardCuiPresenter スリーカードポーカーCUIプレゼンタークラス
type ThreeCardCuiPresenter struct {
}

// Output ゲーム状態を出力
func (tp *ThreeCardCuiPresenter) Output(tc interfaces.ThreeCardGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("threecard.chipsLine", "chips", strconv.Itoa(tc.GetChips())))
	fmt.Fprintf(&sb, "%s\n", i18n.Tf("threecard.phaseLine", "phase", tp.phaseStr(tc.GetPhase())))

	playerHand := tc.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("threecard.playerHeader")) + " ---\n")
		rank := tc.GetPlayerHandRank()
		if rank > 0 && rank < len(domain.ThreeCardHandNames) {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("threecard.handLine", "hand", domain.ThreeCardHandNames[rank]))
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	dealerHand := tc.GetDealerHand()
	if len(dealerHand) > 0 && tc.GetPhase() == domain.ThreeCardPhaseEnd {
		sb.WriteString("--- " + color.Bold(i18n.T("threecard.dealerHeader")) + " ---\n")
		rank := tc.GetDealerHandRank()
		if rank > 0 && rank < len(domain.ThreeCardHandNames) {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("threecard.handLine", "hand", domain.ThreeCardHandNames[rank]))
		}
		if tc.GetDealerQualified() {
			sb.WriteString(i18n.T("threecard.qualified") + "\n")
		} else {
			sb.WriteString(i18n.T("threecard.notQualified") + "\n")
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
		fmt.Fprintf(&sb, "%s\n", color.Red(lastErr.Error()))
	}

	if tc.GetGameEndFlag() {
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("threecard.anteLine", "ante", strconv.Itoa(tc.GetAnteBet())))
		if tc.GetPlayBet() > 0 {
			fmt.Fprintf(&sb, "%s\n", i18n.Tf("threecard.playLine", "play", strconv.Itoa(tc.GetPlayBet())))
		}
		switch tc.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("threecard.playerWins")) + "\n")
		case domain.GameResultLose:
			if tc.GetPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("threecard.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("threecard.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("threecard.push")) + "\n")
		default:
		}
		fmt.Fprintf(&sb, "%s\n", i18n.Tf("threecard.totalPayoutLine", "payout", strconv.Itoa(tc.GetTotalPayout())))
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (tp *ThreeCardCuiPresenter) ActionLogOutput(tc interfaces.ThreeCardGame) string {
	return actionLogOutputText(tc)
}

// phaseStr フェーズ文字列
func (tp *ThreeCardCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.ThreeCardPhaseBet:
		return i18n.T("threecard.phaseBet")
	case domain.ThreeCardPhaseAction:
		return i18n.T("threecard.phaseAction")
	case domain.ThreeCardPhaseEnd:
		return i18n.T("threecard.phaseEnd")
	default:
		return i18n.T("threecard.phaseUnknown")
	}
}
