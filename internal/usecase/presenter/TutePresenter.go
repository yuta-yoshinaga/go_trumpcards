//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TutePresenter トゥーテのプレゼンターインタフェース
type TutePresenter interface {
	GamePresenter[interfaces.TuteGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TuteGame) string
}
