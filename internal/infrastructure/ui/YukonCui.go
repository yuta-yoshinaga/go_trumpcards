package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewYukonCui コンストラクタ
func NewYukonCui() *genericCuiGame {
	yukon := domain.NewYukon(domain.NewTrumpCards(0))
	yc := controller.NewYukonCuiController(usecase.NewYukonInteractor(yukon, new(presenter.YukonCuiPresenter)))
	return newCuiGame(yc, []string{
		i18n.T("yukon.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("yukon.helpMove"),
		i18n.T("yukon.helpMoveTF"),
		i18n.T("yukon.helpMoveTT"),
		i18n.T("yukon.helpGiveUp"),
		i18n.T("yukon.helpHint"),
		i18n.T("yukon.helpAutoComplete"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
