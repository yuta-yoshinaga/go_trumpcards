package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewDurakCui コンストラクタ
func NewDurakCui() *genericCuiGame {
	dc := controller.NewDurakCuiController(
		usecase.NewDurakInteractor(domain.NewDefaultDurak(), new(presenter.DurakCuiPresenter)),
	)
	return newCuiGame(dc, []string{
		i18n.T("durak.helpTitle"),
		"",
		i18n.T("gameCommands"),
		"  a <idx>                  attack with card",
		"  d <atkIdx> <handIdx>     defend attack card",
		"  p                        pass (stop attacking)",
		"  t                        take cards (give up defense)",
		"  sort <0|1>               sort hand (0=suit, 1=value)",
		"  sd <0-2>                 set CPU difficulty",
		"  l                        action log",
		"",
		i18n.T("commonCommands"),
	})
}
