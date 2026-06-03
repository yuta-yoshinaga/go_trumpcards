//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CruelPresenter クルーエルプレゼンターインタフェース
type CruelPresenter interface {
	GamePresenter[interfaces.CruelGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.CruelGame) string
}
