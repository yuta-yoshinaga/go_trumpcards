//go:build !js || !wasm || casino

package domain

import (
	"math/rand"
)

// pokerShowdownHandSize は 5 カードドローで各プレイヤーが持つ枚数。
const pokerShowdownHandSize = 5

// CalcPokerEquity は 5 カードドロー用のエクイティをモンテカルロで求める。
//
// **ホールデム系と違い、コミュニティカードが無い。**各プレイヤーは自分の 5 枚
// だけで勝負するので、シミュレーションでは相手 1 人につき 5 枚を配り、
// 共有カード無しで比較する。
//
// humanCards: 人間の手札(5枚), activePlayers: アクティブな相手の人数,
// simulations: シミュレーション回数, rng: 乱数生成器 (nil ならグローバル)。
func CalcPokerEquity(humanCards []*Card, activePlayers, simulations int, rng *rand.Rand) HoldemEquityResult {
	return calcEquityCore(humanCards, nil, activePlayers, simulations, rng, equityConfig{
		holeCardsPerOpponent: pokerShowdownHandSize,
		handNames:            PokerHandNames,
		buildPool:            buildFullDeckPool,
		evalHuman:            evalFiveCardShowdown,
		evalOpponent:         evalFiveCardShowdown,
		compareHighCards:     compareHighCardsSlice,
	})
}

// evalFiveCardShowdown は 5 枚をそのまま評価する。共有カードは無いので
// simCommunity は使わない (エンジンの署名に合わせるための引数)。
func evalFiveCardShowdown(cards, _ []*Card) (int, []*Card) {
	if len(cards) < pokerShowdownHandSize {
		return PokerHandHighCard, nil
	}
	best := make([]*Card, len(cards))
	copy(best, cards)
	return evalFiveCardHand(best), best
}
