//go:build !js || !wasm || casino

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// FollowTheQueenPresenter フォロー・ザ・クイーンプレゼンターインタフェース
type FollowTheQueenPresenter interface {
	GamePresenter[interfaces.FollowTheQueenGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.FollowTheQueenGame) string
}
