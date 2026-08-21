//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BigBenPresenter ビッグ・ベン プレゼンターインタフェース
type BigBenPresenter interface {
	GamePresenter[interfaces.BigBenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(gc interfaces.BigBenGame) string
}
