//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ContinentalRummyPresenter はコンチネンタル・ラミーのプレゼンターインタフェース。
type ContinentalRummyPresenter interface {
	GamePresenter[interfaces.ContinentalRummyGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ContinentalRummyGame) string
}
