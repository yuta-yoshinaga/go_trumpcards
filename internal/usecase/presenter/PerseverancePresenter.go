//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PerseverancePresenter パーシビアランスプレゼンターインタフェース
type PerseverancePresenter interface {
	GamePresenter[interfaces.PerseveranceGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bd interfaces.PerseveranceGame) string
	// TargetsOutput 列 col の一番下の札を置ける先を一覧出力する
	TargetsOutput(bd interfaces.PerseveranceGame, col int) string
}
