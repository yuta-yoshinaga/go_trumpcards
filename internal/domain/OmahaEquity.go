//go:build !js || !wasm || casino

package domain

import (
	"math/rand"
)

// CalcOmahaEquity モンテカルロシミュレーションによるオマハエクイティ計算
// humanCards: 人間の手札(4枚), communityCards: コミュニティカード,
// activePlayers: アクティブ相手プレイヤー数, simulations: シミュレーション回数,
// rng: 乱数生成器 (nilの場合はグローバルrand使用)
func CalcOmahaEquity(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
	return calcOmahaEquityWithHoleCount(humanCards, communityCards, activePlayers, simulations, rng, 4)
}

// CalcOmahaHiLoEquity モンテカルロシミュレーションによるオマハ Hi-Lo エクイティ計算
// humanCards: 人間の手札(4枚), communityCards: コミュニティカード,
// activePlayers: アクティブ相手プレイヤー数, simulations: シミュレーション回数,
// rng: 乱数生成器 (nilの場合はグローバルrand使用)
func CalcOmahaHiLoEquity(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
	return calcOmahaHiLoEquityWithHoleCount(humanCards, communityCards, activePlayers, simulations, rng, 4)
}

// calcOmahaHiLoEquityWithHoleCount calculates Omaha Hi-Lo equity as the
// human player's expected share of the whole pot. LowProbability is the
// expected share of the low half, before its 50% pot weighting is applied.
func calcOmahaHiLoEquityWithHoleCount(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand, holeCardCount int) HoldemEquityResult {
	if activePlayers == 0 {
		return HoldemEquityResult{Equity: 1.0, HandOdds: buildEmptyHandOddsFromNames(PokerHandNames)}
	}
	if simulations == 0 {
		return HoldemEquityResult{HandOdds: buildEmptyHandOddsFromNames(PokerHandNames)}
	}

	knownSet := make(map[[2]int]bool)
	for _, c := range humanCards {
		knownSet[[2]int{c.GetDesign(), c.GetValue()}] = true
	}
	for _, c := range communityCards {
		knownSet[[2]int{c.GetDesign(), c.GetValue()}] = true
	}
	pool := buildFullDeckPool(knownSet)
	remainingCommunity := 5 - len(communityCards)
	neededCards := remainingCommunity + activePlayers*holeCardCount

	totalShare, totalLowShare, totalHandCounts := runParallelSimulations(simulations, rng,
		func(sims int, localRng *rand.Rand) (float64, float64, []int) {
			share, lowShare := 0.0, 0.0
			handCounts := make([]int, len(PokerHandNames))
			shufflePool := make([]*Card, len(pool))
			simCommunity := make([]*Card, 0, 5)
			if neededCards > len(shufflePool) {
				return share, lowShare, handCounts
			}
			for i := 0; i < sims; i++ {
				copy(shufflePool, pool)
				shuffleCards(shufflePool, localRng)
				simCommunity = append(simCommunity[:0], communityCards...)
				idx := 0
				for j := 0; j < remainingCommunity; j++ {
					simCommunity = append(simCommunity, shufflePool[idx])
					idx++
				}

				holes := make([][]*Card, activePlayers+1)
				holes[0] = humanCards
				for o := 1; o <= activePlayers; o++ {
					holes[o] = shufflePool[idx : idx+holeCardCount]
					idx += holeCardCount
				}
				highShare, lowSideShare, hasLow, humanRank := omahaHiLoShares(holes, simCommunity)
				handCounts[humanRank]++
				if hasLow {
					share += 0.5 * (highShare + lowSideShare)
					lowShare += lowSideShare
				} else {
					share += highShare
				}
			}
			return share, lowShare, handCounts
		})
	result := buildEquityResultFromNames(totalShare, totalHandCounts, simulations, PokerHandNames)
	result.LowProbability = totalLowShare / float64(simulations)
	return result
}

// omahaHiLoShares evaluates one completed Omaha Hi-Lo deal without mutating
// players. It returns the human high share, low share, whether a low exists,
// and the human high rank for hand-odds aggregation.
func omahaHiLoShares(holeCards [][]*Card, communityCards []*Card) (float64, float64, bool, int) {
	highRanks := make([]int, len(holeCards))
	highHands := make([][]*Card, len(holeCards))
	lowHands := make([][]*Card, len(holeCards))
	bestHigh := -1
	for i, hole := range holeCards {
		highRanks[i], highHands[i] = evalBestFromOmaha(hole, communityCards)
		if highRanks[i] > bestHigh {
			bestHigh = highRanks[i]
		}
		lowHands[i] = bestQualifyingOmahaLow(hole, communityCards)
	}

	highWinners := make([]int, 0, len(holeCards))
	var bestHighHand []*Card
	for i := range holeCards {
		if highRanks[i] != bestHigh {
			continue
		}
		if bestHighHand == nil || compareHighCardsSlice(highHands[i], bestHighHand) > 0 {
			bestHighHand = highHands[i]
			highWinners = highWinners[:0]
			highWinners = append(highWinners, i)
		} else if compareHighCardsSlice(highHands[i], bestHighHand) == 0 {
			highWinners = append(highWinners, i)
		}
	}
	highShare := winnerShare(0, highWinners)

	lowWinners := make([]int, 0, len(holeCards))
	var bestLowHand []*Card
	for i, hand := range lowHands {
		if hand == nil {
			continue
		}
		if bestLowHand == nil || compareRazzCards(hand, bestLowHand) < 0 {
			bestLowHand = hand
			lowWinners = lowWinners[:0]
			lowWinners = append(lowWinners, i)
		} else if compareRazzCards(hand, bestLowHand) == 0 {
			lowWinners = append(lowWinners, i)
		}
	}
	return highShare, winnerShare(0, lowWinners), len(lowWinners) > 0, highRanks[0]
}

func winnerShare(player int, winners []int) float64 {
	for _, winner := range winners {
		if winner == player {
			return 1.0 / float64(len(winners))
		}
	}
	return 0
}

// bestQualifyingOmahaLow finds the best Omaha 8-or-better low without changing
// OmahaPlayer state.
func bestQualifyingOmahaLow(holeCards, communityCards []*Card) []*Card {
	if len(holeCards) < 2 || len(communityCards) < 3 {
		return nil
	}
	var best []*Card
	for _, pair := range combinations(holeCards, 2) {
		for _, triple := range combinations(communityCards, 3) {
			hand := make([]*Card, 0, 5)
			hand = append(hand, pair...)
			hand = append(hand, triple...)
			if isQualifyingOmahaLow(hand) && (best == nil || compareRazzCards(hand, best) < 0) {
				best = hand
			}
		}
	}
	return best
}

// calcOmahaEquityWithHoleCount はオマハ系エクイティ計算の共通実装。
// holeCardCount で相手プレイヤーに配布するホールカード枚数を指定する
// (オマハ=4, Big O=5)。CalcOmahaEquity は4枚版の薄いラッパー。
func calcOmahaEquityWithHoleCount(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand, holeCardCount int) HoldemEquityResult {
	omahaEval := func(holeCards, simCommunity []*Card) (int, []*Card) {
		return evalBestFromOmaha(holeCards, simCommunity)
	}
	return calcEquityCore(humanCards, communityCards, activePlayers, simulations, rng, equityConfig{
		holeCardsPerOpponent: holeCardCount,
		handNames:            PokerHandNames,
		buildPool:            buildFullDeckPool,
		evalHuman:            omahaEval,
		evalOpponent:         omahaEval,
		compareHighCards:     compareHighCardsSlice,
	})
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
