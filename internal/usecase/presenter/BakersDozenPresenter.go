//go:build !js || !wasm || solo

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BakersDozenPresenter ベーカーズダズンプレゼンターインタフェース
type BakersDozenPresenter interface {
	GamePresenter[interfaces.BakersDozenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(bd interfaces.BakersDozenGame) string
	// TargetsOutput 列 col の一番下の札を置ける先を一覧出力する
	TargetsOutput(bd interfaces.BakersDozenGame, col int) string
}
