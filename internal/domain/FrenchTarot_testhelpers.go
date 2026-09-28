//go:build test

package domain

// ComputeBreakdownPublic 現在のディールの得点内訳を返す (テスト用)。
func (g *FrenchTarot) ComputeBreakdownPublic() FrenchTarotBreakdown { return g.computeBreakdown() }

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *FrenchTarot) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *FrenchTarot) LedSuitPublic() int { return g.ledSuit() }

// FrenchTarotBidMultPublic は入札倍率を返す (テスト用)。
func FrenchTarotBidMultPublic(bid FrenchTarotBid) int { return frenchTarotBidMult(bid) }

// FrenchTarotCardHalfPointsPublic はカードのハーフポイントを返す (テスト用)。
func FrenchTarotCardHalfPointsPublic(c *Card) int { return frenchTarotCardHalfPoints(c) }

// FrenchTarotIsBoutPublic はカードがブーか返す (テスト用)。
func FrenchTarotIsBoutPublic(c *Card) bool { return frenchTarotIsBout(c) }

// FrenchTarotIsTrumpPublic はカードが切り札か返す (テスト用)。
func FrenchTarotIsTrumpPublic(c *Card) bool { return frenchTarotIsTrump(c) }

// FrenchTarotIsExcusePublic はカードがエクスキューズか返す (テスト用)。
func FrenchTarotIsExcusePublic(c *Card) bool { return frenchTarotIsExcuse(c) }

// BuildFrenchTarotDeckPublic は 78 枚デッキを構築する (テスト用)。
func BuildFrenchTarotDeckPublic() []*Card { return buildFrenchTarotDeck() }
