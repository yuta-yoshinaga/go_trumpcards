//go:build !js || !wasm || extra4

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// PiedmonteseTarotPresenter はピエモンテ・タロッコのプレゼンターインタフェース。
type PiedmonteseTarotPresenter interface {
	GamePresenter[interfaces.PiedmonteseTarotGame]
	// HintOutput ヒント情報を出力する
	HintOutput(g interfaces.PiedmonteseTarotGame) string
}
