//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GoofspielPresenter ゴフスピールプレゼンターインタフェース
type GoofspielPresenter interface {
	GamePresenter[interfaces.GoofspielGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.GoofspielGame) string
}
