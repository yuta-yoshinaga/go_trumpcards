//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// UnsunKarutaPresenter はうんすんカルタのプレゼンターインタフェース。
type UnsunKarutaPresenter interface {
	GamePresenter[interfaces.UnsunKarutaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.UnsunKarutaGame) string
}
