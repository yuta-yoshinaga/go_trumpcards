//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PasurPresenter パスールプレゼンターインタフェース
type PasurPresenter interface {
	GamePresenter[interfaces.PasurGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.PasurGame) string
}
