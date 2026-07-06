package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// WizardPresenter ウィザードプレゼンターインタフェース
type WizardPresenter interface {
	GamePresenter[interfaces.WizardGame]
	// HintOutput ヒント情報を出力する
	HintOutput(o interfaces.WizardGame) string
}
