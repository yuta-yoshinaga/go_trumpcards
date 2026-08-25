//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CirullaPresenter はチルッラのプレゼンターインタフェース。
type CirullaPresenter interface {
	GamePresenter[interfaces.CirullaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CirullaGame) string
}
