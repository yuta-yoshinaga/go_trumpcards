//go:build !js || !wasm || casino

package presenter

import (
	"sort"
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
	sb.WriteString(i18n.Tf("threecard.chipsLine", "chips", strconv.Itoa(tc.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("threecard.phaseLine", "phase", tp.phaseStr(tc.GetPhase())) + "\n")

	playerHand := tc.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("threecard.playerHeader")) + " ---\n")
		rank := tc.GetPlayerHandRank()
		if rank > 0 && rank < len(domain.ThreeCardHandNames) {
			sb.WriteString(i18n.Tf("threecard.handLine", "hand", domain.ThreeCardHandNames[rank]) + "\n")
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
			sb.WriteString(i18n.Tf("threecard.handLine", "hand", domain.ThreeCardHandNames[rank]) + "\n")
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
		sb.WriteString(color.Red(lastErr.Error()) + "\n")
	}

	if tc.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("threecard.anteLine", "ante", strconv.Itoa(tc.GetAnteBet())) + "\n")
		if tc.GetPlayBet() > 0 {
			sb.WriteString(i18n.Tf("threecard.playLine", "play", strconv.Itoa(tc.GetPlayBet())) + "\n")
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
		// Side-bet / bonus payout breakdown (omitted when zero to stay concise).
		if bonus := tc.GetAnteBonusPayout(); bonus != 0 {
			sb.WriteString(i18n.Tf("threecard.anteBonusPayoutLine", "payout", strconv.Itoa(bonus)) + "\n")
		}
		if pairPlus := tc.GetPairPlusPayout(); pairPlus != 0 {
			sb.WriteString(i18n.Tf("threecard.pairPlusPayoutLine", "payout", strconv.Itoa(pairPlus)) + "\n")
		}
		sb.WriteString(i18n.Tf("threecard.totalPayoutLine", "payout", strconv.Itoa(tc.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// threeCardShouldPlay reports whether the player's three-card hand meets the
// classic Q-6-4 "play" threshold. Any pair-or-better always plays; a high-card
// hand plays only when its Ace-high descending ranks are at least Q-6-4.
func threeCardShouldPlay(hand []*domain.Card, rank int) bool {
	if len(hand) != 3 {
		return false
	}
	if rank > domain.ThreeCardHandHighCard {
		return true
	}
	vals := make([]int, 3)
	for i, c := range hand {
		v := c.GetValue()
		if v == 1 {
			v = 14 // Ace is high in Three Card Poker.
		}
		vals[i] = v
	}
	sort.Sort(sort.Reverse(sort.IntSlice(vals)))
	for i, threshold := range []int{12, 6, 4} {
		if vals[i] != threshold {
			return vals[i] > threshold
		}
	}
	return true // exactly Q-6-4 is a play.
}

// HintOutput emits a play/fold recommendation (Q-6-4 strategy) during the
// action phase; other phases have no decision to advise.
func (tp *ThreeCardCuiPresenter) HintOutput(tc interfaces.ThreeCardGame) string {
	if tc.GetPhase() != domain.ThreeCardPhaseAction {
		return i18n.T("threecard.hintNone") + "\n"
	}
	if threeCardShouldPlay(tc.GetPlayerHand(), tc.GetPlayerHandRank()) {
		return color.Yellow(i18n.T("threecard.hintPlay")) + "\n"
	}
	return color.Yellow(i18n.T("threecard.hintFold")) + "\n"
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
