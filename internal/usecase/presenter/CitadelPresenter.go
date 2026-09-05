//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CitadelPresenter Beleaguered Castle プレゼンターインタフェース
type CitadelPresenter interface {
	GamePresenter[interfaces.CitadelGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bc interfaces.CitadelGame) string
}
