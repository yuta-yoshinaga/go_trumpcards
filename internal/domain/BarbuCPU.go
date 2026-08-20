//go:build !js || !wasm || extra4

package domain

// BarbuCPU.go は CPU の意思決定 (コントラクト選択 / トリックプレイ / 7 並べ) を担う。
// 乱数を使わない決定的なヒューリスティックで、テストの再現性を担保する。

// --- コントラクト選択 ---

// cpuSelectContract はディーラー (CPU) が選ぶコントラクトと切り札を返す。
// 未使用のコントラクトの中から、手札評価で最も期待値の高いものを選ぶ。
// Easy 難易度では未使用のうち最小 id を選ぶ (決定的)。
func (b *Barbu) cpuSelectContract(dealerIdx int) (contract, trumpSuit int) {
	remaining := make([]int, 0, BarbuContractCnt)
	for c := 0; c < BarbuContractCnt; c++ {
		if !b.usedContracts[dealerIdx][c] {
			remaining = append(remaining, c)
		}
	}
	if len(remaining) == 0 {
		return BarbuContractNoTricks, -1 // 到達しない (28 ディールで使い切る)
	}

	chosen := remaining[0]
	if b.config.CpuDifficulty != BarbuDifficultyEasy {
		bestScore := b.contractExpectedScore(dealerIdx, chosen)
		for _, c := range remaining[1:] {
			s := b.contractExpectedScore(dealerIdx, c)
			if s > bestScore {
				bestScore = s
				chosen = c
			}
		}
	}

	trump := -1
	if chosen == BarbuContractTrumps {
		trump = b.longestSuit(dealerIdx)
	}
	return chosen, trump
}

// contractExpectedScore は手札からコントラクトの期待得点 (粗い見積り) を返す。
func (b *Barbu) contractExpectedScore(playerIdx, contract int) int {
	p := b.players[playerIdx]
	highCards, hearts, highHearts, queens := 0, 0, 0, 0
	hasKingHeart, lowCards, sevens := false, 0, 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		v := c.GetValue()
		if v >= 11 || v == 1 { // J,Q,K,A は強い
			highCards++
		}
		if v <= 4 {
			lowCards++
		}
		if v == barbuDominoSevenValue {
			sevens++
		}
		if c.GetDesign() == CardDesignHeart {
			hearts++
			if v >= 10 {
				highHearts++
			}
			if v == 13 {
				hasKingHeart = true
			}
		}
		if v == 12 {
			queens++
		}
	}

	switch contract {
	case BarbuContractNoTricks:
		return -highCards * BarbuNoTrickPenalty
	case BarbuContractNoHearts:
		return -highHearts * BarbuHeartPenalty
	case BarbuContractNoQueens:
		return -queens * BarbuQueenPenalty
	case BarbuContractKingHeart:
		if hasKingHeart {
			// K♥ を持っていれば捨て札にしやすく安全寄り。
			return -BarbuKingHeartPenalty / 4
		}
		return -BarbuKingHeartPenalty / 2
	case BarbuContractNoLastTrick:
		if lowCards >= 3 {
			return -BarbuLastTrickPenalty / 5
		}
		return -BarbuLastTrickPenalty / 2
	case BarbuContractTrumps:
		return (highCards + b.longestSuitLen(playerIdx)) * BarbuTrumpReward / 2
	case BarbuContractDominoes:
		return BarbuDominoScores[1] + sevens*5
	default:
		return 0
	}
}

// longestSuit はプレイヤーの最長スートを返す。
func (b *Barbu) longestSuit(playerIdx int) int {
	counts := map[int]int{}
	p := b.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	best, bestCnt := CardDesignSpade, -1
	for suit := CardDesignSpade; suit <= CardDesignDiamond; suit++ {
		if counts[suit] > bestCnt {
			bestCnt = counts[suit]
			best = suit
		}
	}
	return best
}

// longestSuitLen は最長スートの枚数を返す。
func (b *Barbu) longestSuitLen(playerIdx int) int {
	counts := map[int]int{}
	p := b.players[playerIdx]
	for i := 0; i < p.GetCardsSize(); i++ {
		counts[p.GetCard(i).GetDesign()]++
	}
	best := 0
	for _, c := range counts {
		if c > best {
			best = c
		}
	}
	return best
}

// --- トリックプレイ ---

// getValidTrickIndices はトリックコントラクトでプレイ可能な手札インデックスを返す。
func (b *Barbu) getValidTrickIndices(playerIdx int) []int {
	p := b.players[playerIdx]
	var valid []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if b.validateTrickPlay(playerIdx, p.GetCard(i)) == nil {
			valid = append(valid, i)
		}
	}
	return valid
}

// currentWinningCard は進行中トリックで現在勝っているカードを返す (空なら nil)。
func (b *Barbu) currentWinningCard() *Card {
	return currentTrickWinnerCard(b.currentTrick, b)
}

