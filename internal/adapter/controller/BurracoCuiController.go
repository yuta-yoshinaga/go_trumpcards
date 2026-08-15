//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// burracoNoArgCommands maps no-arg CUI commands to Burraco interactor methods.
var burracoNoArgCommands = cuiutil.NewCommandMap[usecase.BurracoInteractorIF]().
	Add(usecase.BurracoInteractorIF.DrawFromStock, "ds", "drawstock").
	Add(usecase.BurracoInteractorIF.SkipMeld, "sm", "skipmeld").
	Add(usecase.BurracoInteractorIF.GoOut, "go", "goout").
	Add(usecase.BurracoInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.BurracoInteractorIF.Hint, "h", "hint").
	Add(usecase.BurracoInteractorIF.ActionLog, "log", "l")

// burracoArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var burracoArgfulCommands = []string{
	"dd", "drawdiscard", "m", "meld", "d", "discard",
	"sd", "setdifficulty", "sl", "setlimit",
}

// BurracoCuiController ブラーコCUIコントローラークラス
type BurracoCuiController struct {
	ci usecase.BurracoInteractorIF
}

// NewBurracoCuiController コンストラクタ
func NewBurracoCuiController(ci usecase.BurracoInteractorIF) *BurracoCuiController {
	return &BurracoCuiController{ci: ci}
}

// Exec コマンド実行
func (c *BurracoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		append(burracoNoArgCommands.Names(), burracoArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := burracoNoArgCommands.Lookup(cmd); ok {
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
					cfg.CpuDifficulty = domain.BurracoCpuDifficulty(v)
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
