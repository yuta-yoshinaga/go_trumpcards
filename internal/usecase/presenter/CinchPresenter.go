//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CinchPresenter はチンチ (Cinch) のプレゼンターインタフェース。
type CinchPresenter interface {
	GamePresenter[interfaces.CinchGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CinchGame) string
}
