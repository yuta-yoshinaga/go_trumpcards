//go:build !js || !wasm || extra7

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// ZhengPresenter 争上游プレゼンターインタフェース
type ZhengPresenter interface {
	GamePresenter[interfaces.ZhengGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.ZhengGame) string
}
