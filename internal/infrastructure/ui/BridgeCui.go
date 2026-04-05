package ui

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/presenter"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NewBridgeCui コンストラクタ
func NewBridgeCui() *genericCuiGame {
	config := domain.DefaultBridgeConfig()
	players := []*domain.BridgePlayer{
		domain.NewBridgePlayer(true, 0),  // North (human, team 0)
		domain.NewBridgePlayer(false, 1), // East (CPU, team 1)
		domain.NewBridgePlayer(false, 0), // South (CPU, team 0)
		domain.NewBridgePlayer(false, 1), // West (CPU, team 1)
	}
	bridge := domain.NewBridge(domain.NewTrumpCards(0), players, config)
	bc := controller.NewBridgeCuiController(usecase.NewBridgeInteractor(bridge, new(presenter.BridgeCuiPresenter)))
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
