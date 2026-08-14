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

// FreeBetBlackjackCuiPresenter フリーベット・ブラックジャックCUIプレゼンタークラス
type FreeBetBlackjackCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *FreeBetBlackjackCuiPresenter) Output(c interfaces.FreeBetBlackjackGame, lastErr error) string {
	return buildCuiOutput(i18n.T("freebet.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("freebet.phaseLine", "phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("freebet.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"chips", strconv.Itoa(c.GetChips())) + "\n")
		if c.GetAnteBet() > 0 {
			sb.WriteString(i18n.Tf("freebet.anteLine", "ante", strconv.Itoa(c.GetAnteBet())) + "\n")
		}
		cp.writeDealer(sb, c)
		cp.writeHands(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeDealer はディーラーの札を書き出す。
func (cp *FreeBetBlackjackCuiPresenter) writeDealer(sb *strings.Builder, c interfaces.FreeBetBlackjackGame) {
	cards := c.GetDealerCards()
	if len(cards) == 0 {
		return
	}
	sb.WriteString("----------\n")
	sb.WriteString(i18n.Tf("freebet.dealerLine",
		"cards", freeBetCardsStr(cards),
		"score", strconv.Itoa(c.GetDealerScore())) + "\n")
}

// writeHands はプレイヤーの手札を書き出す。
//
// **ハウス持ちのぶんを分けて出す。** 「いくら失うのか」がこのゲームで
// いちばん見せたい数字なので、合算した額だけを出すと意味が伝わらない。
func (cp *FreeBetBlackjackCuiPresenter) writeHands(sb *strings.Builder, c interfaces.FreeBetBlackjackGame) {
	hands := c.GetHands()
	if len(hands) == 0 {
		return
	}
	for i, h := range hands {
		mark := " "
		if i == c.GetActiveHandIdx() && c.GetPhase() == domain.FreeBetPhasePlay {
			mark = "*"
		}
		line := i18n.Tf("freebet.handLine",
			"mark", mark,
			"idx", strconv.Itoa(i+1),
			"cards", freeBetCardsStr(h.GetCards()),
			"score", strconv.Itoa(h.GetScore()),
			"bet", strconv.Itoa(h.GetBet()))
		if free := c.GetFreeBet(i); free > 0 {
			line += i18n.Tf("freebet.freeSuffix", "free", strconv.Itoa(free))
		}
		sb.WriteString(line + "\n")
	}
	if c.GetPhase() == domain.FreeBetPhasePlay {
		var avail []string
		if c.CanFreeDouble() {
			avail = append(avail, i18n.T("freebet.actionFreeDouble"))
		}
		if c.CanFreeSplit() {
			avail = append(avail, i18n.T("freebet.actionFreeSplit"))
		}
		if len(avail) > 0 {
			sb.WriteString(i18n.Tf("freebet.freeAvailableLine", "actions", strings.Join(avail, " / ")) + "\n")
		}
	}
}

// writeResult は決着と収支を書き出す。
func (cp *FreeBetBlackjackCuiPresenter) writeResult(sb *strings.Builder, c interfaces.FreeBetBlackjackGame) {
	if c.GetPhase() != domain.FreeBetPhaseResult {
		return
	}
	// **プレイヤーが出した金だけを数える。** ハウスの出資は失う対象ではない。
	staked := 0
	for _, h := range c.GetHands() {
		staked += h.GetBet()
	}
	if c.IsDealerPushed22() {
		sb.WriteString(i18n.T("freebet.dealer22Line") + "\n")
	}
	for i, r := range c.GetResults() {
		sb.WriteString(i18n.Tf("freebet.resultLine",
			"idx", strconv.Itoa(i+1),
			"result", i18n.T("freebet.result"+strings.ToUpper(domain.FreeBetResultName(r)[:1])+
				domain.FreeBetResultName(r)[1:])) + "\n")
	}
	net := c.GetPayout() - staked
	msg := i18n.Tf("freebet.netLine", "net", strconv.Itoa(net))
	if net >= 0 {
		sb.WriteString(color.Green(msg) + "\n")
	} else {
		sb.WriteString(color.Red(msg) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.T("freebet.brokeLine") + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *FreeBetBlackjackCuiPresenter) ActionLogOutput(c interfaces.FreeBetBlackjackGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *FreeBetBlackjackCuiPresenter) HintOutput(c interfaces.FreeBetBlackjackGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("freebet.hintNone")
	}
	return i18n.Tf("freebet.hint",
		"action", i18n.T("freebet.action"+strings.ToUpper(h.Action[:1])+h.Action[1:]),
		"reason", i18n.T("freebet."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *FreeBetBlackjackCuiPresenter) phaseStr(phase domain.FreeBetPhase) string {
	switch phase {
	case domain.FreeBetPhaseBet:
		return i18n.T("freebet.phaseBet")
	case domain.FreeBetPhasePlay:
		return i18n.T("freebet.phasePlay")
	case domain.FreeBetPhaseResult:
		return i18n.T("freebet.phaseResult")
	default:
		return i18n.T("freebet.phaseUnknown")
	}
}

// freeBetCardsStr は札の並びを 1 行の文字列にする。
func freeBetCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}
