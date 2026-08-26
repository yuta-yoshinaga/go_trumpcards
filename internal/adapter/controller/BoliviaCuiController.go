//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// boliviaNoArgCommands maps no-arg CUI commands to Bolivia interactor methods.
var boliviaNoArgCommands = cuiutil.NewCommandMap[usecase.BoliviaInteractorIF]().
	Add(usecase.BoliviaInteractorIF.DrawFromStock, "ds", "drawstock").
	Add(usecase.BoliviaInteractorIF.SkipMeld, "sm", "skipmeld").
	Add(usecase.BoliviaInteractorIF.GoOut, "go", "goout").
	Add(usecase.BoliviaInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.BoliviaInteractorIF.ActionLog, "log", "l")

// boliviaArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var boliviaArgfulCommands = []string{
	"dd", "drawdiscard", "m", "meld", "d", "discard",
	"sd", "setdifficulty", "sl", "setlimit",
}

// BoliviaCuiController ボリビアCUIコントローラークラス
type BoliviaCuiController struct {
	ci usecase.BoliviaInteractorIF
}

// NewBoliviaCuiController コンストラクタ
func NewBoliviaCuiController(ci usecase.BoliviaInteractorIF) *BoliviaCuiController {
	return &BoliviaCuiController{ci: ci}
}

// Exec コマンド実行
func (c *BoliviaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		append(boliviaNoArgCommands.Names(), boliviaArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := boliviaNoArgCommands.Lookup(cmd); ok {
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
					cfg.CpuDifficulty = domain.BoliviaCpuDifficulty(v)
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
