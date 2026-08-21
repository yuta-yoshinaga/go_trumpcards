//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// RankAndFilePresenter ランク・アンド・ファイルプレゼンターインタフェース
type RankAndFilePresenter interface {
	GamePresenter[interfaces.RankAndFileGame]
	// HintOutput ヒント情報を出力する
	HintOutput(ft interfaces.RankAndFileGame) string
}
