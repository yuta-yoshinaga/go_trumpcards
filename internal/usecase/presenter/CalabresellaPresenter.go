//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CalabresellaPresenter カラブレセッラ (Calabresella) のプレゼンターインタフェース
type CalabresellaPresenter interface {
	GamePresenter[interfaces.CalabresellaGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.CalabresellaGame) string
}
