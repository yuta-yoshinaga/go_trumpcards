package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewKlondikeCui コンストラクタ
func NewKlondikeCui() *genericCuiGame {
	klondike := domain.NewKlondike(domain.NewTrumpCards(0))
	kc := controller.NewKlondikeCuiController(usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeCuiPresenter)))
	return newCuiGame(kc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "klondike.helpTitle",
		CommandKeys: []string{
			"klondike.helpDraw",
			"klondike.helpMove",
			"klondike.helpMoveWF",
			"klondike.helpMoveTF",
			"klondike.helpMoveTT",
			"klondike.helpGiveUp",
			"klondike.helpHint",
			"klondike.helpAutoComplete",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
