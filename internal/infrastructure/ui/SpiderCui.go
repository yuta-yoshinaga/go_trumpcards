package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewSpiderCui コンストラクタ
func NewSpiderCui() *genericCuiGame {
	sc := controller.NewSpiderCuiController(usecase.NewSpiderInteractor(domain.NewDefaultSpider(), new(presenter.SpiderCuiPresenter)))
	return newCuiGame(sc, BuildCuiHelp(CuiHelpSpec{
		TitleKey: "spider.helpTitle",
		CommandKeys: []string{
			"spider.helpDeal",
			"spider.helpMove",
			"spider.helpGiveUp",
			"spider.helpHint",
			"spider.helpAutoComplete",
		},
		ExtraCommandLines: []string{"  l                        action log"},
	}))
}
