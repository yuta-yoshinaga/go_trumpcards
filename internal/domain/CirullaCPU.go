//go:build !js || !wasm || extra3

package domain

// CirullaCPU.go は CPU の手番の方策。
//
// **点になるのは枚数・デナリ・7♦・プリミエラ・スコパ。** どれも「取った札」に
// 紐づくので、取れる手のうち一番価値の高いものを選び、取れないときは一番
// 安い札を置く。

// cirullaCaptureValue は捕獲の価値を返す (大きいほど良い)。
func cirullaCaptureValue(played *Card, taken []*Card, sweeps bool) int {
	score := 0
	if sweeps {
		score += 50 // スコパは 1 点そのもの
	}
	for _, card := range append([]*Card{played}, taken...) {
		score += 2 // 枚数
		if CirullaIsDenari(card) {
			score += 3
		}
		if CirullaIsSetteBello(card) {
			score += 20
		}
		if card != nil && card.GetValue() == 7 {
			score += 4 // プリミエラで効く
		}
	}
	return score
}

// cirullaChoice は 1 手の候補。
type cirullaChoice struct {
	handIdx  int
	captures []int
	value    int
}

// cpuChoose は CPU の手を返す。
func (c *Cirulla) cpuChoose(playerIdx int) (int, []int) {
	options := c.enumerateChoices(playerIdx)
	if len(options) == 0 {
		return 0, nil
	}
	if c.config.CpuDifficulty == CirullaCpuDifficultyEasy {
		pick := options[cirullaRandIntn(len(options))]
		return pick.handIdx, pick.captures
	}
	return c.smartChoose(playerIdx)
}

// smartChoose は難易度に関わらず「良い」手を返す。
func (c *Cirulla) smartChoose(playerIdx int) (int, []int) {
	options := c.enumerateChoices(playerIdx)
	if len(options) == 0 {
		return 0, nil
	}
	best := options[0]
	for _, o := range options[1:] {
		if o.value > best.value {
			best = o
		}
	}
	return best.handIdx, best.captures
}

// enumerateChoices は打てる手をすべて列挙する。
//
// **取れる札があるなら置く手は無い。** 規則で取ることが強制されるので、
// 候補に混ぜると CPU が打てない手を選びうる。
func (c *Cirulla) enumerateChoices(playerIdx int) []cirullaChoice {
	p := c.players[playerIdx]
	out := make([]cirullaChoice, 0, p.GetCardsSize())
	for i := 0; i < p.GetCardsSize(); i++ {
		card := p.GetCard(i)
		groups := EnumerateCirullaCaptures(card, c.table)
		if len(groups) == 0 {
			// 取れない札は置くしかない。安い札ほど置きやすい。
			out = append(out, cirullaChoice{handIdx: i, value: -CirullaCardValue(card)})
			continue
		}
		for _, g := range groups {
			taken := make([]*Card, 0, len(g))
			for _, idx := range g {
				taken = append(taken, c.table[idx])
			}
			sweeps := len(g) == len(c.table) && !c.isFinalPlay()
			out = append(out, cirullaChoice{
				handIdx:  i,
				captures: g,
				value:    cirullaCaptureValue(card, taken, sweeps),
			})
		}
	}
	// 取れる手が 1 つでもあるなら、置く手は落とす。
	hasCapture := false
	for _, o := range out {
		if len(o.captures) > 0 {
			hasCapture = true
			break
		}
	}
	if !hasCapture {
		return out
	}
	filtered := make([]cirullaChoice, 0, len(out))
	for _, o := range out {
		if len(o.captures) > 0 {
			filtered = append(filtered, o)
		}
	}
	return filtered
}
