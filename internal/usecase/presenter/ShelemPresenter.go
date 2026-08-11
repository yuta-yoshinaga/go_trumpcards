//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ShelemPresenter シェレムプレゼンターインタフェース
type ShelemPresenter interface {
	GamePresenter[interfaces.ShelemGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.ShelemGame) string
}
