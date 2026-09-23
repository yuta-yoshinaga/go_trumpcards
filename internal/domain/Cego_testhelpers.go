//go:build test

package domain

// ComputeBreakdownPublic 現在のディールの得点内訳を返す (テスト用)。
func (g *Cego) ComputeBreakdownPublic() CegoBreakdown { return g.computeBreakdown() }

// TrickWinnerPublic 現在のトリックの勝者を返す (テスト用)。
func (g *Cego) TrickWinnerPublic() int { return g.trickWinner() }

// LedSuitPublic 現在のトリックのリードスートを返す (テスト用)。
func (g *Cego) LedSuitPublic() int { return g.ledSuit() }

// CegoBidMultPublic は入札倍率を返す (テスト用)。
func CegoBidMultPublic(bid CegoBid) int { return cegoBidMult(bid) }

// CegoCardPointsPublic はカードのカードポイントを返す (テスト用)。
func CegoCardPointsPublic(c *Card) int { return cegoCardPoints(c) }

// CegoIsTrullPublic はカードがトゥルルか返す (テスト用)。
func CegoIsTrullPublic(c *Card) bool { return cegoIsTrull(c) }

// CegoIsTrumpPublic はカードが切り札か返す (テスト用)。
func CegoIsTrumpPublic(c *Card) bool { return cegoIsTrump(c) }

// CegoIsSkusPublic はカードがスキュースか返す (テスト用)。
func CegoIsSkusPublic(c *Card) bool { return cegoIsSkus(c) }

// CegoIsKingPublic はカードがスートのキングか返す (テスト用)。
func CegoIsKingPublic(c *Card) bool { return cegoIsKing(c) }

// BuildCegoDeckPublic は 54 枚デッキを構築する (テスト用)。
func BuildCegoDeckPublic() []*Card { return buildCegoDeck() }
