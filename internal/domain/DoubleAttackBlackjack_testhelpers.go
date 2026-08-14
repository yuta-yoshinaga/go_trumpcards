//go:build test

package domain

// SetupSplitForTest は「同じ数字 2 枚が配られ、追加ベットも済んだ」局面を組み立てる。
//
// **配りは乱数のままでは固定できない**ので、賭けまで進めてから手札を差し替える。
// 別パッケージ (プレゼンタ) からスプリットの表示を検証するための入口。
func (g *DoubleAttackBlackjack) SetupSplitForTest(value int) {
	hand := NewBlackJackHand()
	hand.AddCard(NewCard(CardDesignSpade, value, false))
	hand.AddCard(NewCard(CardDesignHeart, value, false))
	hand.SetBet(g.anteBet + g.attackBet)
	g.hands = []*BlackJackHand{hand}
	g.results = []DoubleAttackResult{DoubleAttackResultNone}
	g.activeHand = 0
	g.phase = DoubleAttackPhasePlay

	g.dealerHand = NewBlackJackHand()
	g.dealerHand.AddCard(NewCard(CardDesignClover, 13, false))
	g.dealerHand.AddCard(NewCard(CardDesignDiamond, 7, false)) // ハード 17 で止まる
	g.dealerHoleDealt = true
}
