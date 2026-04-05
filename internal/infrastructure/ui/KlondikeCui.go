package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewKlondikeCui コンストラクタ
func NewKlondikeCui() *genericCuiGame {
	klondike := domain.NewKlondike(domain.NewTrumpCards(0))
	kc := controller.NewKlondikeCuiController(usecase.NewKlondikeInteractor(klondike, new(presenter.KlondikeCuiPresenter)))
	return newCuiGame(kc, []string{
		i18n.T("klondike.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("klondike.helpDraw"),
		i18n.T("klondike.helpMove"),
		i18n.T("klondike.helpMoveWF"),
		i18n.T("klondike.helpMoveTF"),
		i18n.T("klondike.helpMoveTT"),
		i18n.T("klondike.helpGiveUp"),
		i18n.T("klondike.helpHint"),
		i18n.T("klondike.helpAutoComplete"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
