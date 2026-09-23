//go:build !js || !wasm || extra4

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/i18n"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BassetCuiController handles Basset terminal commands.
type BassetCuiController struct{ ci usecase.BassetInteractorIF }

// NewBassetCuiController creates a Basset terminal controller.
func NewBassetCuiController(ci usecase.BassetInteractorIF) *BassetCuiController {
	return &BassetCuiController{ci: ci}
}

// Exec executes reset, bet, deal, take, paroli, next, log, or quit.
func (c *BassetCuiController) Exec(command string) string {
	return execCuiCommand(command, func(_ []string) string { return c.ci.Reset() }, []string{"b", "bet", "deal", "d", "take", "paroli", "next", "n", "log"}, func(cmd string, args []string) (string, bool) {
		switch cmd {
		case "b", "bet":
			if len(args) < 2 {
				return i18n.MarkError(i18n.T("basset.errBetArgs")), true
			}
			rank, e1 := strconv.Atoi(args[0])
			amount, e2 := strconv.Atoi(args[1])
			if e1 != nil || e2 != nil {
				return i18n.MarkError(i18n.T("basset.errBetArgs")), true
			}
			return c.ci.PlaceBet(rank, amount), true
		case "deal", "d":
			return c.ci.DealTurn(), true
		case "take":
			return c.ci.TakeWinnings(), true
		case "paroli":
			return c.ci.PressParoli(), true
		case "next", "n":
			return c.ci.NextRound(), true
		default:
			return handleCuiLog(cmd, c.ci.ActionLog)
		}
	})
}
