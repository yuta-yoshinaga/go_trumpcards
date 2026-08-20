//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PreferencePresenter プレフェランスのプレゼンターインタフェース
type PreferencePresenter interface {
	GamePresenter[interfaces.PreferenceGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.PreferenceGame) string
}
