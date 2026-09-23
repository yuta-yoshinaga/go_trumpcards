//go:build !js || !wasm || extra6

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CaribbeanDrawPresenter カリビアン・ドロー・ポーカープレゼンターインタフェース
type CaribbeanDrawPresenter interface {
	GamePresenter[interfaces.CaribbeanDrawGame]
	// HintOutput ヒント情報を出力する
	HintOutput(cs interfaces.CaribbeanDrawGame) string
	// ClearSession clears accumulated CUI session statistics.
	ClearSession()
}
