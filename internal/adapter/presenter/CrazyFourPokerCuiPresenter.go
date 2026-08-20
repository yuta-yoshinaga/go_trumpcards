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

// crazyFourPokerQueensUpTableStr は配当表を "Four of a Kind 50:1 / ..." で返す。
//
// **表を写さない** (#5775)。ドメインの CrazyFourPokerQueensUpPayout が唯一の出所。
func crazyFourPokerQueensUpTableStr() string {
	rows := domain.CrazyFourPokerQueensUpPayout()
	parts := make([]string, 0, len(rows))
	for _, r := range rows {
		name := ""
		if r.Hand >= 0 && r.Hand < len(domain.FourCardHandNames) {
			name = domain.FourCardHandNames[r.Hand]
		}
		parts = append(parts, fmt.Sprintf("%s %d:1", name, r.Multiplier))
	}
	return strings.Join(parts, " / ")
}

// CrazyFourPokerCuiPresenter クレイジー 4 ポーカーCUIプレゼンタークラス
type CrazyFourPokerCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *CrazyFourPokerCuiPresenter) Output(c interfaces.CrazyFourPokerGame, lastErr error) string {
	return buildCuiOutput(i18n.T("crazyfourpoker.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("crazyfourpoker.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("crazyfourpoker.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"chips", strconv.Itoa(c.GetChips())) + "\n")

		// **賭ける前に見えなければ意味がない** (#5775)。何が当たれば何倍かを
		// 知って額を決めるものなので、置く前の局面で出す。
		if c.GetPhase() == domain.CrazyFourPokerPhaseBet {
			sb.WriteString(i18n.Tf("crazyfourpoker.queensUpPayoutLine",
				"table", crazyFourPokerQueensUpTableStr()) + "\n")
		}
		if c.GetAnteBet() > 0 {
			sb.WriteString(i18n.Tf("crazyfourpoker.betLine",
				"ante", strconv.Itoa(c.GetAnteBet()),
				"super", strconv.Itoa(c.GetSuperBet()),
				"queensUp", strconv.Itoa(c.GetQueensUpBet())) + "\n")
		}
		cp.writeHands(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeHands は手札と役を書き出す。
func (cp *CrazyFourPokerCuiPresenter) writeHands(sb *strings.Builder, c interfaces.CrazyFourPokerGame) {
	if len(c.GetPlayerHand()) == 0 {
		return
	}
	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("crazyfourpoker.playerHandLine",
		"cards", crazyFourPokerCardsStr(c.GetPlayerHand())) + "\n")
	sb.WriteString(i18n.Tf("crazyfourpoker.playerBestLine",
		"cards", crazyFourPokerCardsStr(c.GetPlayerBest()),
		"rank", crazyFourPokerRankName(c.GetPlayerHandRank())) + "\n")

	// **エース以上なら倍率を動かせる。** 選べることを画面に出す。
	if c.GetPhase() == domain.CrazyFourPokerPhaseDecide {
		sb.WriteString(i18n.Tf("crazyfourpoker.maxMultiplierLine",
			"max", strconv.Itoa(c.MaxPlayMultiplier())) + "\n")
	}

	// **決着するまでディーラーの手は伏せる。**
	if c.GetPhase() != domain.CrazyFourPokerPhaseResult {
		return
	}
	sb.WriteString(i18n.Tf("crazyfourpoker.dealerHandLine",
		"cards", crazyFourPokerCardsStr(c.GetDealerHand())) + "\n")
	sb.WriteString(i18n.Tf("crazyfourpoker.dealerBestLine",
		"cards", crazyFourPokerCardsStr(c.GetDealerBest()),
		"rank", crazyFourPokerRankName(c.GetDealerHandRank())) + "\n")
	if !c.DealerQualifies() {
		sb.WriteString(i18n.T("crazyfourpoker.notQualifiedLine") + "\n")
	}
}

// writeResult は決着と収支を書き出す。
func (cp *CrazyFourPokerCuiPresenter) writeResult(sb *strings.Builder, c interfaces.CrazyFourPokerGame) {
	if c.GetPhase() != domain.CrazyFourPokerPhaseResult {
		return
	}
	staked := c.GetAnteBet() + c.GetSuperBet() + c.GetQueensUpBet() + c.GetPlayBet()
	net := c.GetPayout() - staked
	msg := i18n.Tf("crazyfourpoker.resultLine",
		"result", crazyFourPokerResultName(c.GetResult()),
		"net", strconv.Itoa(net))
	if net >= 0 {
		sb.WriteString(color.Green(msg) + "\n")
	} else {
		sb.WriteString(color.Red(msg) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.T("crazyfourpoker.brokeLine") + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *CrazyFourPokerCuiPresenter) ActionLogOutput(c interfaces.CrazyFourPokerGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *CrazyFourPokerCuiPresenter) HintOutput(c interfaces.CrazyFourPokerGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("crazyfourpoker.hintNone")
	}
	if h.Multiplier == 0 {
		return i18n.Tf("crazyfourpoker.hintFold", "reason", i18n.T("crazyfourpoker."+h.Reason))
	}
	return i18n.Tf("crazyfourpoker.hintPlay",
		"multiplier", strconv.Itoa(h.Multiplier),
		"reason", i18n.T("crazyfourpoker."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *CrazyFourPokerCuiPresenter) phaseStr(phase domain.CrazyFourPokerPhase) string {
	switch phase {
	case domain.CrazyFourPokerPhaseBet:
		return i18n.T("crazyfourpoker.phaseBet")
	case domain.CrazyFourPokerPhaseDecide:
		return i18n.T("crazyfourpoker.phaseDecide")
	case domain.CrazyFourPokerPhaseResult:
		return i18n.T("crazyfourpoker.phaseResult")
	default:
		return i18n.T("crazyfourpoker.phaseUnknown")
	}
}

// crazyFourPokerResultName は決着の表示名を返す。
func crazyFourPokerResultName(r domain.CrazyFourPokerResult) string {
	return i18n.T("crazyfourpoker.result" + strings.ToUpper(domain.CrazyFourPokerResultName(r)[:1]) +
		domain.CrazyFourPokerResultName(r)[1:])
}

// crazyFourPokerRankName は 4 枚役の表示名を返す。
func crazyFourPokerRankName(rank int) string {
	if rank <= 0 || rank >= len(domain.FourCardHandNames) {
		return i18n.T("crazyfourpoker.rankNone")
	}
	return domain.FourCardHandNames[rank]
}

// crazyFourPokerCardsStr は札の並びを 1 行の文字列にする。
func crazyFourPokerCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}
