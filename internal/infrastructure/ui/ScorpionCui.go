package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewScorpionCui コンストラクタ
func NewScorpionCui() *genericCuiGame {
	scorpion := domain.NewScorpion(domain.NewTrumpCards(0))
	sc := controller.NewScorpionCuiController(usecase.NewScorpionInteractor(scorpion, new(presenter.ScorpionCuiPresenter)))
	return newCuiGame(sc, []string{
		i18n.T("scorpion.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("scorpion.helpMove"),
		i18n.T("scorpion.helpMoveTT"),
		i18n.T("scorpion.helpDeal"),
		i18n.T("scorpion.helpGiveUp"),
		i18n.T("scorpion.helpHint"),
		i18n.T("scorpion.helpAutoComplete"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
