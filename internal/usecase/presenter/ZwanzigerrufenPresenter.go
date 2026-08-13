//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ZwanzigerrufenPresenter ツヴァンツィガールーフェンのプレゼンターインタフェース。
type ZwanzigerrufenPresenter interface {
	GamePresenter[interfaces.ZwanzigerrufenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ZwanzigerrufenGame) string
}
