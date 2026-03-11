package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HoldemCui テキサスホールデムCUIクラス
type HoldemCui struct {
	hc *controller.HoldemCuiController
}

// NewHoldemCui コンストラクタ
func NewHoldemCui() *HoldemCui {
	cfg := domain.DefaultHoldemConfig()
	holdem := domain.NewHoldem(domain.NewTrumpCards(0), domain.NewPlayersForTable(cfg.TableSize), cfg)
	return &HoldemCui{
		hc: controller.NewHoldemCuiController(usecase.NewHoldemInteractor(holdem, new(presenter.HoldemCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *HoldemCui) Controller() CuiExecer { return cui.hc }

// HelpLines returns the game's help lines.
func (cui *HoldemCui) HelpLines() []string {
	return []string{
		i18n.T("holdem.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("holdem.helpFold"),
		i18n.T("holdem.helpCheck"),
		i18n.T("holdem.helpCall"),
		i18n.T("holdem.helpBet"),
		i18n.T("holdem.helpRaise"),
		i18n.T("holdem.helpAllIn"),
		"  rb                   rebuy",
		"  sr                   skip rebuy",
		"  ad                   add-on",
		"  sa                   skip add-on",
		"",
		i18n.T("settings"),
		i18n.T("holdem.helpBettingLimit"),
		i18n.T("holdem.helpTournament"),
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
func (cui *HoldemCui) Exec() {
	RunCuiLoop(cui.hc, cui.HelpLines())
}
