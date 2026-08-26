//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// OmbrePresenter オンブル (Ombre) のプレゼンターインタフェース
type OmbrePresenter interface {
	GamePresenter[interfaces.OmbreGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.OmbreGame) string
}
