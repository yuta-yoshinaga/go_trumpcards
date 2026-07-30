//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ChineseTenPresenter 撿紅點プレゼンターインタフェース
type ChineseTenPresenter interface {
	GamePresenter[interfaces.ChineseTenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.ChineseTenGame) string
}
