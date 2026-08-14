//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SnapPresenter スナッププレゼンターインタフェース
type SnapPresenter interface {
	GamePresenter[interfaces.SnapGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SnapGame) string
}
