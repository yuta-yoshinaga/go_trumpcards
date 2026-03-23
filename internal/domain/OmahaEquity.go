package domain

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// omahaEquitySimulations デフォルトのモンテカルロシミュレーション回数
const omahaEquitySimulations = 50000

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

	// ワーカー数を決定しシミュレーションを分割
	numWorkers := runtime.NumCPU()
	if numWorkers > simulations {
		numWorkers = simulations
	}
	simsPerWorker := simulations / numWorkers

	// ワーカーごとのシードを事前に生成
	workerSeeds := make([]int64, numWorkers)
	for i := 0; i < numWorkers; i++ {
		if rng != nil {
			workerSeeds[i] = rng.Int63()
		} else {
			workerSeeds[i] = time.Now().UnixNano() + int64(i)
		}
	}

	// ワーカーごとの結果格納
	type workerResult struct {
		wins       float64
		handCounts []int
	}
	results := make([]workerResult, numWorkers)

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerIdx int) {
			defer wg.Done()

			localRng := rand.New(rand.NewSource(workerSeeds[workerIdx]))
			localSims := simsPerWorker
			// 最後のワーカーに余りを割り当て
			if workerIdx == numWorkers-1 {
				localSims = simulations - simsPerWorker*(numWorkers-1)
			}

			localWins := 0.0
			localHandCounts := make([]int, len(PokerHandNames))
			shufflePool := make([]*Card, len(pool))

			for i := 0; i < localSims; i++ {
				copy(shufflePool, pool)
				shuffleCards(shufflePool, localRng)

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

				// 人間のハンド評価 (オマハルール: 2枚+3枚)
				humanRank, humanBest := evalBestFromOmaha(humanCards, simCommunity)
				localHandCounts[humanRank]++

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
					localWins++
				}
			}

			results[workerIdx] = workerResult{
				wins:       localWins,
				handCounts: localHandCounts,
			}
		}(w)
	}
	wg.Wait()

	// 結果を集約
	totalWins := 0.0
	totalHandCounts := make([]int, len(PokerHandNames))
	for _, r := range results {
		totalWins += r.wins
		for i, c := range r.handCounts {
			totalHandCounts[i] += c
		}
	}

	// ハンドオッズ構築
	handOdds := make([]HoldemHandOdds, len(PokerHandNames))
	for i := 0; i < len(PokerHandNames); i++ {
		handOdds[i] = HoldemHandOdds{
			HandRank:    i,
			HandName:    PokerHandNames[i],
			Probability: float64(totalHandCounts[i]) / float64(simulations),
		}
	}

	return HoldemEquityResult{
		Equity:   totalWins / float64(simulations),
		HandOdds: handOdds,
	}
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
