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

// CaribbeanDrawCuiPresenter カリビアン・ドロー・ポーカーCUIプレゼンタークラス
type CaribbeanDrawCuiPresenter struct {
}

// Output ゲーム状態を出力
func (cp *CaribbeanDrawCuiPresenter) Output(cs interfaces.CaribbeanDrawGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("caribbeandraw.chipsLine", "chips", strconv.Itoa(cs.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("caribbeandraw.phaseLine", "phase", cp.phaseStr(cs.GetPhase())) + "\n")

	// ジャックポットは任意の追加ベット。Web には常設の説明があるのに CUI には
	// 一言も無く、"b <ante> <jackpot>" で実チップを賭けさせていた (#5528)。
	// 賭け終わった後は出さない -- もう選べないものの説明は場所を取るだけ。
	if cs.GetPhase() == domain.CaribbeanDrawPhaseBet {
		sb.WriteString(i18n.T("caribbeandraw.jackpotHelp") + "\n")
	}
	// **交換できるのはこの一瞬だけ。** 手数料が要ることも併せて出さないと、
	// 引いてから残高が減っていることに気付くことになる。
	if cs.GetPhase() == domain.CaribbeanDrawPhaseDraw {
		sb.WriteString(i18n.Tf("caribbeandraw.drawHelp",
			"max", strconv.Itoa(domain.CaribbeanDrawMaxExchange),
			"cost", strconv.Itoa(cs.GetAnteBet()*domain.CaribbeanDrawExchangeCostRatio)) + "\n")
	}

	playerHand := cs.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("caribbeandraw.playerHeader")) + " ---\n")
		rank := cs.GetPlayerHandRank()
		if rank >= 0 && rank < len(domain.PokerHandNames) {
			sb.WriteString(i18n.Tf("caribbeandraw.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
		}
		parts := make([]string, len(playerHand))
		for i, card := range playerHand {
			parts[i] = cuiCardStr(card)
		}
		sb.WriteString(strings.Join(parts, ","))
		sb.WriteString("\n")
	}

	dealerHand := cs.GetDealerHand()
	if len(dealerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("caribbeandraw.dealerHeader")) + " ---\n")
		if cs.GetPhase() == domain.CaribbeanDrawPhaseEnd {
			rank := cs.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				sb.WriteString(i18n.Tf("caribbeandraw.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
			}
			if cs.GetDealerQualified() {
				sb.WriteString(i18n.T("caribbeandraw.qualified") + "\n")
			} else {
				sb.WriteString(i18n.T("caribbeandraw.notQualified") + "\n")
			}
			parts := make([]string, len(dealerHand))
			for i, card := range dealerHand {
				parts[i] = cuiCardStr(card)
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
		} else {
			// Action phase: show only the first dealer card; hide the rest behind "??".
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

	if cs.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("caribbeandraw.anteLine", "ante", strconv.Itoa(cs.GetAnteBet())) + "\n")
		if cs.GetPlayBet() > 0 {
			sb.WriteString(i18n.Tf("caribbeandraw.playLine", "play", strconv.Itoa(cs.GetPlayBet())) + "\n")
		}
		switch cs.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("caribbeandraw.playerWins")) + "\n")
		case domain.GameResultLose:
			if cs.GetPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("caribbeandraw.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("caribbeandraw.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("caribbeandraw.push")) + "\n")
		default:
		}
		// **交換手数料も見せる。** 配当ではないぶん配当行には現れず、引いた
		// ラウンドだけ残高の引き算が合わなく見えていた。
		if cost := cs.GetDrawCost(); cost > 0 {
			sb.WriteString(i18n.Tf("caribbeandraw.drawCostLine", "cost", strconv.Itoa(cost)) + "\n")
		}
		sb.WriteString(i18n.Tf("caribbeandraw.totalPayoutLine", "payout", strconv.Itoa(cs.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// caribbeanDrawAceValue / caribbeanDrawKingValue はエースとキングのカード値。
const (
	caribbeanDrawAceValue  = 1
	caribbeanDrawKingValue = 13
)

// HintOutput はアクションフェーズで Play / Fold を助言する。
//
// **判定はフロントの getCaribbeanDrawHint と同じ規則。**ワンペア以上なら Play、
// 役なしでも A と K を持っていれば Play、それ以外は Fold。ずれると同じ手札で
// CUI と Web が逆の助言を出す。
func (cp *CaribbeanDrawCuiPresenter) HintOutput(cs interfaces.CaribbeanDrawGame) string {
	if len(cs.GetPlayerHand()) == 0 {
		return i18n.T("caribbeandraw.hintNone") + "\n"
	}
	// **ドローも判断のひとつ。** クローン元にはこのフェーズが無いので助言も
	// 無かったが、ここで何を捨てるかは Play/Fold と同じくらい効く。
	if cs.GetPhase() == domain.CaribbeanDrawPhaseDraw {
		reason := "caribbeandraw.hintStandPat"
		if cs.GetPlayerHandRank() < domain.PokerHandOnePair {
			reason = "caribbeandraw.hintDrawWeak"
		}
		return color.Yellow(i18n.T(reason)) + "\n"
	}
	if cs.GetPhase() != domain.CaribbeanDrawPhaseAction {
		return i18n.T("caribbeandraw.hintNone") + "\n"
	}
	action, reason := i18n.T("caribbeandraw.hintFold"), "caribbeandraw.hintWeakHand"
	switch {
	case cs.GetPlayerHandRank() >= domain.PokerHandOnePair:
		action, reason = i18n.T("caribbeandraw.hintPlay"), "caribbeandraw.hintPairOrBetter"
	case caribbeanDrawHasAceKing(cs.GetPlayerHand()):
		action, reason = i18n.T("caribbeandraw.hintPlay"), "caribbeandraw.hintAceKingHigh"
	}
	return color.Yellow(i18n.Tf("caribbeandraw.hintDecision",
		"action", action, "reason", i18n.T(reason))) + "\n"
}

// caribbeanDrawHasAceKing は手札に A と K の両方があるかを返す。
func caribbeanDrawHasAceKing(cards []*domain.Card) bool {
	hasAce, hasKing := false, false
	for _, c := range cards {
		switch c.GetValue() {
		case caribbeanDrawAceValue:
			hasAce = true
		case caribbeanDrawKingValue:
			hasKing = true
		}
	}
	return hasAce && hasKing
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *CaribbeanDrawCuiPresenter) ActionLogOutput(cs interfaces.CaribbeanDrawGame) string {
	return actionLogOutputText(cs)
}

// phaseStr フェーズ文字列
func (cp *CaribbeanDrawCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.CaribbeanDrawPhaseBet:
		return i18n.T("caribbeandraw.phaseBet")
	case domain.CaribbeanDrawPhaseDraw:
		return i18n.T("caribbeandraw.phaseDraw")
	case domain.CaribbeanDrawPhaseAction:
		return i18n.T("caribbeandraw.phaseAction")
	case domain.CaribbeanDrawPhaseEnd:
		return i18n.T("caribbeandraw.phaseEnd")
	default:
		return i18n.T("caribbeandraw.phaseUnknown")
	}
}
