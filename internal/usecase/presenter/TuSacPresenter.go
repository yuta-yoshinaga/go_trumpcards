//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TuSacPresenter 四色牌プレゼンターインタフェース
type TuSacPresenter interface {
	GamePresenter[interfaces.TuSacGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.TuSacGame) string
}
