//go:build !js || !wasm || extra2

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TehonbikiCuiController accepts bet <type> <numbers...> <chips>.
type TehonbikiCuiController struct{ ci usecase.TehonbikiInteractorIF }

func NewTehonbikiCuiController(ci usecase.TehonbikiInteractorIF) *TehonbikiCuiController {
	return &TehonbikiCuiController{ci}
}
func (c *TehonbikiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.ci.Reset() },
		[]string{"bet", "b", "next", "hint", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				if len(args) < 3 {
					_, errMsg, _ := cuiutil.ParseIntArgKeys(nil, "betRequired", "", 0, math.MaxInt)
					return errMsg, true
				}
				chips, errMsg, ok := cuiutil.ParseIntArgKeys(args[len(args)-1:], "betRequired", "invalidBetANumber", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				numbers := make([]int, len(args)-2)
				for i := range numbers {
					number, errMsg, ok := cuiutil.ParseIntArgKeys(args[i+1:i+2], "", "invalidIndexANumber", 1, math.MaxInt)
					if !ok {
						return errMsg, true
					}
					numbers[i] = number
				}
				return c.ci.PlaceBet(numbers, domain.TehonbikiBetType(args[0]), chips), true
			case "next":
				return c.ci.NextRound(), true
			case "hint":
				return c.ci.Hint(), true
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
