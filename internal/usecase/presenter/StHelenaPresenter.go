//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// StHelenaPresenter セント・ヘレナ・ソリティアのプレゼンターインタフェース。
type StHelenaPresenter interface {
	GamePresenter[interfaces.StHelenaGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(cr interfaces.StHelenaGame) string
}
