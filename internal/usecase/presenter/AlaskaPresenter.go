//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AlaskaPresenter アラスカプレゼンターインタフェース
type AlaskaPresenter interface {
	GamePresenter[interfaces.AlaskaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(r interfaces.AlaskaGame) string
}
