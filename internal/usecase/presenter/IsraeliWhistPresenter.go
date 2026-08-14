//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// IsraeliWhistPresenter イスラエリホイストプレゼンターインタフェース
type IsraeliWhistPresenter interface {
	GamePresenter[interfaces.IsraeliWhistGame]
	// HintOutput ヒント情報を出力する
	HintOutput(w interfaces.IsraeliWhistGame) string
}
