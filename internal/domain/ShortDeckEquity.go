package domain

import (
	"math/rand"
)

// shortDeckEquitySimulations デフォルトのモンテカルロシミュレーション回数
const shortDeckEquitySimulations = 50000

// CalcShortDeckEquity モンテカルロシミュレーションによるショートデックエクイティ計算
// humanCards: 人間の手札(2枚), communityCards: コミュニティカード,
// activePlayers: アクティブ相手プレイヤー数, simulations: シミュレーション回数,
// rng: 乱数生成器 (nilの場合はグローバルrand使用)
func CalcShortDeckEquity(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
	if activePlayers == 0 {
		return HoldemEquityResult{
			Equity:   1.0,
			HandOdds: buildEmptyShortDeckHandOdds(),
		}
	}
	if simulations == 0 {
		return HoldemEquityResult{
			Equity:   0.0,
			HandOdds: buildEmptyShortDeckHandOdds(),
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

	// 未知カードプール構築 (ショートデック: A,6,7,8,9,10,J,Q,K のみ)
	pool := make([]*Card, 0, len(ShortDeckValues)*4-len(knownSet))
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for _, v := range ShortDeckValues {
			if !knownSet[[2]int{d, v}] {
				pool = append(pool, NewCard(d, v, false))
			}
		}
	}

	remainingCommunity := 5 - len(communityCards)
	neededCards := remainingCommunity + activePlayers*2

	totalWins, totalHandCounts := runParallelSimulations(simulations, rng,
		func(sims int, localRng *rand.Rand) (float64, []int) {
			wins := 0.0
			handCounts := make([]int, len(ShortDeckHandNames))
			shufflePool := make([]*Card, len(pool))
			simCommunity := make([]*Card, 0, 5)

			for i := 0; i < sims; i++ {
				copy(shufflePool, pool)
				shuffleCards(shufflePool, localRng)

				if neededCards > len(shufflePool) {
					continue
				}

				// シミュレーション用コミュニティカードを構築
				simCommunity = simCommunity[:0]
				simCommunity = append(simCommunity, communityCards...)
				idx := 0
				for j := 0; j < remainingCommunity; j++ {
					simCommunity = append(simCommunity, shufflePool[idx])
					idx++
				}

				// 人間のハンド評価 (ショートデック: C(7,5)=21通り)
				humanAll := make([]*Card, 0, 7)
				humanAll = append(humanAll, humanCards...)
				humanAll = append(humanAll, simCommunity...)
				humanRank, humanBest := evalBestFromShortDeck(humanAll)
				handCounts[humanRank]++

				// 相手のハンド評価
				humanWins := true
				for o := 0; o < activePlayers; o++ {
					oppCards := make([]*Card, 0, 7)
					oppCards = append(oppCards, shufflePool[idx], shufflePool[idx+1])
					idx += 2
					oppCards = append(oppCards, simCommunity...)
					oppRank, oppBest := evalBestFromShortDeck(oppCards)
					if oppRank > humanRank || (oppRank == humanRank && compareShortDeckHighCardsSlice(oppBest, humanBest) > 0) {
						humanWins = false
						break
					}
				}
				if humanWins {
					wins++
				}
			}
			return wins, handCounts
		})

	return buildShortDeckEquityResult(totalWins, totalHandCounts, simulations)
}

// evalBestFromShortDeck ショートデック用: 7枚からベスト5枚のハンドランクと手を評価
func evalBestFromShortDeck(cards []*Card) (int, []*Card) {
	if len(cards) < 5 {
		return ShortDeckHandHighCard, nil
	}
	combos := combinations(cards, 5)
	bestRank := -1
	var bestCards []*Card
	for _, combo := range combos {
		rank := evalShortDeckFiveCardHand(combo)
		if rank > bestRank || (rank == bestRank && compareShortDeckHighCardsSlice(combo, bestCards) > 0) {
			bestRank = rank
			bestCards = make([]*Card, 5)
			copy(bestCards, combo)
		}
	}
	return bestRank, bestCards
}

// buildShortDeckEquityResult 集約結果からHoldemEquityResultを構築 (ショートデック用ハンド名)
func buildShortDeckEquityResult(totalWins float64, totalHandCounts []int, simulations int) HoldemEquityResult {
	handOdds := make([]HoldemHandOdds, len(ShortDeckHandNames))
	for i := range ShortDeckHandNames {
		handOdds[i] = HoldemHandOdds{
			HandRank:    i,
			HandName:    ShortDeckHandNames[i],
			Probability: float64(totalHandCounts[i]) / float64(simulations),
		}
	}
	return HoldemEquityResult{
		Equity:   totalWins / float64(simulations),
		HandOdds: handOdds,
	}
}

// buildEmptyShortDeckHandOdds 空のハンドオッズを構築 (ショートデック用)
func buildEmptyShortDeckHandOdds() []HoldemHandOdds {
	handOdds := make([]HoldemHandOdds, len(ShortDeckHandNames))
	for i := 0; i < len(ShortDeckHandNames); i++ {
		handOdds[i] = HoldemHandOdds{
			HandRank: i,
			HandName: ShortDeckHandNames[i],
		}
	}
	return handOdds
}
