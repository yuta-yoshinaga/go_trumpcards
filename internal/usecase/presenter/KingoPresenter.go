//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// KingoPresenter キンゴプレゼンターインタフェース
type KingoPresenter interface {
	GamePresenter[interfaces.KingoGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.KingoGame) string
}
