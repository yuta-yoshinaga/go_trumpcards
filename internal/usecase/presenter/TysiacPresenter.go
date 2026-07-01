//go:build !js || !wasm || extra

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// TysiacPresenter サウザンド (Tysiąc) のプレゼンターインタフェース
type TysiacPresenter interface {
	GamePresenter[interfaces.TysiacGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.TysiacGame) string
}
