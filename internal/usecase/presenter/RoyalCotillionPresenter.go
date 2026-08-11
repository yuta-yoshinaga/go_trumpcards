//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RoyalCotillionPresenter ロイヤルコティヨン プレゼンターインタフェース
type RoyalCotillionPresenter interface {
	GamePresenter[interfaces.RoyalCotillionGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.RoyalCotillionGame) string
}
