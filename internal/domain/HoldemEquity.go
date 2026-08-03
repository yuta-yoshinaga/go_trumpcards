//go:build !js || !wasm || casino

package domain

import (
	"math/rand"
	"runtime"
	"sync"
)

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
	return calcEquityCore(humanCards, communityCards, activePlayers, simulations, rng, equityConfig{
		holeCardsPerOpponent: 2,
		handNames:            PokerHandNames,
		buildPool:            buildFullDeckPool,
		evalHuman:            evalSevenCardHand,
		evalOpponent:         evalSevenCardHand,
		compareHighCards:     compareHighCardsSlice,
	})
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
		rng = rand.New(rand.NewSource(rand.Int63()))
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
	var totalHandCounts []int
	for _, r := range results {
		totalWins += r.wins
		if totalHandCounts == nil {
			totalHandCounts = make([]int, len(r.handCounts))
		}
		for i, c := range r.handCounts {
			totalHandCounts[i] += c
		}
	}
	return totalWins, totalHandCounts
}
