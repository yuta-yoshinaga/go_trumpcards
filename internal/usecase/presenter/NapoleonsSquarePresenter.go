//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// NapoleonsSquarePresenter ナポレオンズ・スクエア プレゼンターインタフェース
type NapoleonsSquarePresenter interface {
	GamePresenter[interfaces.NapoleonsSquareGame]
	// HintOutput ヒント情報を出力する
	HintOutput(ns interfaces.NapoleonsSquareGame) string
}
