//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AmericanToadPresenter アメリカン・トード プレゼンターインタフェース
type AmericanToadPresenter interface {
	GamePresenter[interfaces.AmericanToadGame]
	// HintOutput ヒント情報を出力する
	HintOutput(at interfaces.AmericanToadGame) string
}
