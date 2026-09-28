//go:build test

package domain

// PlayerScartoValidateForTest はテスト用にスカルトの検証だけを実行する。
func (g *Scarto) PlayerScartoValidateForTest(cardIndices []int) error {
	return g.validateScarto(g.players[g.dealerIdx], cardIndices)
}

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *Scarto) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *Scarto) LedSuitPublic() int { return g.ledSuit() }

// ScartoCardHalfPointsPublic はカードのハーフポイントを返す (テスト用)。
func ScartoCardHalfPointsPublic(c *Card) int { return scartoCardHalfPoints(c) }

// ScartoIsBoutPublic はカードがブーか返す (テスト用)。
func ScartoIsBoutPublic(c *Card) bool { return scartoIsBout(c) }

// ScartoIsTrumpPublic はカードが切り札か返す (テスト用)。
func ScartoIsTrumpPublic(c *Card) bool { return scartoIsTrump(c) }

// ScartoIsExcusePublic はカードがエクスキューズか返す (テスト用)。
func ScartoIsExcusePublic(c *Card) bool { return scartoIsExcuse(c) }

// ScartoDiscardablePublic はカードが通常スカルトに出せるか返す (テスト用)。
func ScartoDiscardablePublic(c *Card) bool { return scartoDiscardable(c) }

// BuildScartoDeckPublic は 78 枚デッキを構築する (テスト用)。
func BuildScartoDeckPublic() []*Card { return buildScartoDeck() }
