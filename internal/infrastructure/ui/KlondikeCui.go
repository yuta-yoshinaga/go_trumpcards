package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// KlondikeCui クロンダイクCUIクラス
type KlondikeCui struct {
	kc *controller.KlondikeCuiController
}

// NewKlondikeCui コンストラクタ
func NewKlondikeCui() *KlondikeCui {
	klondike := domain.NewKlondike(domain.NewTrumpCards(0))
	return &KlondikeCui{
		kc: controller.NewKlondikeCuiController(usecase.NewKlondikeInteractor(klondike, presenter.NewKlondikeCuiPresenter())),
	}
}

// Exec ゲーム実行
func (cui *KlondikeCui) Exec() {
	RunCuiLoop(cui.kc, []string{
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"d・・・draw (stock → waste)",
		"m w t <col>・・・move waste → tableau",
		"m w f・・・move waste → foundation",
		"m t <col> f・・・move tableau → foundation",
		"m t <col> <idx> t <col>・・・move tableau → tableau",
		"g・・・give up",
		"h・・・hint",
		"ac・・・auto-complete",
		"l・・・action log",
	})
}
