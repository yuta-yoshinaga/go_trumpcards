//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CrescentPresenter クレセント・ソリティアのプレゼンターインタフェース。
type CrescentPresenter interface {
	GamePresenter[interfaces.CrescentGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(cr interfaces.CrescentGame) string
}
