//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FortressPresenter Fortress プレゼンターインタフェース
type FortressPresenter interface {
	GamePresenter[interfaces.FortressGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bc interfaces.FortressGame) string
}
