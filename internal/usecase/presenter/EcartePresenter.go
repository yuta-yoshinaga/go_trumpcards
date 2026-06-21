//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// EcartePresenter エカルテプレゼンターインタフェース
type EcartePresenter interface {
	GamePresenter[interfaces.EcarteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.EcarteGame) string
}
