//go:build !js || !wasm || extra3

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// HasenpfefferPresenter ハーゼンプフェファープレゼンターインタフェース
type HasenpfefferPresenter interface {
	GamePresenter[interfaces.HasenpfefferGame]
	// HintOutput ヒント情報を出力する
	HintOutput(h interfaces.HasenpfefferGame) string
}
