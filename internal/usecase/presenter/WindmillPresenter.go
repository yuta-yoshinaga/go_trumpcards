//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WindmillPresenter ウィンドミル プレゼンターインタフェース
type WindmillPresenter interface {
	GamePresenter[interfaces.WindmillGame]
	// HintOutput ヒント情報を出力する
	HintOutput(w interfaces.WindmillGame) string
}
