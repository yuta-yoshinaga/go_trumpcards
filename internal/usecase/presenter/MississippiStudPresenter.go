//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MississippiStudPresenter ミシシッピ・スタッドプレゼンターインタフェース
type MississippiStudPresenter interface {
	GamePresenter[interfaces.MississippiStudGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.MississippiStudGame) string
}
