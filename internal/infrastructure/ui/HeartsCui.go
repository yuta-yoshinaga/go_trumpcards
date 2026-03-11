package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HeartsCui ハーツCUIクラス
type HeartsCui struct {
	hc *controller.HeartsCuiController
}

// NewHeartsCui コンストラクタ
func NewHeartsCui() *HeartsCui {
	config := domain.DefaultHeartsConfig()
	players := []*domain.HeartsPlayer{
		domain.NewHeartsPlayer(true),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
		domain.NewHeartsPlayer(false),
	}
	hearts := domain.NewHearts(domain.NewTrumpCards(0), players, config)
	return &HeartsCui{
		hc: controller.NewHeartsCuiController(usecase.NewHeartsInteractor(hearts, new(presenter.HeartsCuiPresenter))),
	}
}

// Controller returns the game controller.
func (cui *HeartsCui) Controller() CuiExecer { return cui.hc }

// HelpLines returns the game's help lines.
func (cui *HeartsCui) HelpLines() []string {
	return []string{
		"Please enter a command.",
		"q・・・quit",
		"r・・・reset",
		"pass <i1> <i2> <i3>・・・pass 3 cards",
		"p <i>・・・play card at index",
		"n・・・next trick",
		"nr・・・next round",
		"sd <0-2>・・・set CPU difficulty (0=Easy, 1=Normal, 2=Hard)",
		"sl <n>・・・set point limit",
		"l・・・action log",
	}
}

// Exec ゲーム実行
func (cui *HeartsCui) Exec() {
	RunCuiLoop(cui.hc, cui.HelpLines())
}
