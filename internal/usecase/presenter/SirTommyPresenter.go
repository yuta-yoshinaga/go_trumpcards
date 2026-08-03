//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SirTommyPresenter サー・トミープレゼンターインタフェース
type SirTommyPresenter interface {
	GamePresenter[interfaces.SirTommyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SirTommyGame) string
}
