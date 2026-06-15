//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// SpoilFivePresenter スポイル・ファイブのプレゼンターインタフェース
type SpoilFivePresenter interface {
	GamePresenter[interfaces.SpoilFiveGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.SpoilFiveGame) string
}
