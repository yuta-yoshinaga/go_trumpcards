//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FourCardPokerCuiController is the Four Card Poker CUI controller.
type FourCardPokerCuiController struct {
	ti usecase.FourCardPokerInteractorIF
}

// NewFourCardPokerCuiController constructs the controller.
func NewFourCardPokerCuiController(ti usecase.FourCardPokerInteractorIF) *FourCardPokerCuiController {
	return &FourCardPokerCuiController{ti: ti}
}

// Exec runs a single CUI command.
// Commands: "r"/"reset", "b <ante> [acesUp]"/"bet ...", "p <mul>"/"play <mul>",
// "f"/"fold", "log"/"l", "q"/"quit".
func (c *FourCardPokerCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ti.Reset() },
		[]string{"b", "bet", "p", "play", "f", "fold", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				ante, errMsg, ok := cuiutil.ParseIntArgKeys(args, "anteAmountRequired", "invalidAnteAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				acesUp := 0
				if len(args) > 1 {
					acesUp, errMsg, ok = cuiutil.ParseIntArgKeys(args[1:], "", "invalidAcesUpAmount", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
				}
				return c.ti.Bet(ante, acesUp), true
			case "p", "play":
				// Default to 1x if not specified.
				mul := 1
				if len(args) >= 1 {
					m, errMsg, ok := cuiutil.ParseIntArgKeys(args, "", "invalidPlayMultiplier12Or3", 1, 3)
					if !ok {
						return errMsg, true
					}
					mul = m
				}
				return c.ti.Play(mul), true
			case "f", "fold":
				return c.ti.Fold(), true
			default:
				return handleCuiLog(cmd, c.ti.ActionLog)
			}
		},
	)
}
