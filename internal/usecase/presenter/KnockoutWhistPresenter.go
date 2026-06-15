//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KnockoutWhistPresenter ノックアウト・ホイストのプレゼンターインタフェース
type KnockoutWhistPresenter interface {
	GamePresenter[interfaces.KnockoutWhistGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.KnockoutWhistGame) string
}
