//go:build test

package domain

// GetCallableKingSuitsForTest はテスト用に席 idx が呼べる王のスートを返す。
// GetCallableKingSuits と違いフェーズを問わない。
func (g *Quadrille) GetCallableKingSuitsForTest(idx int) []int {
	return g.callableKingSuits(idx)
}

// SetPartnerForTest はテスト用に味方の席と公開状態を設定する。
func (g *Quadrille) SetPartnerForTest(idx int, revealed bool) {
	g.partnerIdx = idx
	g.partnerRevealed = revealed
}

// SetRoiSeulForTest はテスト用に単独プレイを設定する。
func (g *Quadrille) SetRoiSeulForTest(v bool) { g.roiSeul = v }

// SameSideForTest はテスト用に 2 席が同じ陣営かを返す。
func (g *Quadrille) SameSideForTest(a, b int) bool { return g.sameSide(a, b) }

// SetCalledKingSuitForTest はテスト用に呼ばれた王のスートを設定する。
func (g *Quadrille) SetCalledKingSuitForTest(suit int) { g.calledKingSuit = suit }
