//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SomersetPresenter Somerset プレゼンターインタフェース
type SomersetPresenter interface {
	GamePresenter[interfaces.SomersetGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bc interfaces.SomersetGame) string
}
