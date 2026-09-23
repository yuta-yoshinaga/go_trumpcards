//go:build test

package domain

// CardRankPublic カードランク取得 (テスト用公開メソッド)
func (e *Omi) CardRankPublic(card *Card) int { return e.cardRank(card) }

// EffectiveSuitPublic 実効スート取得 (テスト・互換用公開メソッド)
func (e *Omi) EffectiveSuitPublic(card *Card) int { return card.GetDesign() }
