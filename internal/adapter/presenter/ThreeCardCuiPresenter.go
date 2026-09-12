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

	// Web と同じく、ベット判断の前に配当倍率を表示する。倍率はドメインの
	// 定数から渡し、ロケールに数字を重複して持たせない。
	if tc.GetPhase() == domain.ThreeCardPhaseBet && !tc.GetGameEndFlag() {
		sb.WriteString(tp.payoutTable())
	}

	playerHand := tc.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("threecard.playerHeader")) + " ---\n")
		rank := tc.GetPlayerHandRank()
		if rank > 0 && rank < len(domain.ThreeCardHandNames) {
			sb.WriteString(i18n.Tf("threecard.handLine", "hand", threeCardHandName(rank)) + "\n")
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
			sb.WriteString(i18n.Tf("threecard.handLine", "hand", threeCardHandName(rank)) + "\n")
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
		sb.WriteString(i18n.MarkErrorLine(color.Red(lastErr.Error())) + "\n")
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
		// **合計だけでは、どちらの賭けが返ってきたのか読めない。** ディーラーが
		// クオリファイしなければアンテだけ配当が付き、プレイは押し戻される ──
		// Web はこの内訳を行ごとに出しているのに、CUI は合計とボーナスしか
		// 出していなかった。0 の行は既存のボーナス行と同じく省く。
		if ante := tc.GetAntePayout(); ante != 0 {
			sb.WriteString(i18n.Tf("threecard.antePayoutLine", "payout", strconv.Itoa(ante)) + "\n")
		}
		if play := tc.GetPlayPayout(); play != 0 {
			sb.WriteString(i18n.Tf("threecard.playPayoutLine", "payout", strconv.Itoa(play)) + "\n")
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

// payoutTable はベットフェーズの配当倍率一覧を返す。
func (tp *ThreeCardCuiPresenter) payoutTable() string {
	var sb strings.Builder
	sb.WriteString(color.Bold(i18n.T("threecard.payoutRef.title")) + "\n")
	anteBonusWidth := len(strconv.Itoa(domain.ThreeCardPairPlusStraightFlush))
	sb.WriteString("  " + i18n.T("threecard.payoutRef.anteBonusHeader") + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.anteBonusStraight", "payout", threeCardPayoutRate(domain.ThreeCardAnteBonusStraight, anteBonusWidth)) + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.anteBonusThreeOfAKind", "payout", threeCardPayoutRate(domain.ThreeCardAnteBonusThreeOfAKind, anteBonusWidth)) + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.anteBonusStraightFlush", "payout", threeCardPayoutRate(domain.ThreeCardAnteBonusStraightFlush, anteBonusWidth)) + "\n")
	sb.WriteString("  " + i18n.T("threecard.payoutRef.pairPlusHeader") + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.pairPlusPair", "payout", threeCardPayoutRate(domain.ThreeCardPairPlusPair, anteBonusWidth)) + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.pairPlusFlush", "payout", threeCardPayoutRate(domain.ThreeCardPairPlusFlush, anteBonusWidth)) + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.pairPlusStraight", "payout", threeCardPayoutRate(domain.ThreeCardPairPlusStraight, anteBonusWidth)) + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.pairPlusThreeOfAKind", "payout", threeCardPayoutRate(domain.ThreeCardPairPlusThreeOfAKind, anteBonusWidth)) + "\n")
	sb.WriteString("  " + i18n.Tf("threecard.payoutRef.pairPlusStraightFlush", "payout", threeCardPayoutRate(domain.ThreeCardPairPlusStraightFlush, anteBonusWidth)) + "\n")
	return sb.String()
}

func threeCardPayoutRate(rate, width int) string {
	payout := strconv.Itoa(rate)
	return strings.Repeat(" ", width-len(payout)) + payout
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

// threeCardHandKeys は役ランクと共通役名キーの対応。
var threeCardHandKeys = []string{
	"", // 0 は未使用
	"highCard",
	"pair",
	"flush",
	"straight",
	"threeOfAKind",
	"straightFlush",
}

// threeCardHandName は役名をロケールに応じて返す。
//
// **`domain.ThreeCardHandNames` は英語の表示名配列。**そのまま埋めていたので
// 日本語ロケールでも Straight / Flush と出ていた (#4694)。訳は 3 カード固有では
// ないので、ポーカー役の共通表 `pokerhand` を引く。訳が無ければ英語名に落とす。
func threeCardHandName(rank int) string {
	if rank > 0 && rank < len(threeCardHandKeys) {
		full := "pokerhand." + threeCardHandKeys[rank]
		if name := i18n.T(full); name != full {
			return name
		}
	}
	if rank > 0 && rank < len(domain.ThreeCardHandNames) {
		return domain.ThreeCardHandNames[rank]
	}
	return ""
}
