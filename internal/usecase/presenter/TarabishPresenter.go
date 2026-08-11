//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TarabishPresenter タラビッシュプレゼンターインタフェース
type TarabishPresenter interface {
	GamePresenter[interfaces.TarabishGame]
	// HintOutput ヒント情報を出力する
	HintOutput(t interfaces.TarabishGame) string
}
