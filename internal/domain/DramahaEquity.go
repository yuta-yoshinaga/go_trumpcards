//go:build !js || !wasm || casino

package domain

import (
	"math/rand"
)

// CalcDramahaEquity モンテカルロシミュレーションによるドラマハエクイティ計算。
// humanCards: 人間の手札(5枚), communityCards: コミュニティカード,
// activePlayers: アクティブ相手プレイヤー数, simulations: シミュレーション回数,
// rng: 乱数生成器 (nilの場合はグローバルrand使用)。
//
// **相手にも 5 枚配る。** 以前はクローン元の Omaha に合わせて 4 枚を渡して
// おり、ドラマハの卓では起こりえない配りを前提に勝率を出していた。
func CalcDramahaEquity(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
	return calcDramahaEquityWithHoleCount(
		humanCards, communityCards, activePlayers, simulations, rng, DramahaHoleCards)
}

// calcDramahaEquityWithHoleCount はドラマハ系エクイティ計算の共通実装。
// holeCardCount で相手プレイヤーに配布するホールカード枚数を指定する
// ドラマハは常に 5 枚 (DramahaHoleCards)。
func calcDramahaEquityWithHoleCount(humanCards, communityCards []*Card, activePlayers, simulations int, rng *rand.Rand, holeCardCount int) HoldemEquityResult {
	dramahaEval := func(holeCards, simCommunity []*Card) (int, []*Card) {
		return evalBestFromDramaha(holeCards, simCommunity)
	}
	return calcEquityCore(humanCards, communityCards, activePlayers, simulations, rng, equityConfig{
		holeCardsPerOpponent: holeCardCount,
		handNames:            PokerHandNames,
		buildPool:            buildFullDeckPool,
		evalHuman:            dramahaEval,
		evalOpponent:         dramahaEval,
		compareHighCards:     compareHighCardsSlice,
	})
}

// evalBestFromDramaha ドラマハルールでベスト5枚のハンドランクと手を評価
// ホールカードから必ず2枚、コミュニティカードから必ず3枚を使う
func evalBestFromDramaha(holeCards, communityCards []*Card) (int, []*Card) {
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
