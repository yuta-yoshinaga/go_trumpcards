package domain

import "math/rand"

// holdemEquitySimulations デフォルトのモンテカルロシミュレーション回数
const holdemEquitySimulations = 5000

// HoldemEquityResult エクイティ計算結果
type HoldemEquityResult struct {
	Equity   float64          // 勝率 (0.0 - 1.0)
	HandOdds []HoldemHandOdds // 各ハンドランクの確率
}

// HoldemHandOdds 各ハンドランクの確率
type HoldemHandOdds struct {
	HandRank    int     // ハンドランク
	HandName    string  // ハンド名
	Probability float64 // 確率 (0.0 - 1.0)
}

// CalcEquity モンテカルロシミュレーションによるエクイティ計算
// humanCards: 人間の手札, communityCards: コミュニティカード,
// activePlayers: アクティブ相手プレイヤー数, simulations: シミュレーション回数,
// rng: 乱数生成器 (nilの場合はグローバルrand使用)
func CalcEquity(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
	if activePlayers == 0 {
		return HoldemEquityResult{
			Equity:   1.0,
			HandOdds: buildEmptyHandOdds(),
		}
	}
	if simulations == 0 {
		return HoldemEquityResult{
			Equity:   0.0,
			HandOdds: buildEmptyHandOdds(),
		}
	}

	// 既知カードのセットを構築
	knownSet := make(map[[2]int]bool)
	for _, c := range humanCards {
		knownSet[[2]int{c.GetDesign(), c.GetValue()}] = true
	}
	for _, c := range communityCards {
		knownSet[[2]int{c.GetDesign(), c.GetValue()}] = true
	}

	// 未知カードプール構築
	pool := make([]*Card, 0, CardCnt-len(knownSet))
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= CardValueMax; v++ {
			if !knownSet[[2]int{d, v}] {
				pool = append(pool, NewCard(d, v, false))
			}
		}
	}

	remainingCommunity := 5 - len(communityCards)
	neededCards := remainingCommunity + activePlayers*2

	wins := 0.0
	handCounts := make([]int, len(PokerHandNames))

	shufflePool := make([]*Card, len(pool))

	for i := 0; i < simulations; i++ {
		copy(shufflePool, pool)
		shuffleCards(shufflePool, rng)

		if neededCards > len(shufflePool) {
			continue
		}

		// シミュレーション用コミュニティカードを構築
		simCommunity := make([]*Card, 0, 5)
		simCommunity = append(simCommunity, communityCards...)
		idx := 0
		for j := 0; j < remainingCommunity; j++ {
			simCommunity = append(simCommunity, shufflePool[idx])
			idx++
		}

		// 人間のハンド評価
		humanAll := make([]*Card, 0, 7)
		humanAll = append(humanAll, humanCards...)
		humanAll = append(humanAll, simCommunity...)
		humanRank, humanBest := evalBestFromSeven(humanAll)
		handCounts[humanRank]++

		// 相手のハンド評価
		humanWins := true
		for o := 0; o < activePlayers; o++ {
			oppCards := make([]*Card, 0, 7)
			oppCards = append(oppCards, shufflePool[idx], shufflePool[idx+1])
			idx += 2
			oppCards = append(oppCards, simCommunity...)
			oppRank, oppBest := evalBestFromSeven(oppCards)
			if oppRank > humanRank || (oppRank == humanRank && compareHighCardsSlice(oppBest, humanBest) > 0) {
				humanWins = false
				break
			}
		}
		if humanWins {
			wins++
		}
	}

	// ハンドオッズ構築
	handOdds := make([]HoldemHandOdds, len(PokerHandNames))
	for i := 0; i < len(PokerHandNames); i++ {
		handOdds[i] = HoldemHandOdds{
			HandRank:    i,
			HandName:    PokerHandNames[i],
			Probability: float64(handCounts[i]) / float64(simulations),
		}
	}

	return HoldemEquityResult{
		Equity:   wins / float64(simulations),
		HandOdds: handOdds,
	}
}

// CalcPotOdds ポットオッズを計算 (パーセンテージ 0-100)
func CalcPotOdds(pot, callAmount int) float64 {
	if callAmount == 0 {
		return 0.0
	}
	total := pot + callAmount
	if total == 0 {
		return 0.0
	}
	return float64(callAmount) / float64(total) * 100.0
}

// shuffleCards Fisher-Yatesシャッフル
func shuffleCards(cards []*Card, rng *rand.Rand) {
	for i := len(cards) - 1; i > 0; i-- {
		var j int
		if rng != nil {
			j = rng.Intn(i + 1)
		} else {
			j = rand.Intn(i + 1)
		}
		cards[i], cards[j] = cards[j], cards[i]
	}
}

// evalBestFromSeven 7枚からベスト5枚のハンドランクと手を評価
func evalBestFromSeven(cards []*Card) (int, []*Card) {
	if len(cards) < 5 {
		return PokerHandHighCard, nil
	}
	combos := combinations(cards, 5)
	bestRank := -1
	var bestCards []*Card
	for _, combo := range combos {
		rank := evalFiveCardHand(combo)
		if rank > bestRank || (rank == bestRank && compareHighCardsSlice(combo, bestCards) > 0) {
			bestRank = rank
			bestCards = make([]*Card, 5)
			copy(bestCards, combo)
		}
	}
	return bestRank, bestCards
}

// buildEmptyHandOdds 空のハンドオッズを構築
func buildEmptyHandOdds() []HoldemHandOdds {
	handOdds := make([]HoldemHandOdds, len(PokerHandNames))
	for i := 0; i < len(PokerHandNames); i++ {
		handOdds[i] = HoldemHandOdds{
			HandRank: i,
			HandName: PokerHandNames[i],
		}
	}
	return handOdds
}
