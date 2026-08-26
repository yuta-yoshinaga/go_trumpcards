//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SalicLawPresenter サリカ法典 プレゼンターインタフェース
type SalicLawPresenter interface {
	GamePresenter[interfaces.SalicLawGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.SalicLawGame) string
}
