//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DesmochePresenter デスモチェプレゼンターインタフェース
type DesmochePresenter interface {
	GamePresenter[interfaces.DesmocheGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.DesmocheGame) string
}
