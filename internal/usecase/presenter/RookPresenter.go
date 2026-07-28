//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RookPresenter ルーク(Rook)プレゼンターインタフェース
type RookPresenter interface {
	GamePresenter[interfaces.RookGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.RookGame) string
}
