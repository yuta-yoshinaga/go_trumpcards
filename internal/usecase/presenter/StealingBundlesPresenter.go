//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// StealingBundlesPresenter スティーリングバンドルプレゼンターインタフェース
type StealingBundlesPresenter interface {
	GamePresenter[interfaces.StealingBundlesGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.StealingBundlesGame) string
}
