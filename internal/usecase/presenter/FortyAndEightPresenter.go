//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FortyAndEightPresenter フォーティ・アンド・エイトプレゼンターインタフェース
type FortyAndEightPresenter interface {
	GamePresenter[interfaces.FortyAndEightGame]
	// HintOutput ヒント情報を出力する
	HintOutput(ft interfaces.FortyAndEightGame) string
}
