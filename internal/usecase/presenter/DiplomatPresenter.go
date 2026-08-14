//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// DiplomatPresenter ディプロマット プレゼンターインタフェース
type DiplomatPresenter interface {
	GamePresenter[interfaces.DiplomatGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.DiplomatGame) string
}
