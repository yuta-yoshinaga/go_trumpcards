//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MariasPresenter マリアーシュのプレゼンターインタフェース
type MariasPresenter interface {
	GamePresenter[interfaces.MariasGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MariasGame) string
}
