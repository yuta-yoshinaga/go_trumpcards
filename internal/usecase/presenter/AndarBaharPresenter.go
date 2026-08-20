//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AndarBaharPresenter アンダーバハールプレゼンターインタフェース
type AndarBaharPresenter interface {
	GamePresenter[interfaces.AndarBaharGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.AndarBaharGame) string
}
