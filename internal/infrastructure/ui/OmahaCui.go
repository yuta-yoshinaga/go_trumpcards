package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OmahaCui オマハホールデムCUIクラス
type OmahaCui struct {
	oc *controller.OmahaCuiController
}

// NewOmahaCui コンストラクタ
func NewOmahaCui() *OmahaCui {
	cfg := domain.DefaultOmahaConfig()
	omaha := domain.NewOmaha(domain.NewTrumpCards(0), domain.NewOmahaPlayersForTable(cfg.TableSize), cfg)
	return &OmahaCui{
		oc: controller.NewOmahaCuiController(usecase.NewOmahaInteractor(omaha, new(presenter.OmahaCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *OmahaCui) Controller() CuiExecer { return cui.oc }

// HelpLines returns the game's help lines.
func (cui *OmahaCui) HelpLines() []string {
	return []string{
		i18n.T("omaha.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("omaha.helpFold"),
		i18n.T("omaha.helpCheck"),
		i18n.T("omaha.helpCall"),
		i18n.T("omaha.helpBet"),
		i18n.T("omaha.helpRaise"),
		i18n.T("omaha.helpAllIn"),
		"  rb                   rebuy",
		"  sr                   skip rebuy",
		"  ad                   add-on",
		"  sa                   skip add-on",
		"",
		i18n.T("settings"),
		i18n.T("omaha.helpBettingLimit"),
		i18n.T("omaha.helpTournament"),
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
func (cui *OmahaCui) Exec() {
	RunCuiLoop(cui.oc, cui.HelpLines())
}
