//go:build !js || !wasm || extra2

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MissMilliganPresenter ミス・ミリガン プレゼンターインタフェース
type MissMilliganPresenter interface {
	GamePresenter[interfaces.MissMilliganGame]
	// HintOutput ヒント情報を出力する
	HintOutput(mm interfaces.MissMilliganGame) string
}
