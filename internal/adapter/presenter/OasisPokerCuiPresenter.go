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

// OasisPokerCuiPresenter オアシスポーカーCUIプレゼンタークラス
type OasisPokerCuiPresenter struct {
}

// Output ゲーム状態を出力
func (op *OasisPokerCuiPresenter) Output(g interfaces.OasisPokerGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("oasispoker.chipsLine", "chips", strconv.Itoa(g.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("oasispoker.phaseLine", "phase", op.phaseStr(g.GetPhase())) + "\n")

	playerHand := g.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("oasispoker.playerHeader")) + " ---\n")
		// プレイヤーハンドランクは End フェーズのみ表示（交換前は誤った印象を与えるため）
		if g.GetPhase() == domain.OasisPokerPhaseEnd {
			rank := g.GetPlayerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				sb.WriteString(i18n.Tf("oasispoker.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
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
		sb.WriteString("--- " + color.Bold(i18n.T("oasispoker.dealerHeader")) + " ---\n")
		if g.GetPhase() == domain.OasisPokerPhaseEnd {
			rank := g.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				sb.WriteString(i18n.Tf("oasispoker.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
			}
			if g.GetDealerQualified() {
				sb.WriteString(i18n.T("oasispoker.qualified") + "\n")
			} else {
				sb.WriteString(i18n.T("oasispoker.notQualified") + "\n")
			}
			// **アンティがプッシュになる理由が読めなかった** (#5595)。成立/不成立の
			// バッジは出ているのに、その条件はどこにも書かれていなかった。
			sb.WriteString(i18n.T("oasispoker.qualifyRule") + "\n")
			parts := make([]string, len(dealerHand))
			for i, card := range dealerHand {
				parts[i] = cuiCardStr(card)
			}
			sb.WriteString(strings.Join(parts, ","))
			sb.WriteString("\n")
		} else {
			// 交換/アクションフェーズ: ディーラーの1枚目のみ表示し、残りは "??" でマスクする。
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
		sb.WriteString(i18n.Tf("oasispoker.exchangeLine",
			"count", strconv.Itoa(g.GetExchangeCount()),
			"fee", strconv.Itoa(g.GetExchangeFee())) + "\n")
	}

	if g.GetGameEndFlag() {
		sb.WriteString(i18n.Tf("oasispoker.anteLine", "ante", strconv.Itoa(g.GetAnteBet())) + "\n")
		if g.GetPlayBet() > 0 {
			sb.WriteString(i18n.Tf("oasispoker.playLine", "play", strconv.Itoa(g.GetPlayBet())) + "\n")
		}
		switch g.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("oasispoker.playerWins")) + "\n")
		case domain.GameResultLose:
			if g.GetPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("oasispoker.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("oasispoker.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("oasispoker.push")) + "\n")
		default:
		}
		sb.WriteString(i18n.Tf("oasispoker.totalPayoutLine", "payout", strconv.Itoa(g.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// ActionLogOutput 棋譜をテキスト出力
func (op *OasisPokerCuiPresenter) ActionLogOutput(g interfaces.OasisPokerGame) string {
	return actionLogOutputText(g)
}

// phaseStr フェーズ文字列
func (op *OasisPokerCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.OasisPokerPhaseBet:
		return i18n.T("oasispoker.phaseBet")
	case domain.OasisPokerPhaseExchange:
		return i18n.T("oasispoker.phaseExchange")
	case domain.OasisPokerPhaseAction:
		return i18n.T("oasispoker.phaseAction")
	case domain.OasisPokerPhaseEnd:
		return i18n.T("oasispoker.phaseEnd")
	default:
		return i18n.T("oasispoker.phaseUnknown")
	}
}

// oasisPokerExchangeIndices returns the hand indices worth exchanging: cards
// that are neither part of a pair nor a high card (J/Q/K/A). Holding pairs and
// high cards is the standard draw heuristic.
func oasisPokerExchangeIndices(hand []*domain.Card) []int {
	rankCount := map[int]int{}
	for _, c := range hand {
		rankCount[c.GetValue()]++
	}
	var ex []int
	for i, c := range hand {
		v := c.GetValue()
		isPair := rankCount[v] >= 2
		isHigh := v == 1 || v >= 11 // Ace or J/Q/K
		if !isPair && !isHigh {
			ex = append(ex, i)
		}
	}
	return ex
}

// HintOutput advises the exchange (which cards to swap, or stand) and the action
// (play/fold via basic strategy) decisions. Other phases get no hint.
func (p *OasisPokerCuiPresenter) HintOutput(g interfaces.OasisPokerGame) string {
	switch g.GetPhase() {
	case domain.OasisPokerPhaseExchange:
		ex := oasisPokerExchangeIndices(g.GetPlayerHand())
		if len(ex) == 0 {
			return color.Yellow(i18n.T("oasispoker.hintStand")) + "\n"
		}
		hand := g.GetPlayerHand()
		parts := make([]string, len(ex))
		for i, idx := range ex {
			parts[i] = "[" + strconv.Itoa(idx) + "]" + cuiCardStr(hand[idx])
		}
		return color.Yellow(i18n.Tf("oasispoker.hintExchange", "cards", strings.Join(parts, " "))) + "\n"
	case domain.OasisPokerPhaseAction:
		if g.RecommendPlay() {
			return color.Yellow(i18n.T("oasispoker.hintPlay")) + "\n"
		}
		return color.Yellow(i18n.T("oasispoker.hintFold")) + "\n"
	default:
		return i18n.T("oasispoker.hintNone") + "\n"
	}
}
