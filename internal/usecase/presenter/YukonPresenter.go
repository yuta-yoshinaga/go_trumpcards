//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// YukonPresenter ユーコンプレゼンターインタフェース
type YukonPresenter interface {
	GamePresenter[interfaces.YukonGame]
	// HintOutput ヒント情報を出力する
	HintOutput(y interfaces.YukonGame) string
}
