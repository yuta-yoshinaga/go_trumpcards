package domain

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
)

// holdemEquitySimulations デフォルトのモンテカルロシミュレーション回数
const holdemEquitySimulations = 50000

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
			return wins, handCounts
		})

	return buildEquityResult(totalWins, totalHandCounts, simulations)
}

// CalcPotOdds ポットオッズを計算 (パーセンテージ 0-100)
func CalcPotOdds(pot, callAmount int) float64 {
	if callAmount <= 0 {
		return 0.0
	}
	return float64(callAmount) / float64(pot+callAmount) * 100.0
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

// equityWorkerFn シミュレーションワーカー関数の型
type equityWorkerFn func(sims int, rng *rand.Rand) (wins float64, handCounts []int)

// runParallelSimulations モンテカルロシミュレーションを複数ワーカーで並列実行
func runParallelSimulations(totalSims int, rng *rand.Rand, worker equityWorkerFn) (float64, []int) {
	numWorkers := runtime.NumCPU()
	if numWorkers > totalSims {
		numWorkers = totalSims
	}
	simsPerWorker := totalSims / numWorkers

	// rng が nil の場合、一度だけ新しい RNG を生成してシード源とする
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	workerSeeds := make([]int64, numWorkers)
	for i := range numWorkers {
		workerSeeds[i] = rng.Int63()
	}

	type workerResult struct {
		wins       float64
		handCounts []int
	}
	results := make([]workerResult, numWorkers)

	var wg sync.WaitGroup
	for w := range numWorkers {
		wg.Add(1)
		go func(workerIdx int) {
			defer wg.Done()

			localRng := rand.New(rand.NewSource(workerSeeds[workerIdx]))
			localSims := simsPerWorker
			if workerIdx == numWorkers-1 {
				localSims = totalSims - simsPerWorker*(numWorkers-1)
			}

			wins, handCounts := worker(localSims, localRng)
			results[workerIdx] = workerResult{wins: wins, handCounts: handCounts}
		}(w)
	}
	wg.Wait()

	totalWins := 0.0
	totalHandCounts := make([]int, len(PokerHandNames))
	for _, r := range results {
		totalWins += r.wins
		for i, c := range r.handCounts {
			totalHandCounts[i] += c
		}
	}
	return totalWins, totalHandCounts
}

// buildEquityResult 集約結果からHoldemEquityResultを構築
func buildEquityResult(totalWins float64, totalHandCounts []int, simulations int) HoldemEquityResult {
	handOdds := make([]HoldemHandOdds, len(PokerHandNames))
	for i := range PokerHandNames {
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
