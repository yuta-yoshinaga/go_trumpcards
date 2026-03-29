package domain

import (
	"math/rand"
)

// CalcOmahaEquity モンテカルロシミュレーションによるオマハエクイティ計算
// humanCards: 人間の手札(4枚), communityCards: コミュニティカード,
// activePlayers: アクティブ相手プレイヤー数, simulations: シミュレーション回数,
// rng: 乱数生成器 (nilの場合はグローバルrand使用)
func CalcOmahaEquity(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
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
	neededCards := remainingCommunity + activePlayers*4

	totalWins, totalHandCounts := runParallelSimulations(simulations, rng,
		func(sims int, localRng *rand.Rand) (float64, []int) {
			wins := 0.0
			handCounts := make([]int, len(PokerHandNames))
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

				// 人間のハンド評価 (オマハルール: 2枚+3枚)
				humanRank, humanBest := evalBestFromOmaha(humanCards, simCommunity)
				handCounts[humanRank]++

				// 相手のハンド評価
				humanWins := true
				for o := 0; o < activePlayers; o++ {
					oppHole := shufflePool[idx : idx+4]
					idx += 4
					oppRank, oppBest := evalBestFromOmaha(oppHole, simCommunity)
					if oppRank > humanRank || (oppRank == humanRank && compareHighCardsSlice(oppBest, humanBest) > 0) {
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

	return buildEquityResult(totalWins, totalHandCounts, simulations)
}

// evalBestFromOmaha オマハルールでベスト5枚のハンドランクと手を評価
// ホールカードから必ず2枚、コミュニティカードから必ず3枚を使う
func evalBestFromOmaha(holeCards, communityCards []*Card) (int, []*Card) {
	if len(holeCards) < 2 || len(communityCards) < 3 {
		return PokerHandHighCard, nil
	}

	holePairs := combinations(holeCards, 2)
	commTriples := combinations(communityCards, 3)

	bestRank := -1
	var bestCards []*Card

	for _, pair := range holePairs {
		for _, triple := range commTriples {
			hand := make([]*Card, 0, 5)
			hand = append(hand, pair...)
			hand = append(hand, triple...)
			rank := evalFiveCardHand(hand)
			if rank > bestRank || (rank == bestRank && compareHighCardsSlice(hand, bestCards) > 0) {
				bestRank = rank
				bestCards = make([]*Card, 5)
				copy(bestCards, hand)
			}
		}
	}

	return bestRank, bestCards
}
