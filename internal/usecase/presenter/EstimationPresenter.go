//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// EstimationPresenter エスティメーションプレゼンターインタフェース
type EstimationPresenter interface {
	GamePresenter[interfaces.EstimationGame]
	// HintOutput ヒント情報を出力する
	HintOutput(e interfaces.EstimationGame) string
}
