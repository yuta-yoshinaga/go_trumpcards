//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (b *Coinche) CardRankPublic(card *Card) int { return b.cardRank(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (b *Coinche) CardPointsPublic(card *Card) int { return coincheCardPoints(card, b.trumpSuit) }
