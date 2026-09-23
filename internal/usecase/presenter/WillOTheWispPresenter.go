//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WillOTheWispPresenter ウィル・オ・ザ・ウィスププレゼンターインタフェース
type WillOTheWispPresenter interface {
	GamePresenter[interfaces.WillOTheWispGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.WillOTheWispGame) string
}
