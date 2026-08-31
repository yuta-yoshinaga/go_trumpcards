//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BristolPresenter ブリストルプレゼンターインタフェース
type BristolPresenter interface {
	GamePresenter[interfaces.BristolGame]
	// HintOutput ヒント情報を出力する
	HintOutput(b interfaces.BristolGame) string
	// TargetsOutput 移動元ゾーン zone の列 col の札を置ける先を一覧出力する
	TargetsOutput(b interfaces.BristolGame, zone string, col int) string
}
