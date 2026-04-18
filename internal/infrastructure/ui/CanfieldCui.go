package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewCanfieldCui コンストラクタ
func NewCanfieldCui() *genericCuiGame {
	canfield := domain.NewCanfield(domain.NewTrumpCards(0))
	cc := controller.NewCanfieldCuiController(usecase.NewCanfieldInteractor(canfield, new(presenter.CanfieldCuiPresenter)))
	return newCuiGame(cc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "canfield.helpTitle",
		CommandKeys: []string{
			"canfield.helpDraw",
			"canfield.helpMove",
			"canfield.helpMoveWF",
			"canfield.helpMoveRT",
			"canfield.helpMoveRF",
			"canfield.helpMoveTF",
			"canfield.helpMoveTT",
			"canfield.helpGiveUp",
			"canfield.helpHint",
			"canfield.helpAutoComplete",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
