//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DoppelkopfPresenter ドッペルコップのプレゼンターインタフェース
type DoppelkopfPresenter interface {
	GamePresenter[interfaces.DoppelkopfGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.DoppelkopfGame) string
}
