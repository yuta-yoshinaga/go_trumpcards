//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TienLenPresenter Tien Lenプレゼンターインタフェース
type TienLenPresenter interface {
	GamePresenter[interfaces.TienLenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(tg interfaces.TienLenGame) string
}
