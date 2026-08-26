//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BaccaratBanquePresenter はバカラ・バンクのプレゼンターインタフェース。
type BaccaratBanquePresenter interface {
	GamePresenter[interfaces.BaccaratBanqueGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.BaccaratBanqueGame) string
}
