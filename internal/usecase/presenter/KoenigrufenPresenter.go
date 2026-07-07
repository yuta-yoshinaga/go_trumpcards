//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KoenigrufenPresenter ケーニッヒルーフェン (Königrufen) のプレゼンターインタフェース
type KoenigrufenPresenter interface {
	GamePresenter[interfaces.KoenigrufenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.KoenigrufenGame) string
}
