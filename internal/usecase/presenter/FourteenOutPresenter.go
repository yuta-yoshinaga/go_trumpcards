//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FourteenOutPresenter はフォーティーンアウト・ソリティアのプレゼンターインタフェース。
type FourteenOutPresenter interface {
	GamePresenter[interfaces.FourteenOutGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.FourteenOutGame) string
}
