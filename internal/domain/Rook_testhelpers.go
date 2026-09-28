//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用)
func (g *Rook) CardRankPublic(card *Card) int { return g.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト用)
func (g *Rook) EffectiveSuitPublic(card *Card) int { return g.effectiveSuit(card) }

// CardPointsPublic カード得点取得 (テスト用)
func (g *Rook) CardPointsPublic(card *Card) int { return rookCardPoints(card) }
