package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpiderCui スパイダーソリティアCUIクラス
type SpiderCui struct {
	sc *controller.SpiderCuiController
}

// NewSpiderCui コンストラクタ
func NewSpiderCui() *SpiderCui {
	spider := domain.NewSpider(domain.NewTrumpCardsWithSuits(domain.SpiderTotalCards, []int{domain.CardDesignSpade}))
	return &SpiderCui{
		sc: controller.NewSpiderCuiController(usecase.NewSpiderInteractor(spider, new(presenter.SpiderCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *SpiderCui) Controller() CuiExecer { return cui.sc }

// HelpLines returns the game's help lines.
func (cui *SpiderCui) HelpLines() []string {
	return []string{
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
	}
}

// Exec ゲーム実行
func (cui *SpiderCui) Exec() {
	RunCuiLoop(cui.sc, cui.HelpLines())
}
