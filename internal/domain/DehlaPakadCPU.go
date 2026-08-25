//go:build !js || !wasm || extra

package domain

// DehlaPakadCPU.go は CPU の手番の方策。
//
// **狙うのはトリックではなく 10 と、山を引き取る「2 連勝」。** トリックを
// 取っただけでは札は手に入らないので、「いま山に 10 が乗っているか」と
// 「直前のトリックを味方が取っているか」で価値が変わる。

// cpuPickTrump は最初の 5 枚から切り札を選ぶ。
func (d *DehlaPakad) cpuPickTrump(playerIdx int) int {
	if d.config.CpuDifficulty == DehlaPakadCpuDifficultyEasy {
		return CardDesignSpade + dehlaPakadRandIntn(4)
	}
	return d.smartTrumpFor(playerIdx)
}

// smartTrumpFor は難易度に関わらず「良い」切り札を返す。
//
// **枚数が主、強さが従。** 5 枚しか見ていない段階では、長いスートを切り札に
// するほうが確実に効く。
func (d *DehlaPakad) smartTrumpFor(playerIdx int) int {
	p := d.players[playerIdx]
	if p == nil {
		return CardDesignSpade
	}
	var count, strength [CardDesignDiamond + 1]int
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		if c == nil || c.GetDesign() < CardDesignSpade || c.GetDesign() > CardDesignDiamond {
			continue
		}
		count[c.GetDesign()]++
		strength[c.GetDesign()] += DehlaPakadCardStrength(c)
	}
	best := CardDesignSpade
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if count[suit] > count[best] || (count[suit] == count[best] && strength[suit] > strength[best]) {
			best = suit
		}
	}
	return best
}

// cpuSelectCard は CPU が出す札を選ぶ。
func (d *DehlaPakad) cpuSelectCard(playerIdx int) int {
	valid := d.GetPlayableIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if d.config.CpuDifficulty == DehlaPakadCpuDifficultyEasy {
		return valid[dehlaPakadRandIntn(len(valid))]
	}
	return d.smartCardFor(playerIdx, valid)
}

// smartCardFor は難易度に関わらず「良い」札を返す。
func (d *DehlaPakad) smartCardFor(playerIdx int, valid []int) int {
	p := d.players[playerIdx]
	bestSeat, bestRank := d.currentTrickLeaderSeat()
	friendLeads := bestSeat >= 0 && DehlaPakadTeamOf(bestSeat) == DehlaPakadTeamOf(playerIdx)

	// **山に 10 が乗っているかで価値が変わる。** 乗っていて、しかも自分が
	// 直前のトリックを取っていれば、ここを取ると山ごと引き取れる。
	wantsIt := d.pileHoldsATen() || d.prevTrickWinner == playerIdx
	if friendLeads && !d.pileHoldsATen() {
		wantsIt = false
	}

	winning, losing := -1, -1
	for _, i := range valid {
		c := p.GetCard(i)
		rank := dehlaPakadWinRank(c, d.currentLeadSuit(), d.trumpSuit)
		if rank > bestRank {
			// 取れる札のうち一番安いもの。
			if winning < 0 || rank < dehlaPakadWinRank(p.GetCard(winning), d.currentLeadSuit(), d.trumpSuit) {
				winning = i
			}
			continue
		}
		// 取れない札のうち一番高いもの (10 は最後まで抱える)。
		if losing < 0 || d.dumpScore(p.GetCard(i)) > d.dumpScore(p.GetCard(losing)) {
			losing = i
		}
	}

	if wantsIt && winning >= 0 {
		return winning
	}
	// **味方が取っているなら 10 を乗せる。** 山ごと味方に渡るのが狙い。
	if friendLeads && d.prevTrickWinner == bestSeat {
		if idx := d.tenAmong(playerIdx, valid); idx >= 0 {
			return idx
		}
	}
	if losing >= 0 {
		return losing
	}
	if winning >= 0 {
		return winning
	}
	return valid[0]
}

// dumpScore は「捨てるならどれから」の順位を返す (10 は最後)。
func (d *DehlaPakad) dumpScore(c *Card) int {
	if c == nil {
		return -1
	}
	if c.GetValue() == DehlaPakadTenValue {
		return -1 // 10 は捨てない
	}
	return DehlaPakadCardStrength(c)
}

// tenAmong は候補のうち 10 のインデックスを返す (無ければ -1)。
func (d *DehlaPakad) tenAmong(playerIdx int, valid []int) int {
	p := d.players[playerIdx]
	for _, i := range valid {
		if c := p.GetCard(i); c != nil && c.GetValue() == DehlaPakadTenValue {
			return i
		}
	}
	return -1
}

// currentLeadSuit は進行中のトリックの台札スートを返す (-1 = リード前)。
func (d *DehlaPakad) currentLeadSuit() int {
	if len(d.currentTrick) == 0 || d.currentTrick[0].Card == nil {
		return -1
	}
	return d.currentTrick[0].Card.GetDesign()
}

// currentTrickLeaderSeat は現時点で勝っている席とその順位を返す。
func (d *DehlaPakad) currentTrickLeaderSeat() (int, int) {
	lead := d.currentLeadSuit()
	seat, best := -1, -1
	for _, tc := range d.currentTrick {
		if tc == nil {
			continue
		}
		if r := dehlaPakadWinRank(tc.Card, lead, d.trumpSuit); r > best {
			seat, best = tc.PlayerIdx, r
		}
	}
	return seat, best
}
