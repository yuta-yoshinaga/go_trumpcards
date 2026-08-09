//go:build !js || !wasm || extra

package domain

import "math/rand"

// KingCPU.go は CPU の意思決定 (コントラクト選択 / トリックプレイ) を担う。
// Easy 難易度は乱数を使う (テストは Easy を避けるか retry ループで両分岐を網羅する)。
// Normal / Hard は決定的なヒューリスティックで、テストの再現性を担保する。

// --- コントラクト選択 ---

// cpuSelectContract は親 (CPU) が選ぶコントラクトと切り札を返す。
// 未使用のコントラクトの中から、手札評価で最も期待値の高いものを選ぶ。
// Easy 難易度では未使用のうち最小 id を選ぶ (決定的)。
func (g *King) cpuSelectContract() (contract, trumpSuit int) {
	remaining := make([]int, 0, KingContractCnt)
	for c := 0; c < KingContractCnt; c++ {
		if !g.usedContracts[c] {
			remaining = append(remaining, c)
		}
	}
	if len(remaining) == 0 {
		return KingContractNoTricks, -1 // 到達しない (全コントラクト使い切る)
	}

	chosen := remaining[0]
	if g.config.CpuDifficulty != KingDifficultyEasy {
		bestScore := g.contractExpectedScore(g.dealerIdx, chosen)
		for _, c := range remaining[1:] {
			s := g.contractExpectedScore(g.dealerIdx, c)
			if s > bestScore {
				bestScore = s
				chosen = c
			}
		}
	}

	trump := -1
	if chosen == KingContractKingTrump {
		trump = g.longestSuit(g.dealerIdx)
	}
	return chosen, trump
}

