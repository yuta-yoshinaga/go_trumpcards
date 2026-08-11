//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MendikotPresenter メンディコットプレゼンターインタフェース
type MendikotPresenter interface {
	GamePresenter[interfaces.MendikotGame]
	// HintOutput ヒント情報を出力する
	HintOutput(m interfaces.MendikotGame) string
}
