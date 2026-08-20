//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FlowerGardenPresenter Flower Garden プレゼンターインタフェース
type FlowerGardenPresenter interface {
	GamePresenter[interfaces.FlowerGardenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bc interfaces.FlowerGardenGame) string
}
