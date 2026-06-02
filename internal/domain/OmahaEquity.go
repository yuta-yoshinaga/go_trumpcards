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
