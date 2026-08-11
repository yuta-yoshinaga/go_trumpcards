//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HokmPresenter ホクムプレゼンターインタフェース
type HokmPresenter interface {
	GamePresenter[interfaces.HokmGame]
	// HintOutput ヒント情報を出力する
	HintOutput(h interfaces.HokmGame) string
}
