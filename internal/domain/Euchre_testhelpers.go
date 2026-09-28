//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (e *Euchre) CardRankPublic(card *Card) int { return e.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト用公開メソッド)
func (e *Euchre) EffectiveSuitPublic(card *Card) int { return e.effectiveSuit(card) }
