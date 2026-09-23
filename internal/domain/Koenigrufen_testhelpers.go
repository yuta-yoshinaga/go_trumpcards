//go:build test

package domain

// ComputeBreakdownPublic 現在のディールの得点内訳を返す (テスト用)。
func (g *Koenigrufen) ComputeBreakdownPublic() KoenigrufenBreakdown { return g.computeBreakdown() }

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *Koenigrufen) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *Koenigrufen) LedSuitPublic() int { return g.ledSuit() }

// KoenigrufenBidMultPublic は入札倍率を返す (テスト用)。
func KoenigrufenBidMultPublic(bid KoenigrufenBid) int { return koenigrufenBidMult(bid) }

// KoenigrufenCardPointsPublic はカードのカードポイントを返す (テスト用)。
func KoenigrufenCardPointsPublic(c *Card) int { return koenigrufenCardPoints(c) }

// KoenigrufenIsTrullPublic はカードがトゥルルか返す (テスト用)。
func KoenigrufenIsTrullPublic(c *Card) bool { return koenigrufenIsTrull(c) }

// KoenigrufenIsTrumpPublic はカードが切り札か返す (テスト用)。
func KoenigrufenIsTrumpPublic(c *Card) bool { return koenigrufenIsTrump(c) }

// KoenigrufenIsSkusPublic はカードがスキュースか返す (テスト用)。
func KoenigrufenIsSkusPublic(c *Card) bool { return koenigrufenIsSkus(c) }

// KoenigrufenIsKingPublic はカードがスートのキングか返す (テスト用)。
func KoenigrufenIsKingPublic(c *Card) bool { return koenigrufenIsKing(c) }

// BuildKoenigrufenDeckPublic は 54 枚デッキを構築する (テスト用)。
func BuildKoenigrufenDeckPublic() []*Card { return buildKoenigrufenDeck() }
