//go:build !js || !wasm || solo

package domain

import "math/rand"

// CometHint は人間への推奨手。
type CometHint struct {
	// HandIdx は勧める手札 (-1 = パスするしかない)。
	HandIdx int
	// Reason は理由の識別子。
	Reason string
}

// cometCardWeight は 1 枚の「抱えたくなさ」を返す。大きいほど先に出したい。
//
// **コメットと K は最後まで取っておく。** どちらも連なりを切って主導権を渡さず
// 次の先頭を自分で選べる札なので、序盤に切ると手札を捌く順番を失う。
// 逆に、抱えたまま上がられるとコメットは失点になる。
func cometCardWeight(c *Card) int {
	switch {
	case IsCometWild(c):
		return 0 // 最後に取っておく。
	case c.GetValue() >= CometMaxRank:
		return 1 // K も温存。
	default:
		return 10 + c.GetValue()
	}
}

// smartCometChoice は最も出したい札の位置を返す (-1 = 出せる札が無い)。
//
// **ヒントは難易度で鈍らせない。** CPU の難易度は CPU の腕であって、人間への
// 助言の質ではない ── ここを共有すると Easy の卓では助言が出鱈目になる。
func (c *Comet) smartCometChoice(seat int) int {
	idxs := c.PlayableIdxs(seat)
	if len(idxs) == 0 {
		return -1
	}
	hand := c.players[seat].GetHand()
	// 手札が残り 1 枚なら迷う余地はない ── 出せば上がり。
	best, bestW := idxs[0], cometCardWeight(hand[idxs[0]])
	for _, i := range idxs[1:] {
		if w := cometCardWeight(hand[i]); w > bestW {
			best, bestW = i, w
		}
	}
	return best
}

// CpuPlay は CPU が 1 手打つ。
func (c *Comet) CpuPlay() {
	if c.gameEndFlag || c.phase != CometPhasePlay {
		return
	}
	seat := c.currentIdx
	if c.players[seat].GetIsHuman() {
		return
	}
	idxs := c.PlayableIdxs(seat)
	if len(idxs) == 0 {
		c.applyPass(seat)
		return
	}
	pick := c.smartCometChoice(seat)
	if c.config.CpuDifficulty == CometCpuDifficultyEasy {
		pick = idxs[rand.Intn(len(idxs))] //nolint:gosec // ゲームの手選びに暗号強度は要らない
	}
	_ = c.applyPlay(seat, pick)
}

// GetHint は人間への推奨手を返す。
func (c *Comet) GetHint() *CometHint {
	human := findHumanIdx(c.players)
	if human < 0 || c.gameEndFlag || c.phase != CometPhasePlay || c.currentIdx != human {
		return &CometHint{HandIdx: -1, Reason: "none"}
	}
	idx := c.smartCometChoice(human)
	if idx < 0 {
		return &CometHint{HandIdx: -1, Reason: "pass"}
	}
	card := c.players[human].GetHand()[idx]
	switch {
	case c.players[human].GetCardsSize() == 1:
		return &CometHint{HandIdx: idx, Reason: "go_out"}
	case IsCometWild(card):
		return &CometHint{HandIdx: idx, Reason: "comet"}
	case card.GetValue() >= CometMaxRank:
		return &CometHint{HandIdx: idx, Reason: "king"}
	default:
		return &CometHint{HandIdx: idx, Reason: "follow"}
	}
}
