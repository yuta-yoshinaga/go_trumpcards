//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// biribaNoArgCommands maps no-arg CUI commands to Biriba interactor methods.
var biribaNoArgCommands = cuiutil.NewCommandMap[usecase.BiribaInteractorIF]().
	Add(usecase.BiribaInteractorIF.DrawFromStock, "ds", "drawstock").
	Add(usecase.BiribaInteractorIF.SkipMeld, "sm", "skipmeld").
	Add(usecase.BiribaInteractorIF.GoOut, "go", "goout").
	Add(usecase.BiribaInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.BiribaInteractorIF.Hint, "h", "hint").
	Add(usecase.BiribaInteractorIF.ActionLog, "log", "l")

// biribaArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var biribaArgfulCommands = []string{
	"dd", "drawdiscard", "m", "meld", "d", "discard",
	"sd", "setdifficulty", "sl", "setlimit",
}

// BiribaCuiController ビリバCUIコントローラークラス
type BiribaCuiController struct {
	ci usecase.BiribaInteractorIF
}

// NewBiribaCuiController コンストラクタ
func NewBiribaCuiController(ci usecase.BiribaInteractorIF) *BiribaCuiController {
	return &BiribaCuiController{ci: ci}
}

// Exec コマンド実行
func (c *BiribaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		append(biribaNoArgCommands.Names(), biribaArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := biribaNoArgCommands.Lookup(cmd); ok {
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
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.BiribaCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit", 1, math.MaxInt, func(v int) string {
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
