//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FourSeasonsPresenter フォーシーズンズプレゼンターインタフェース
type FourSeasonsPresenter interface {
	GamePresenter[interfaces.FourSeasonsGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.FourSeasonsGame) string
}
