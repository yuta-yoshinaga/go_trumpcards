//go:build test

package domain

// ApplyDeclarePenaltyForTest は宣言ペナルティを適用する (テスト用)。
func (g *PageOne) ApplyDeclarePenaltyForTest(playerIdx int) {
	g.applyDeclarePenalty(playerIdx)
}
