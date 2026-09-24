//go:build !js || !wasm || extra11

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FortyFivesPresenter オークション・フォーティファイブズのプレゼンターインタフェース
type FortyFivesPresenter interface {
	GamePresenter[interfaces.FortyFivesGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.FortyFivesGame) string
}
