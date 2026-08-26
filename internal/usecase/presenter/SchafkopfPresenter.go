//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SchafkopfPresenter シャーフコップのプレゼンターインタフェース
type SchafkopfPresenter interface {
	GamePresenter[interfaces.SchafkopfGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SchafkopfGame) string
}
