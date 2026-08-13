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

// DoubleAttackBlackjackCuiPresenter 追加ベット・ブラックジャックCUIプレゼンタークラス
type DoubleAttackBlackjackCuiPresenter struct{}

// Output ゲーム状態を出力
func (cp *DoubleAttackBlackjackCuiPresenter) Output(c interfaces.DoubleAttackBlackjackGame, lastErr error) string {
	return buildCuiOutput(i18n.T("doubleattack.outputTitle"), func(sb *strings.Builder) {
		sb.WriteString(i18n.Tf("doubleattack.phaseLine",
			"phase", cp.phaseStr(c.GetPhase())) + "\n")
		sb.WriteString(i18n.Tf("doubleattack.roundLine",
			"round", strconv.Itoa(c.GetRoundNumber()),
			"chips", strconv.Itoa(c.GetChips())) + "\n")

		if c.GetAnteBet() > 0 {
			sb.WriteString(i18n.Tf("doubleattack.betLine",
				"ante", strconv.Itoa(c.GetAnteBet()),
				"attack", strconv.Itoa(c.GetAttackBet()),
				"bustIt", strconv.Itoa(c.GetBustItBet())) + "\n")
		}
		cp.writeDealer(sb, c)
		cp.writeHands(sb, c)
		cuiErrorBlock(sb, lastErr)
		cp.writeResult(sb, c)
	})
}

// writeDealer はディーラーの札を書き出す。
func (cp *DoubleAttackBlackjackCuiPresenter) writeDealer(sb *strings.Builder, c interfaces.DoubleAttackBlackjackGame) {
	cards := c.GetDealerCards()
	if len(cards) == 0 {
		return
	}
	sb.WriteString("----------\n")
	if !c.IsDealerHoleDealt() {
		// **アップカードだけ。** 2 枚目は追加ベットの後にしか存在しない。
		sb.WriteString(i18n.Tf("doubleattack.upCardLine",
			"card", cuiCardStr(cards[0])) + "\n")
		return
	}
	sb.WriteString(i18n.Tf("doubleattack.dealerLine",
		"cards", doubleAttackCardsStr(cards),
		"score", strconv.Itoa(c.GetDealerScore())) + "\n")
}

// writeHands はプレイヤーの手札を書き出す。
func (cp *DoubleAttackBlackjackCuiPresenter) writeHands(sb *strings.Builder, c interfaces.DoubleAttackBlackjackGame) {
	hands := c.GetHands()
	if len(hands) == 0 {
		return
	}
	for i, h := range hands {
		mark := " "
		if i == c.GetActiveHandIdx() && c.GetPhase() == domain.DoubleAttackPhasePlay {
			mark = "*"
		}
		sb.WriteString(i18n.Tf("doubleattack.handLine",
			"mark", mark,
			"idx", strconv.Itoa(i+1),
			"cards", doubleAttackCardsStr(h.GetCards()),
			"score", strconv.Itoa(h.GetScore()),
			"bet", strconv.Itoa(h.GetBet())) + "\n")
	}
	if c.GetPhase() == domain.DoubleAttackPhaseAttack {
		sb.WriteString(i18n.Tf("doubleattack.maxAttackLine",
			"max", strconv.Itoa(c.MaxAttackBet())) + "\n")
	}
}

// writeResult は決着と収支を書き出す。
func (cp *DoubleAttackBlackjackCuiPresenter) writeResult(sb *strings.Builder, c interfaces.DoubleAttackBlackjackGame) {
	if c.GetPhase() != domain.DoubleAttackPhaseResult {
		return
	}
	// **手札の賭け金をそのまま足す。**
	//
	// 以前は「アンティ + 追加ベット」を土台にして手札ごとの差分を足していたが、
	// **スプリットすると両方の手札が土台と同額**になるため差分が 0 になり、
	// 2 つ目の手札を作るのに払ったぶんが丸ごと消えていた (アンティ 50 で分割して
	// 両方プッシュ = 収支 0 が、+50 の黒字として表示される)。手札の `GetBet()` は
	// ダブルもスプリットも反映済みなので、素直に合計すればよい。
	staked := c.GetBustItBet()
	for _, h := range c.GetHands() {
		staked += h.GetBet()
	}
	net := c.GetPayout() - staked
	for i, r := range c.GetResults() {
		sb.WriteString(i18n.Tf("doubleattack.resultLine",
			"idx", strconv.Itoa(i+1),
			"result", i18n.T("doubleattack.result"+
				strings.ToUpper(domain.DoubleAttackResultName(r)[:1])+
				domain.DoubleAttackResultName(r)[1:])) + "\n")
	}
	if c.GetBustItPayout() > 0 {
		sb.WriteString(i18n.Tf("doubleattack.bustItLine",
			"payout", strconv.Itoa(c.GetBustItPayout())) + "\n")
	}
	msg := i18n.Tf("doubleattack.netLine", "net", strconv.Itoa(net))
	if net >= 0 {
		sb.WriteString(color.Green(msg) + "\n")
	} else {
		sb.WriteString(color.Red(msg) + "\n")
	}
	if c.GetGameEndFlag() {
		sb.WriteString(i18n.T("doubleattack.brokeLine") + "\n")
	}
}

// ActionLogOutput 棋譜をテキスト出力
func (cp *DoubleAttackBlackjackCuiPresenter) ActionLogOutput(c interfaces.DoubleAttackBlackjackGame) string {
	return actionLogOutputText(c)
}

// HintOutput ヒントをテキスト出力
func (cp *DoubleAttackBlackjackCuiPresenter) HintOutput(c interfaces.DoubleAttackBlackjackGame) string {
	h := c.GetHint()
	if h == nil {
		return i18n.T("doubleattack.hintNone")
	}
	return i18n.Tf("doubleattack.hint",
		"action", i18n.T("doubleattack.action"+strings.ToUpper(h.Action[:1])+h.Action[1:]),
		"reason", i18n.T("doubleattack."+h.Reason))
}

// phaseStr フェーズ文字列
func (cp *DoubleAttackBlackjackCuiPresenter) phaseStr(phase domain.DoubleAttackPhase) string {
	switch phase {
	case domain.DoubleAttackPhaseBet:
		return i18n.T("doubleattack.phaseBet")
	case domain.DoubleAttackPhaseAttack:
		return i18n.T("doubleattack.phaseAttack")
	case domain.DoubleAttackPhasePlay:
		return i18n.T("doubleattack.phasePlay")
	case domain.DoubleAttackPhaseResult:
		return i18n.T("doubleattack.phaseResult")
	default:
		return i18n.T("doubleattack.phaseUnknown")
	}
}

// doubleAttackCardsStr は札の並びを 1 行の文字列にする。
func doubleAttackCardsStr(cards []*domain.Card) string {
	parts := make([]string, 0, len(cards))
	for _, c := range cards {
		parts = append(parts, cuiCardStr(c))
	}
	return strings.Join(parts, " ")
}
