//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MrsMopPresenter ミセス・モップソリティアプレゼンターインタフェース
type MrsMopPresenter interface {
	GamePresenter[interfaces.MrsMopGame]
	// HintOutput ヒント情報を出力する
	HintOutput(s interfaces.MrsMopGame) string
	// TargetsOutput 列 fromCol の札 cardIndex から始まる並びを置ける先を一覧出力する
	TargetsOutput(s interfaces.MrsMopGame, fromCol, cardIndex int) string
}
