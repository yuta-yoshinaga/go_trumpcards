package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewBridgeCui コンストラクタ
func NewBridgeCui() *genericCuiGame {
	bc := controller.NewBridgeCuiController(usecase.NewBridgeInteractor(domain.NewDefaultBridge(), new(presenter.BridgeCuiPresenter)))
	return newCuiGame(bc, []string{
		"=== Contract Bridge ===",
		"",
		"Game Commands:",
		"  b <type> <level> <suit>  bid (type: 0=pass,1=bid,2=dbl,3=rdbl; level: 1-7; suit: 1-5)",
		"  p <index>                play a card",
		"  n                        next trick",
		"  nr                       next round (score & proceed)",
		"  h                        hint",
		"  l                        action log",
		"",
		"Settings:",
		"  sd <0-2>                 set CPU difficulty (0=Easy,1=Normal,2=Hard)",
		"",
		"Session:",
		"  r                        reset game",
		"  q                        quit",
		"  help                     show this help",
	})
}
