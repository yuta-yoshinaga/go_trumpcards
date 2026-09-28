//go:build test

package domain

// SettleForTest は in フラグを設定済みの状態でラウンドを解決する (テスト用)。乱数配札を
// 迂回して勝敗解決・マッチ精算を決定的に検証するためのショートカット。
func (g *Guts) SettleForTest() { g.settle() }
