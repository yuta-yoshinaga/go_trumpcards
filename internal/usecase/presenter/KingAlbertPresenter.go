//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KingAlbertPresenter King Albert プレゼンターインタフェース
type KingAlbertPresenter interface {
	GamePresenter[interfaces.KingAlbertGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bc interfaces.KingAlbertGame) string
}
