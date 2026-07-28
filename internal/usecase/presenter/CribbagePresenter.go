//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CribbagePresenter クリベッジプレゼンターインタフェース
type CribbagePresenter interface {
	GamePresenter[interfaces.CribbageGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CribbageGame) string
}
