//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// handAndFootNoArgCommands maps no-arg CUI commands to HandAndFoot interactor methods.
var handAndFootNoArgCommands = cuiutil.NewCommandMap[usecase.HandAndFootInteractorIF]().
	Add(usecase.HandAndFootInteractorIF.DrawFromStock, "ds", "drawstock").
	Add(usecase.HandAndFootInteractorIF.SkipMeld, "sm", "skipmeld").
	Add(usecase.HandAndFootInteractorIF.GoOut, "go", "goout").
	Add(usecase.HandAndFootInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.HandAndFootInteractorIF.Hint, "h", "hint").
	Add(usecase.HandAndFootInteractorIF.ActionLog, "log", "l")

// handAndFootArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var handAndFootArgfulCommands = []string{
	"dd", "drawdiscard", "m", "meld", "d", "discard",
	"sd", "setdifficulty", "sl", "setlimit",
}

// HandAndFootCuiController ハンドアンドフットCUIコントローラークラス
type HandAndFootCuiController struct {
	ci usecase.HandAndFootInteractorIF
}

// NewHandAndFootCuiController コンストラクタ
func NewHandAndFootCuiController(ci usecase.HandAndFootInteractorIF) *HandAndFootCuiController {
	return &HandAndFootCuiController{ci: ci}
}

// Exec コマンド実行
func (c *HandAndFootCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		append(handAndFootNoArgCommands.Names(), handAndFootArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := handAndFootNoArgCommands.Lookup(cmd); ok {
				return fn(c.ci), true
			}
			switch cmd {
			case "dd", "drawdiscard":
				indices := parseIntList(args)
				return c.ci.DrawFromDiscard(indices), true
			case "m", "meld":
				groups := parseMeldGroups(args)
				return c.ci.Meld(groups), true
			case "d", "discard":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.HandAndFootCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return "", false
			}
		},
	)
}
