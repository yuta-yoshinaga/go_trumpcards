//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// GrandfathersClockPresenter グランドファーザーズ・クロック プレゼンターインタフェース
type GrandfathersClockPresenter interface {
	GamePresenter[interfaces.GrandfathersClockGame]
	// HintOutput ヒント情報を出力する
	HintOutput(gc interfaces.GrandfathersClockGame) string
}
