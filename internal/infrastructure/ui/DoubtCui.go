package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoubtCui ダウトCUIクラス
type DoubtCui struct {
	dc *controller.DoubtCuiController
}

// NewDoubtCui コンストラクタ
func NewDoubtCui() *DoubtCui {
	game := domain.NewDefaultDoubt()
	dc := controller.NewDoubtCuiController(
		usecase.NewDoubtInteractor(game, new(presenter.DoubtCuiPresenter)),
	)
	return &DoubtCui{dc: dc}
}

// Controller returns the game controller.
func (cui *DoubtCui) Controller() CuiExecer { return cui.dc }

// HelpLines returns the game's help lines.
func (cui *DoubtCui) HelpLines() []string {
	return BuildCuiHelp(CuiHelpSpec{
		TitleKey:    "doubt.helpTitle",
		CommandKeys: []string{"doubt.helpPlay", "doubt.helpDoubt", "doubt.helpSkip", "doubt.helpLog"},
		SettingKeys: []string{"doubt.helpSetMemory", "doubt.helpSetPenalty", "doubt.helpSetHesitation"},
	})
}
