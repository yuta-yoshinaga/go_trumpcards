package domain

import "math/rand"

// equityConfig エクイティ計算の設定（ゲーム固有差異を注入）
type equityConfig struct {
	// holeCardsPerOpponent 相手1人あたりのホールカード枚数
	holeCardsPerOpponent int
	// handNames ハンドランク名配列
	handNames []string
	// buildPool 未知カードプールを構築する
	buildPool func(knownSet map[[2]int]bool) []*Card
	// evalHuman 人間のハンドを評価する (humanCards, simCommunity) -> (rank, bestCards)
	evalHuman func(humanCards, simCommunity []*Card) (int, []*Card)
	// evalOpponent 相手のハンドを評価する (oppHoleCards, simCommunity) -> (rank, bestCards)
	evalOpponent func(oppHoleCards, simCommunity []*Card) (int, []*Card)
	// compareHighCards ハイカード比較 (a, b) -> >0 if a wins
	compareHighCards func(a, b []*Card) int
}

// calcEquityCore エクイティ計算の共通エンジン
func calcEquityCore(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand, cfg equityConfig) HoldemEquityResult {
	if activePlayers == 0 {
		return HoldemEquityResult{
			Equity:   1.0,
			HandOdds: buildEmptyHandOddsFromNames(cfg.handNames),
		}
	}
	if simulations == 0 {
		return HoldemEquityResult{
			Equity:   0.0,
			HandOdds: buildEmptyHandOddsFromNames(cfg.handNames),
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

	pool := cfg.buildPool(knownSet)
	remainingCommunity := 5 - len(communityCards)
	neededCards := remainingCommunity + activePlayers*cfg.holeCardsPerOpponent

	totalWins, totalHandCounts := runParallelSimulations(simulations, rng,
		func(sims int, localRng *rand.Rand) (float64, []int) {
			wins := 0.0
			handCounts := make([]int, len(cfg.handNames))
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
				humanRank, humanBest := cfg.evalHuman(humanCards, simCommunity)
				handCounts[humanRank]++

				// 相手のハンド評価
				humanWins := true
				for o := 0; o < activePlayers; o++ {
					oppHole := shufflePool[idx : idx+cfg.holeCardsPerOpponent]
					idx += cfg.holeCardsPerOpponent
					oppRank, oppBest := cfg.evalOpponent(oppHole, simCommunity)
					if oppRank > humanRank || (oppRank == humanRank && cfg.compareHighCards(oppBest, humanBest) > 0) {
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

	return buildEquityResultFromNames(totalWins, totalHandCounts, simulations, cfg.handNames)
}

// buildFullDeckPool 標準52枚デッキからプールを構築
func buildFullDeckPool(knownSet map[[2]int]bool) []*Card {
	pool := make([]*Card, 0, CardCnt-len(knownSet))
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for v := 1; v <= CardValueMax; v++ {
			if !knownSet[[2]int{d, v}] {
				pool = append(pool, NewCard(d, v, false))
			}
		}
	}
	return pool
}

// buildShortDeckPool ショートデック(36枚)からプールを構築
func buildShortDeckPool(knownSet map[[2]int]bool) []*Card {
	pool := make([]*Card, 0, len(ShortDeckValues)*4-len(knownSet))
	for d := CardDesignSpade; d <= CardDesignDiamond; d++ {
		for _, v := range ShortDeckValues {
			if !knownSet[[2]int{d, v}] {
				pool = append(pool, NewCard(d, v, false))
			}
		}
	}
	return pool
}

// evalSevenCardHand 7枚合成パターンの評価ヘルパー
func evalSevenCardHand(humanCards, simCommunity []*Card) (int, []*Card) {
	all := make([]*Card, 0, 7)
	all = append(all, humanCards...)
	all = append(all, simCommunity...)
	return evalBestFromSeven(all)
}

// evalSevenCardShortDeck ショートデック用7枚評価ヘルパー
func evalSevenCardShortDeck(humanCards, simCommunity []*Card) (int, []*Card) {
	all := make([]*Card, 0, 7)
	all = append(all, humanCards...)
	all = append(all, simCommunity...)
	return evalBestFromShortDeck(all)
}

// buildEmptyHandOddsFromNames 空のハンドオッズを構築（汎用）
func buildEmptyHandOddsFromNames(handNames []string) []HoldemHandOdds {
	handOdds := make([]HoldemHandOdds, len(handNames))
	for i := range handNames {
		handOdds[i] = HoldemHandOdds{
			HandRank: i,
			HandName: handNames[i],
		}
	}
	return handOdds
}

// buildEquityResultFromNames 集約結果からHoldemEquityResultを構築（汎用）
func buildEquityResultFromNames(totalWins float64, totalHandCounts []int, simulations int, handNames []string) HoldemEquityResult {
	handOdds := make([]HoldemHandOdds, len(handNames))
	for i := range handNames {
		handOdds[i] = HoldemHandOdds{
			HandRank:    i,
			HandName:    handNames[i],
			Probability: float64(totalHandCounts[i]) / float64(simulations),
		}
	}
	return HoldemEquityResult{
		Equity:   totalWins / float64(simulations),
		HandOdds: handOdds,
	}
}
