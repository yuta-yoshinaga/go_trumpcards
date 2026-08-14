//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// AuldLangSynePresenter オールド・ラング・サインプレゼンターインタフェース
type AuldLangSynePresenter interface {
	GamePresenter[interfaces.AuldLangSyneGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.AuldLangSyneGame) string
}
