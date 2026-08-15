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

// CaribbeanStudCuiPresenter カリビアンスタッドポーカーCUIプレゼンタークラス
type CaribbeanStudCuiPresenter struct {
}

// Output ゲーム状態を出力
func (cp *CaribbeanStudCuiPresenter) Output(cs interfaces.CaribbeanStudGame, lastErr error) string {
	var sb strings.Builder

	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("caribbeanstud.chipsLine", "chips", strconv.Itoa(cs.GetChips())) + "\n")
	sb.WriteString(i18n.Tf("caribbeanstud.phaseLine", "phase", cp.phaseStr(cs.GetPhase())) + "\n")

	playerHand := cs.GetPlayerHand()
	if len(playerHand) > 0 {
		sb.WriteString("--- " + color.Bold(i18n.T("caribbeanstud.playerHeader")) + " ---\n")
		rank := cs.GetPlayerHandRank()
		if rank >= 0 && rank < len(domain.PokerHandNames) {
			sb.WriteString(i18n.Tf("caribbeanstud.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
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
		sb.WriteString("--- " + color.Bold(i18n.T("caribbeanstud.dealerHeader")) + " ---\n")
		if cs.GetPhase() == domain.CaribbeanStudPhaseEnd {
			rank := cs.GetDealerHandRank()
			if rank >= 0 && rank < len(domain.PokerHandNames) {
				sb.WriteString(i18n.Tf("caribbeanstud.handLine", "hand", domain.PokerHandNames[rank]) + "\n")
			}
			if cs.GetDealerQualified() {
				sb.WriteString(i18n.T("caribbeanstud.qualified") + "\n")
			} else {
				sb.WriteString(i18n.T("caribbeanstud.notQualified") + "\n")
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
		sb.WriteString(i18n.Tf("caribbeanstud.anteLine", "ante", strconv.Itoa(cs.GetAnteBet())) + "\n")
		if cs.GetPlayBet() > 0 {
			sb.WriteString(i18n.Tf("caribbeanstud.playLine", "play", strconv.Itoa(cs.GetPlayBet())) + "\n")
		}
		switch cs.GetResult() {
		case domain.GameResultWin:
			sb.WriteString(color.Green(i18n.T("caribbeanstud.playerWins")) + "\n")
		case domain.GameResultLose:
			if cs.GetPlayBet() == 0 {
				sb.WriteString(color.Red(i18n.T("caribbeanstud.playerFolded")) + "\n")
			} else {
				sb.WriteString(color.Red(i18n.T("caribbeanstud.dealerWins")) + "\n")
			}
		case domain.GameResultDraw:
			sb.WriteString(color.Yellow(i18n.T("caribbeanstud.push")) + "\n")
		default:
		}
		sb.WriteString(i18n.Tf("caribbeanstud.totalPayoutLine", "payout", strconv.Itoa(cs.GetTotalPayout())) + "\n")
		sb.WriteString("----------\n")
	}

	return sb.String()
}

// caribbeanStudAceValue / caribbeanStudKingValue はエースとキングのカード値。
const (
	caribbeanStudAceValue  = 1
	caribbeanStudKingValue = 13
)

// HintOutput はアクションフェーズで Play / Fold を助言する。
//
// **判定はフロントの getCaribbeanStudHint と同じ規則。**ワンペア以上なら Play、
// 役なしでも A と K を持っていれば Play、それ以外は Fold。ずれると同じ手札で
// CUI と Web が逆の助言を出す。
func (cp *CaribbeanStudCuiPresenter) HintOutput(cs interfaces.CaribbeanStudGame) string {
	if cs.GetPhase() != domain.CaribbeanStudPhaseAction || len(cs.GetPlayerHand()) == 0 {
		return i18n.T("caribbeanstud.hintNone") + "\n"
	}
	action, reason := i18n.T("caribbeanstud.hintFold"), "caribbeanstud.hintWeakHand"
	switch {
	case cs.GetPlayerHandRank() >= domain.PokerHandOnePair:
		action, reason = i18n.T("caribbeanstud.hintPlay"), "caribbeanstud.hintPairOrBetter"
	case caribbeanStudHasAceKing(cs.GetPlayerHand()):
		action, reason = i18n.T("caribbeanstud.hintPlay"), "caribbeanstud.hintAceKingHigh"
	}
	return color.Yellow(i18n.Tf("caribbeanstud.hintDecision",
		"action", action, "reason", i18n.T(reason))) + "\n"
}

// caribbeanStudHasAceKing は手札に A と K の両方があるかを返す。
func caribbeanStudHasAceKing(cards []*domain.Card) bool {
	hasAce, hasKing := false, false
	for _, c := range cards {
		switch c.GetValue() {
		case caribbeanStudAceValue:
			hasAce = true
		case caribbeanStudKingValue:
			hasKing = true
		}
	}
	return hasAce && hasKing
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *CaribbeanStudCuiPresenter) ActionLogOutput(cs interfaces.CaribbeanStudGame) string {
	return actionLogOutputText(cs)
}

// phaseStr フェーズ文字列
func (cp *CaribbeanStudCuiPresenter) phaseStr(phase int) string {
	switch phase {
	case domain.CaribbeanStudPhaseBet:
		return i18n.T("caribbeanstud.phaseBet")
	case domain.CaribbeanStudPhaseAction:
		return i18n.T("caribbeanstud.phaseAction")
	case domain.CaribbeanStudPhaseEnd:
		return i18n.T("caribbeanstud.phaseEnd")
	default:
		return i18n.T("caribbeanstud.phaseUnknown")
	}
}
