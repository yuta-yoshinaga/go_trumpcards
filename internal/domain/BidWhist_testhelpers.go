//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用)
func (g *BidWhist) CardRankPublic(card *Card) int { return g.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト用)
func (g *BidWhist) EffectiveSuitPublic(card *Card) int { return g.effectiveSuit(card) }
