//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KoiKoiPresenter はこいこい (Koi-Koi) のプレゼンターインタフェース。
type KoiKoiPresenter interface {
	GamePresenter[interfaces.KoiKoiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.KoiKoiGame) string
}
