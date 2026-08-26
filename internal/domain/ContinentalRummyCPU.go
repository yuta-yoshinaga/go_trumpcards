//go:build !js || !wasm || extra2

package domain

import (
	"fmt"
	"math/rand"
)

// continentalRummyCpuLoopCap は CPU の手番を回す上限。
//
// **上限の無いループを Worker に持ち込まない。** 山は 106 − 61 = 45 枚しか
// 無いので 4 席がそれを使い切っても 45 手番で必ず終わるが、規則を書き換えた
// 誰かがそれを崩したときに固まるより、打ち切って盤を見せるほうがましなので、
// 十分に余裕のある壁を置く。
const continentalRummyCpuLoopCap = 400

// runCpuTurns は人間の番になるまで CPU に打たせる。
func (c *ContinentalRummy) runCpuTurns() {
	for n := 0; n < continentalRummyCpuLoopCap; n++ {
		if c.gameEndFlag || c.phase == ContinentalRummyPhaseRoundEnd {
			return
		}
		if c.currentIdx == ContinentalRummyHumanIdx {
			return
		}
		c.playCpuTurn(c.currentIdx)
	}
}

// playCpuTurn は 1 席ぶんの手番を打つ。
func (c *ContinentalRummy) playCpuTurn(seat int) {
	p := c.players[seat]

	// **引く前に、配られたままで上がれるかを見る。** ここを飛ばすと
	// いちばん重い加点 (10 点) に誰も届かない。
	if p.GetCardsSize() == ContinentalRummyHandSize && CanContinentalRummyGoOut(p.GetHand()) {
		c.layDownAndGoOut(seat, -1)
		return
	}

	// **引く前に、捨て札を取ったら上がれるかを見る。** ここを見ないと、
	// 目の前の 1 枚で完成する手を素通りする。
	if top := c.GetDiscardTop(); top != nil && c.discardCompletesFor(p, top) {
		c.discardPile = c.discardPile[:len(c.discardPile)-1]
		p.AddCard(top)
		c.drewThisRound[seat] = true
	} else {
		card := c.drawStock()
		if card == nil {
			c.finishRound(-1)
			return
		}
		p.AddCard(card)
		c.drewThisRound[seat] = true
	}

	if i, ok := continentalGoOutDiscard(p.GetHand()); ok {
		c.layDownAndGoOut(seat, i)
		return
	}

	i := c.cpuDiscardIdx(p)
	card := p.RemoveCard(i)
	c.discardPile = append(c.discardPile, card)
	c.turnsThisRound[seat]++
	c.appendLog(seat, "discard", fmt.Sprintf("seat %d discards", seat), []*Card{card})
	c.advanceNoRecurse()
}

// advanceNoRecurse は手番だけ next へ送る。runCpuTurns のループが続きを回す。
func (c *ContinentalRummy) advanceNoRecurse() {
	c.currentIdx = (c.currentIdx + 1) % ContinentalRummyPlayerCnt
	c.phase = ContinentalRummyPhaseDraw
}

// discardCompletesFor は捨て札の 1 枚を取ったら上がれるかを返す。
func (c *ContinentalRummy) discardCompletesFor(p *ContinentalRummyPlayer, top *Card) bool {
	hand := append(append([]*Card(nil), p.GetHand()...), top)
	_, ok := continentalGoOutDiscard(hand)
	return ok
}

// cpuDiscardIdx は捨てる札を選ぶ。
//
// **セットが無い形式なので、価値は「同スートの近い札が何枚あるか」で決まる。**
// 孤立した札から捨てるのが素直で、Easy はそれをせずランダムに捨てる。
func (c *ContinentalRummy) cpuDiscardIdx(p *ContinentalRummyPlayer) int {
	hand := p.GetHand()
	if len(hand) == 0 {
		return 0
	}
	if c.config.CpuDifficulty == ContinentalRummyCpuDifficultyEasy {
		return rand.Intn(len(hand)) //nolint:gosec // 遊びの手なので暗号強度は要らない
	}

	worst, worstScore := 0, 1<<30
	for i, card := range hand {
		// ジョーカーは絶対に捨てない。どの穴でも埋められる。
		if IsContinentalRummyJoker(card) {
			continue
		}
		score := continentalNeighbourCount(hand, i)*100 - ContinentalRummyCardValue(card)
		if score < worstScore {
			worst, worstScore = i, score
		}
	}
	return worst
}

// continentalNeighbourCount は i 番の札と 1 本になりうる手札の枚数を返す。
func continentalNeighbourCount(hand []*Card, i int) int {
	card := hand[i]
	n := 0
	for j, other := range hand {
		if j == i || other == nil {
			continue
		}
		if IsContinentalRummyJoker(other) {
			n++
			continue
		}
		if other.GetDesign() != card.GetDesign() {
			continue
		}
		if d := other.GetValue() - card.GetValue(); d >= -2 && d <= 2 && d != 0 {
			n++
		}
	}
	return n
}

// ContinentalRummyHint はいまの局面での助言。
type ContinentalRummyHint struct {
	// DiscardIdx は捨てるとよい手札の位置。引く前なら -1。
	DiscardIdx int
	// GoOut は上がれるか。
	GoOut bool
	// Reason は理由の識別子。
	Reason string
}

// GetHint は人間の番の助言を返す。番でなければ nil。
func (c *ContinentalRummy) GetHint() *ContinentalRummyHint {
	if c.gameEndFlag || c.currentIdx != ContinentalRummyHumanIdx {
		return nil
	}
	p := c.players[ContinentalRummyHumanIdx]
	switch c.phase {
	case ContinentalRummyPhaseDraw:
		if c.CanGoOutOnTheDeal() {
			return &ContinentalRummyHint{DiscardIdx: -1, GoOut: true, Reason: "go_out_on_deal"}
		}
		if top := c.GetDiscardTop(); top != nil && c.discardCompletesFor(p, top) {
			return &ContinentalRummyHint{DiscardIdx: -1, Reason: "take_discard"}
		}
		return &ContinentalRummyHint{DiscardIdx: -1, Reason: "draw_stock"}
	case ContinentalRummyPhaseDiscard:
		if i, ok := continentalGoOutDiscard(p.GetHand()); ok {
			return &ContinentalRummyHint{DiscardIdx: i, GoOut: true, Reason: "go_out"}
		}
		return &ContinentalRummyHint{DiscardIdx: c.cpuDiscardIdx(p), Reason: "discard_loose"}
	}
	return nil
}