// contractExpectedScore は手札からコントラクトの期待得点 (粗い見積り) を返す。
func (g *King) contractExpectedScore(playerIdx, contract int) int {
	p := g.players[playerIdx]
	highCards, highHearts, queens := 0, 0, 0
	hasKingHeart, lowCards, men := false, 0, 0
	for i := 0; i < p.GetCardsSize(); i++ {
		c := p.GetCard(i)
		v := c.GetValue()
		if v >= 11 || v == 1 { // J,Q,K,A は強い
			highCards++
		}
		if v <= 4 {
			lowCards++
		}
		if v == 11 || v == 13 {
			men++
		}
		if c.GetDesign() == CardDesignHeart {
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
	case KingContractNoTricks:
		return -highCards * KingNoTrickPenalty
	case KingContractNoHearts:
		return -highHearts * KingHeartPenalty
	case KingContractNoQueens:
		return -queens * KingQueenPenalty
	case KingContractKingHeart:
		if hasKingHeart {
			// K♥ を持っていれば早めに捨てやすく安全寄り。
			return -KingKingHeartPenalty / 4
		}
		return -KingKingHeartPenalty / 2
	case KingContractNoLastTwo:
		if lowCards >= 3 {
			return -KingLastTwoPenalty / 5
		}
		return -KingLastTwoPenalty
	case KingContractNoMen:
		return -men * KingMenPenalty
	case KingContractKingTrump:
		return (highCards + g.longestSuitLen(playerIdx)) * KingTrumpReward / 2
	default:
		return 0
	}
}

// longestSuit はプレイヤーの最長スートを返す。
func (g *King) longestSuit(playerIdx int) int {
	counts := map[int]int{}
	p := g.players[playerIdx]
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
func (g *King) longestSuitLen(playerIdx int) int {
	counts := map[int]int{}
	p := g.players[playerIdx]
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
func (g *King) getValidTrickIndices(playerIdx int) []int {
	p := g.players[playerIdx]
	var valid []int
	for i := 0; i < p.GetCardsSize(); i++ {
		if g.validateTrickPlay(playerIdx, p.GetCard(i)) == nil {
			valid = append(valid, i)
		}
	}
	return valid
}

// currentWinningCard は進行中トリックで現在勝っているカードを返す (空なら nil)。
func (g *King) currentWinningCard() *Card {
	return currentTrickWinnerCard(g.currentTrick, g)
}

// cpuSelectTrickCard は CPU がプレイするトリックカードのインデックスを返す。
func (g *King) cpuSelectTrickCard(playerIdx int) int {
	valid := g.getValidTrickIndices(playerIdx)
	if len(valid) == 0 {
		return 0
	}
	if len(valid) == 1 {
		return valid[0]
	}
	if g.config.CpuDifficulty == KingDifficultyEasy {
		return valid[rand.Intn(len(valid))]
	}
	wantWin := g.currentContract == KingContractKingTrump
	if wantWin {
		return g.cpuTrickWin(playerIdx, valid)
	}
	return g.cpuTrickAvoid(playerIdx, valid)
}

// cpuTrickWin はトリックを取りたいとき (King Trump) の選択。
func (g *King) cpuTrickWin(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return g.highestIndex(p, valid) // リードは最高位
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winning := g.currentWinningCard()
	// 勝てる最小のカードを選ぶ。
	bestWin, bestWinVal := -1, 999
	for _, idx := range valid {
		c := p.GetCard(idx)
		if g.cardBeats(c, winning, leadSuit) && kingCardStrength(c) < bestWinVal {
			bestWin, bestWinVal = idx, kingCardStrength(c)
		}
	}
	if bestWin >= 0 {
		return bestWin
	}
	return g.lowestIndex(p, valid) // 勝てないなら最低位を捨てる
}

// cpuTrickAvoid はトリックを避けたいとき (負のコントラクト) の選択。
func (g *King) cpuTrickAvoid(playerIdx int, valid []int) int {
	p := g.players[playerIdx]
	if len(g.currentTrick) == 0 {
		return g.lowestIndex(p, valid) // リードは最低位
	}
	leadSuit := g.currentTrick[0].Card.GetDesign()
	winning := g.currentWinningCard()
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
			if !g.cardBeats(c, winning, leadSuit) && kingCardStrength(c) > bestDuckVal {
				bestDuck, bestDuckVal = idx, kingCardStrength(c)
			}
		}
		if bestDuck >= 0 {
			return bestDuck
		}
		return g.lowestIndex(p, valid)
	}
	// ボイド: 最も危険なカードを捨てる。
	bestIdx, bestDanger := valid[0], -1
	for _, idx := range valid {
		d := g.discardDanger(p.GetCard(idx))
		if d > bestDanger {
			bestDanger, bestIdx = d, idx
		}
	}
	return bestIdx
}

// discardDanger はコントラクトに応じたカードの危険度を返す (高いほど捨てたい)。
func (g *King) discardDanger(c *Card) int {
	danger := kingCardStrength(c)
	switch g.currentContract {
	case KingContractNoHearts:
		if c.GetDesign() == CardDesignHeart {
			danger += 100
		}
	case KingContractNoQueens:
		if c.GetValue() == 12 {
			danger += 100
		}
	case KingContractKingHeart:
		if c.GetDesign() == CardDesignHeart && c.GetValue() == 13 {
			danger += 200
		}
	case KingContractNoMen:
		if c.GetValue() == 11 || c.GetValue() == 13 {
			danger += 100
		}
	}
	return danger
}

// highestIndex は valid の中で最高位カードのインデックスを返す。
func (g *King) highestIndex(p *KingPlayer, valid []int) int {
	best, bestVal := valid[0], -1
	for _, idx := range valid {
		if v := kingCardStrength(p.GetCard(idx)); v > bestVal {
			best, bestVal = idx, v
		}
	}
	return best
}

// lowestIndex は valid の中で最低位カードのインデックスを返す。
func (g *King) lowestIndex(p *KingPlayer, valid []int) int {
	best, bestVal := valid[0], 999
	for _, idx := range valid {
		if v := kingCardStrength(p.GetCard(idx)); v < bestVal {
			best, bestVal = idx, v
		}
	}
	return best
}

// --- Hint ---

// KingHint はヒント情報。
type KingHint struct {
	CardIndices []int  // 推奨カードインデックス (play フェーズ)
	Reason      string // ヒント理由キー
}

// GetHint は人間プレイヤーの手番における推奨アクションを返す。
func (g *King) GetHint() *KingHint {
	if g.phase != KingPhasePlay {
		return nil
	}
	turn := g.currentPlayer
	player := g.GetPlayer(turn)
	if player == nil || !player.GetIsHuman() {
		return nil
	}
	valid := g.getValidTrickIndices(turn)
	if len(valid) == 0 {
		return nil
	}
	idx := g.cpuTrickAvoid(turn, valid)
	reason := "avoid_low"
	if g.currentContract == KingContractKingTrump {
		idx = g.cpuTrickWin(turn, valid)
		reason = "win_high"
	}
	return &KingHint{CardIndices: []int{idx}, Reason: reason}
}
