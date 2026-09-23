//go:build test

package domain

// ResolveForTest は手札を設定済みの状態でラウンドを解決する (テスト用)。乱数配札を
// 迂回して勝敗解決を決定的に検証するためのショートカット。
func (g *Primero) ResolveForTest() { g.resolveRound() }

// ClearBettingForTest は 1 ラウンドの賭けブックキーピング (raiseCount /
// actedSinceRaise / actionCount) をゼロに戻す (テスト用)。Reset() はディーラー左の
// CPU を人間の手番まで自動進行させるため残留カウンタが乗る。決定的な途中状態を
// 組み立てるテストはこれで残留をクリアする。
func (g *Primero) ClearBettingForTest() {
	g.state.raiseCount = 0
	g.state.actedSinceRaise = 0
	g.state.actionCount = 0
}
