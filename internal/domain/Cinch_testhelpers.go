//go:build test

package domain

// Cinch_testhelpers.go は決定的なテストのための内部ロジック公開ヘルパー。

// CinchCardBeatsForTest は cardBeats を公開する (トリック比較の検証用)。
func CinchCardBeatsForTest(candidate, current *Card, leadSuit int) bool {
	g := &Cinch{trumpSuit: 0}
	// trumpSuit は leadSuit と同一の切り札コンテキストで比較したいので、
	// 切り札を leadSuit に合わせるのではなく、呼び出し側が leadSuit を切り札として
	// 渡すことを想定する。ここでは leadSuit を切り札スートとして扱う。
	g.trumpSuit = leadSuit
	return g.cardBeats(candidate, current, leadSuit)
}

// ValidatePlayForTest は validatePlay を公開する (フォロー規則の検証用)。
func (g *Cinch) ValidatePlayForTest(playerIdx int, c *Card) error {
	return g.validatePlay(playerIdx, c)
}

// CinchSameColorSuitForTest は cinchSameColorSuit を公開する。
func CinchSameColorSuitForTest(suit int) int { return cinchSameColorSuit(suit) }

// CinchPointValueForTest は cinchPointValue を公開する (得点札の点数検証用)。
func CinchPointValueForTest(c *Card, trumpSuit int) int { return cinchPointValue(c, trumpSuit) }
