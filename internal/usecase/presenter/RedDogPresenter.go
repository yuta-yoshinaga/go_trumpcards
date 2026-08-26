//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RedDogPresenter レッドドッグプレゼンターインタフェース
type RedDogPresenter interface {
	GamePresenter[interfaces.RedDogGame]
	// HintOutput ヒント情報を出力する
	HintOutput(rd interfaces.RedDogGame) string
}
