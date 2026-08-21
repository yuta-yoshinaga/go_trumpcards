//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SlyFoxPresenter スライ・フォックス プレゼンターインタフェース
type SlyFoxPresenter interface {
	GamePresenter[interfaces.SlyFoxGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.SlyFoxGame) string
}
