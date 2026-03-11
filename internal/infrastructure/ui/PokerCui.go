package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PokerCui ポーカーCUIクラス
type PokerCui struct {
	pc *controller.PokerCuiController
}

// NewPokerCui コンストラクタ
func NewPokerCui() *PokerCui {
	config := domain.DefaultPokerConfig()
	players := []*domain.PokerPlayer{
		domain.NewPokerPlayer(true, domain.PokerStyleBalanced),
		domain.NewPokerPlayer(false, domain.PokerStyleConservative),
		domain.NewPokerPlayer(false, domain.PokerStyleAggressive),
		domain.NewPokerPlayer(false, domain.PokerStyleBluffer),
	}
	poker := domain.NewPoker(domain.NewTrumpCards(config.JokerCount), players, config)
	return &PokerCui{
		pc: controller.NewPokerCuiController(usecase.NewPokerInteractor(poker, new(presenter.PokerCuiPresenter))),
	}
}

// Exec ゲーム実行
func (cui *PokerCui) Exec() {
	RunCuiLoop(cui.pc, []string{
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"b [amount]・・・bet (e.g. 'b 20')",
		"c・・・call",
		"ra [amount]・・・raise (e.g. 'ra 30')",
		"ck・・・check",
		"f・・・fold",
		"a・・・all-in",
		"e [0-4]・・・exchange (e.g. 'e 0 2 4' to exchange cards at index 0, 2, 4)",
		"s・・・stand (no exchange)",
		"bl [0-2]・・・betting limit (0=Fixed, 1=PotLimit, 2=NoLimit)",
		"lw・・・toggle 2-7 Lowball mode",
	})
}