// cpuSelectTrickCard は CPU がプレイするトリックカードのインデックスを返す。
func (b *Barbu) cpuSelectTrickCard(playerIdx int) int {
	valid := b.getValidTrickIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	wantWin := b.currentContract == BarbuContractTrumps
	if wantWin {
		return b.cpuTrickWin(playerIdx, valid)
	}
	return b.cpuTrickAvoid(playerIdx, valid)
}

// cpuTrickWin はトリックを取りたいとき (Trumps) の選択。
func (b *Barbu) cpuTrickWin(playerIdx int, valid []int) int {
	p := b.players[playerIdx]
	if len(b.currentTrick) == 0 {
		return b.highestIndex(p, valid) // リードは最高位
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	winning := b.currentWinningCard()
	// 勝てる最小のカードを選ぶ。
	bestWin, bestWinVal := -1, 999
	for _, idx := range valid {
		c := p.GetCard(idx)
		if b.cardBeats(c, winning, leadSuit) && barbuCardStrength(c) < bestWinVal {
			bestWin, bestWinVal = idx, barbuCardStrength(c)
		}
	}
	if bestWin >= 0 {
		return bestWin
	}
	return b.lowestIndex(p, valid) // 勝てないなら最低位を捨てる
}

// cpuTrickAvoid はトリックを避けたいとき (負のコントラクト) の選択。
func (b *Barbu) cpuTrickAvoid(playerIdx int, valid []int) int {
	p := b.players[playerIdx]
	if len(b.currentTrick) == 0 {
		return b.lowestIndex(p, valid) // リードは最低位
	}
	leadSuit := b.currentTrick[0].Card.GetDesign()
	winning := b.currentWinningCard()
	hasLead := false
	for _, idx := range valid {
		if p.GetCard(idx).GetDesign() == leadSuit {
			hasLead = true
			break
		}
	}
	if hasLead {
		// 勝たない最高位 (アンダーカード) を出す。なければ最低位。
		bestDuck, bestDuckVal := -1, -1
		for _, idx := range valid {
			c := p.GetCard(idx)
			if !b.cardBeats(c, winning, leadSuit) && barbuCardStrength(c) > bestDuckVal {
				bestDuck, bestDuckVal = idx, barbuCardStrength(c)
			}
		}
		if bestDuck >= 0 {
			return bestDuck
		}
		return b.lowestIndex(p, valid)
	}
	// ボイド: 最も危険なカードを捨てる。
	bestIdx, bestDanger := valid[0], -1
	for _, idx := range valid {
		d := b.discardDanger(p.GetCard(idx))
		if d > bestDanger {
			bestDanger, bestIdx = d, idx
		}
	}
	return bestIdx
}

// discardDanger はコントラクトに応じたカードの危険度を返す (高いほど捨てたい)。
func (b *Barbu) discardDanger(c *Card) int {
	danger := barbuCardStrength(c)
	switch b.currentContract {
	case BarbuContractNoHearts:
		if c.GetDesign() == CardDesignHeart {
			danger += 100
		}
	case BarbuContractNoQueens:
		if c.GetValue() == 12 {
			danger += 100
		}
	case BarbuContractKingHeart:
		if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
			danger += 200
		}
	}
	return danger
}

// highestIndex は valid の中で最高位カードのインデックスを返す。
func (b *Barbu) highestIndex(p *BarbuPlayer, valid []int) int {
	best, bestVal := valid[0], -1
	for _, idx := range valid {
		if v := barbuCardStrength(p.GetCard(idx)); v > bestVal {
			best, bestVal = idx, v
		}
	}
	return best
}

// lowestIndex は valid の中で最低位カードのインデックスを返す。
func (b *Barbu) lowestIndex(p *BarbuPlayer, valid []int) int {
	best, bestVal := valid[0], 999
	for _, idx := range valid {
		if v := barbuCardStrength(p.GetCard(idx)); v < bestVal {
			best, bestVal = idx, v
		}
	}
	return best
}

// --- Dominoes プレイ ---

// cpuDominoPlay は CPU の 7 並べの 1 手 (配置 or パス) を実行する。
func (b *Barbu) cpuDominoPlay(playerIdx int) {
	idxs := b.GetDominoPlayableIndices(playerIdx)
	if len(idxs) == 0 {
		_ = b.applyDominoPlay(playerIdx, -1)
		return
	}
	p := b.players[playerIdx]
	// 7 から遠い (端の) カードを優先的に出して手札を早く軽くする。
	best, bestScore := idxs[0], -1
	for _, idx := range idxs {
		c := p.GetCard(idx)
		dist := c.GetValue() - barbuDominoSevenValue
		if dist < 0 {
			dist = -dist
		}
		if dist > bestScore {
			bestScore, best = dist, idx
		}
	}
	_ = b.applyDominoPlay(playerIdx, best)
}
