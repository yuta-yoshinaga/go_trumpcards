package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PineappleCui パイナップルポーカーCUIクラス
type PineappleCui struct {
	pc *controller.PineappleCuiController
}

// NewPineappleCui コンストラクタ
func NewPineappleCui() *PineappleCui {
	cfg := domain.DefaultPineappleConfig()
	pineapple := domain.NewPineapple(domain.NewTrumpCards(0), domain.NewPineapplePlayersForTable(cfg.TableSize), cfg)
	return &PineappleCui{
		pc: controller.NewPineappleCuiController(usecase.NewPineappleInteractor(pineapple, new(presenter.PineappleCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *PineappleCui) Controller() CuiExecer { return cui.pc }

// HelpLines returns the game's help lines.
func (cui *PineappleCui) HelpLines() []string {
	return []string{
		i18n.T("pineapple.helpTitle"),
		"",
		i18n.T("gameCommands"),
		i18n.T("pineapple.helpFold"),
		i18n.T("pineapple.helpCheck"),
		i18n.T("pineapple.helpCall"),
		i18n.T("pineapple.helpBet"),
		i18n.T("pineapple.helpRaise"),
		i18n.T("pineapple.helpAllIn"),
		"  d <index>            discard",
		"  rb                   rebuy",
		"  sr                   skip rebuy",
		"  ad                   add-on",
		"  sa                   skip add-on",
		"",
		i18n.T("settings"),
		i18n.T("pineapple.helpBettingLimit"),
		i18n.T("pineapple.helpTournament"),
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
func (cui *PineappleCui) Exec() {
	RunCuiLoop(cui.pc, cui.HelpLines())
}
