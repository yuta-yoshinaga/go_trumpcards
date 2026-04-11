package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// CanfieldPresenter キャンフィールドプレゼンターインタフェース
type CanfieldPresenter interface {
	GamePresenter[interfaces.CanfieldGame]
	// HintOutput ヒント情報を出力する
	HintOutput(c interfaces.CanfieldGame) string
}
