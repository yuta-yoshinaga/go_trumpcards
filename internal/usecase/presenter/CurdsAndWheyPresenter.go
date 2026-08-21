//go:build !js || !wasm || classic

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CurdsAndWheyPresenter カーズ・アンド・ホエイのプレゼンターインタフェース。
type CurdsAndWheyPresenter interface {
	GamePresenter[interfaces.CurdsAndWheyGame]
	// HintOutput ヒント情報を出力する。
	HintOutput(g interfaces.CurdsAndWheyGame) string
}
