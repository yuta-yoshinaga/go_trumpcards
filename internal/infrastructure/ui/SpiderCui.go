package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewSpiderCui コンストラクタ
func NewSpiderCui() *genericCuiGame {
	spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
	sc := controller.NewSpiderCuiController(usecase.NewSpiderInteractor(spider, new(presenter.SpiderCuiPresenter)))
	return newCuiGame(sc, []string{
		i18n.T("spider.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("spider.helpDeal"),
		i18n.T("spider.helpMove"),
		i18n.T("spider.helpGiveUp"),
		i18n.T("spider.helpHint"),
		i18n.T("spider.helpAutoComplete"),
		"  l                        action log",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	})
}
