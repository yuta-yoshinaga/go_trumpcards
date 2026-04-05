package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewFreeCellCui コンストラクタ
func NewFreeCellCui() *genericCuiGame {
	freeCell := domain.NewFreeCell(domain.NewTrumpCards(0))
	fc := controller.NewFreeCellCuiController(usecase.NewFreeCellInteractor(freeCell, new(presenter.FreeCellCuiPresenter)))
	return newCuiGame(fc, []string{
		i18n.T("freecell.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("freecell.helpMove"),
		i18n.T("freecell.helpMoveTF"),
		i18n.T("freecell.helpMoveTT"),
		i18n.T("freecell.helpMoveTC"),
		i18n.T("freecell.helpMoveCT"),
		i18n.T("freecell.helpMoveCF"),
		i18n.T("freecell.helpGiveUp"),
		i18n.T("freecell.helpHint"),
		i18n.T("freecell.helpAutoComplete"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
