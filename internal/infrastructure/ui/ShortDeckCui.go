package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShortDeckCui ショートデックCUIクラス
type ShortDeckCui struct {
	sc *controller.ShortDeckCuiController
}

// NewShortDeckCui コンストラクタ
func NewShortDeckCui() *ShortDeckCui {
	cfg := domain.DefaultShortDeckConfig()
	sd := domain.NewShortDeck(domain.NewTrumpCardsShortDeck(), domain.NewShortDeckPlayersForTable(cfg.TableSize), cfg)
	return &ShortDeckCui{
		sc: controller.NewShortDeckCuiController(usecase.NewShortDeckInteractor(sd, new(presenter.ShortDeckCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *ShortDeckCui) Controller() CuiExecer { return cui.sc }

// HelpLines returns the game's help lines.
func (cui *ShortDeckCui) HelpLines() []string {
	return []string{
		i18n.T("shortdeck.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("shortdeck.helpFold"),
		i18n.T("shortdeck.helpCheck"),
		i18n.T("shortdeck.helpCall"),
		i18n.T("shortdeck.helpBet"),
		i18n.T("shortdeck.helpRaise"),
		i18n.T("shortdeck.helpAllIn"),
		"  rb                   rebuy",
		"  sr                   skip rebuy",
		"  ad                   add-on",
		"  sa                   skip add-on",
		"",
		i18n.T("settings"),
		i18n.T("shortdeck.helpBettingLimit"),
		i18n.T("shortdeck.helpTournament"),
		"  sb <amount>          small blind (>=1)",
		"  bb <amount>          big blind (>=2)",
		"  lh <hands>           blind level-up hands (>=1)",
		"  ts [4|6|9]           table size",
		"",
		i18n.T("session"),
		i18n.T("resetEntry"),
		i18n.T("quitEntry"),
		i18n.T("helpEntry"),
	}
}

// Exec ゲーム実行
func (cui *ShortDeckCui) Exec() {
	RunCuiLoop(cui.sc, cui.HelpLines())
}
