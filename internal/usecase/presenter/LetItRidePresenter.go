//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// LetItRidePresenter レット・イット・ライドプレゼンターインタフェース
type LetItRidePresenter interface {
	GamePresenter[interfaces.LetItRideGame]
	// PullConfirmOutput Pull 実行前の確認内容を出力する
	PullConfirmOutput(g interfaces.LetItRideGame) string
}
