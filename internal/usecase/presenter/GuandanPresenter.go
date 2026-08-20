//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GuandanPresenter 掼蛋 (Guandan) プレゼンターインタフェース
//
// **役の下読みは CUI 専用** (#5734)。Web は選択に合わせてカード下に
// 出しているので、Web 側は既存の状態出力を返す。
type GuandanPresenter interface {
	GamePresenter[interfaces.GuandanGame]
	// CheckOutput 指定した手札の組み合わせが何の役になるかを出力する
	CheckOutput(g interfaces.GuandanGame, idxs []int) string
}
