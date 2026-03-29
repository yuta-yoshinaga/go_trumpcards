package domain

import (
	"math/rand"
)

// CalcShortDeckEquity モンテカルロシミュレーションによるショートデックエクイティ計算
// humanCards: 人間の手札(2枚), communityCards: コミュニティカード,
// activePlayers: アクティブ相手プレイヤー数, simulations: シミュレーション回数,
// rng: 乱数生成器 (nilの場合はグローバルrand使用)
func CalcShortDeckEquity(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
	return calcEquityCore(humanCards, communityCards, activePlayers, simulations, rng, equityConfig{
		holeCardsPerOpponent: 2,
		handNames:            ShortDeckHandNames,
		buildPool:            buildShortDeckPool,
		evalHuman:            evalSevenCardShortDeck,
		evalOpponent:         evalSevenCardShortDeck,
		compareHighCards:     compareShortDeckHighCardsSlice,
	})
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
