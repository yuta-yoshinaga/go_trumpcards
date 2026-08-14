//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// IronCrossPresenter アイアンクロスプレゼンターインタフェース
type IronCrossPresenter interface {
	GamePresenter[interfaces.IronCrossGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.IronCrossGame) string
}
