//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CrazyQuiltPresenter クレイジーキルト プレゼンターインタフェース
type CrazyQuiltPresenter interface {
	GamePresenter[interfaces.CrazyQuiltGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.CrazyQuiltGame) string
}
