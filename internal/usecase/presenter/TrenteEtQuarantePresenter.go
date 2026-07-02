//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TrenteEtQuarantePresenter はトラント・エ・カラント (Trente et Quarante) の
// プレゼンターインタフェース。
type TrenteEtQuarantePresenter interface {
	GamePresenter[interfaces.TrenteEtQuaranteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TrenteEtQuaranteGame) string
}
