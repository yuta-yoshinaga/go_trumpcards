//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CincinnatiPresenter シンシナティプレゼンターインタフェース
type CincinnatiPresenter interface {
	GamePresenter[interfaces.CincinnatiGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.CincinnatiGame) string
}
