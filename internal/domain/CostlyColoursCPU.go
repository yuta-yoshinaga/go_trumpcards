//go:build !js || !wasm || extra

package domain

import "math/rand"

// CostlyColoursHint は人間への推奨手。
type CostlyColoursHint struct {
	// HandIdx は勧める手札 (-1 = 出せる札が無い、または打つ場面でない)。
	HandIdx int
	// AcceptMog は交換に応じるべきか (mog フェーズのとき)。
	AcceptMog bool
	// Reason は理由の識別子。
	Reason string
}

// costlyPlayValue は 1 枚出したときに稼げる点を見積もる。
func (c *CostlyColours) costlyPlayValue(seat, handIdx int) int {
	hand := c.players[seat].GetHand()
	if handIdx < 0 || handIdx >= len(hand) {
		return -1
	}
	card := hand[handIdx]
	total := c.total + CostlyCardValue(card)
	if total > CostlyThirtyOne {
		return -1
	}
	pile := append(append([]*Card(nil), c.pile...), card)
	pts, _ := CostlyPlayScore(pile, total)
	// **31 ちょうどは数え上げを畳んで主導権を渡さない。** 同じ点なら
	// こちらを選ぶ。
	if total == CostlyThirtyOne {
		pts++
	}
	// **相手に 15・25・31 を渡さない。** 残りが節目まで届く距離だと
	// 次の一手で持っていかれる。
	for _, mark := range []int{CostlyFifteen, CostlyTwentyFive, CostlyThirtyOne} {
		if gap := mark - total; gap > 0 && gap <= 10 {
			pts--
			break
		}
	}
	return pts
}

// smartCostlyChoice は最も点になる札の位置を返す (-1 = 出せる札が無い)。
//
// **ヒントは難易度で鈍らせない。** CPU の難易度は CPU の腕であって、人間への
// 助言の質ではない。
func (c *CostlyColours) smartCostlyChoice(seat int) int {
	idxs := c.PlayableIdxs(seat)
	if len(idxs) == 0 {
		return -1
	}
	best, bestVal := idxs[0], c.costlyPlayValue(seat, idxs[0])
	for _, i := range idxs[1:] {
		if v := c.costlyPlayValue(seat, i); v > bestVal {
			best, bestVal = i, v
		}
	}
	return best
}

// CpuAct は CPU が 1 手打つ (交換の可否も含む)。
func (c *CostlyColours) CpuAct() {
	if c.gameEndFlag {
		return
	}
	switch c.phase {
	case CostlyColoursPhaseMog:
		seat := c.currentIdx
		if c.players[seat].GetIsHuman() {
			return
		}
		c.resolveMog(seat, c.cpuAcceptsMog(seat))
	case CostlyColoursPhasePlay:
		seat := c.currentIdx
		if c.players[seat].GetIsHuman() {
			return
		}
		idxs := c.PlayableIdxs(seat)
		if len(idxs) == 0 {
			return
		}
		pick := c.smartCostlyChoice(seat)
		if c.config.CpuDifficulty == CostlyColoursCpuDifficultyEasy {
			pick = idxs[rand.Intn(len(idxs))] //nolint:gosec // ゲームの手選びに暗号強度は要らない
		}
		_ = c.applyPlay(seat, pick)
	}
}

// cpuAcceptsMog は CPU が交換に応じるかを決める。
//
// **手札が良ければ断る。** 断れば相手に 1 点入るので、その 1 点より
// 手札の見込みが上回るときだけ断る ── 色役が既にできているなら崩さない。
func (c *CostlyColours) cpuAcceptsMog(seat int) bool {
	hand := c.players[seat].GetHand()
	if c.turnUp == nil {
		return true
	}
	show := append(append([]*Card(nil), hand...), c.turnUp)
	_, pts := CostlyColourCombo(show)
	return pts < CostlyThreeInSuitPoints
}

// GetHint は人間への推奨手を返す。
func (c *CostlyColours) GetHint() *CostlyColoursHint {
	human := findHumanIdx(c.players)
	if human < 0 || c.gameEndFlag {
		return &CostlyColoursHint{HandIdx: -1, Reason: "none"}
	}
	switch c.phase {
	case CostlyColoursPhaseMog:
		accept := c.cpuAcceptsMog(human)
		reason := "mog_accept"
		if !accept {
			reason = "mog_refuse"
		}
		return &CostlyColoursHint{HandIdx: -1, AcceptMog: accept, Reason: reason}
	case CostlyColoursPhasePlay:
		if c.currentIdx != human {
			return &CostlyColoursHint{HandIdx: -1, Reason: "none"}
		}
		idx := c.smartCostlyChoice(human)
		if idx < 0 {
			return &CostlyColoursHint{HandIdx: -1, Reason: "go"}
		}
		hand := c.players[human].GetHand()
		total := c.total + CostlyCardValue(hand[idx])
		switch total {
		case CostlyThirtyOne:
			return &CostlyColoursHint{HandIdx: idx, Reason: "thirty_one"}
		case CostlyTwentyFive:
			return &CostlyColoursHint{HandIdx: idx, Reason: "twenty_five"}
		case CostlyFifteen:
			return &CostlyColoursHint{HandIdx: idx, Reason: "fifteen"}
		}
		return &CostlyColoursHint{HandIdx: idx, Reason: "safe"}
	}
	return &CostlyColoursHint{HandIdx: -1, Reason: "none"}
}
