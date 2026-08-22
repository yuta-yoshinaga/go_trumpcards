//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SpeculationPresenter スペキュレーションプレゼンターインタフェース
type SpeculationPresenter interface {
	GamePresenter[interfaces.SpeculationGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.SpeculationGame) string
}
