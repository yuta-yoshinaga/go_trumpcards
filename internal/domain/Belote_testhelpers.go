//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (b *Belote) CardRankPublic(card *Card) int { return b.cardRank(card) }

// CardPointsPublic カード得点取得 (テスト用公開メソッド)
func (b *Belote) CardPointsPublic(card *Card) int { return beloteCardPoints(card, b.trumpSuit) }
