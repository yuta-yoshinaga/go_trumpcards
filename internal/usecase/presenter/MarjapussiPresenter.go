//go:build !js || !wasm || extra5

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MarjapussiPresenter マルヤプッシ (Marjapussi) のプレゼンターインタフェース
type MarjapussiPresenter interface {
	GamePresenter[interfaces.MarjapussiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MarjapussiGame) string
}
