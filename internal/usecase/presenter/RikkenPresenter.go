//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RikkenPresenter リッケンプレゼンターインタフェース
type RikkenPresenter interface {
	GamePresenter[interfaces.RikkenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.RikkenGame) string
}
