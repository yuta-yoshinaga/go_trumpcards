//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TehonbikiPresenter 手本引きプレゼンターインタフェース
type TehonbikiPresenter interface {
	GamePresenter[interfaces.TehonbikiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.TehonbikiGame) string
}
