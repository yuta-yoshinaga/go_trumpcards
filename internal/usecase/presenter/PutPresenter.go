//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PutPresenter プットプレゼンターインタフェース
type PutPresenter interface {
	GamePresenter[interfaces.PutGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.PutGame) string
}
