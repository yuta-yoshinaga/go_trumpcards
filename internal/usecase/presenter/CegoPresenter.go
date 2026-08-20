//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CegoPresenter チェゴ (Cego) のプレゼンターインタフェース
type CegoPresenter interface {
	GamePresenter[interfaces.CegoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CegoGame) string
}
