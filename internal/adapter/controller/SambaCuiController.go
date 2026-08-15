//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// sambaNoArgCommands maps no-arg CUI commands to Samba interactor methods.
var sambaNoArgCommands = cuiutil.NewCommandMap[usecase.SambaInteractorIF]().
	Add(usecase.SambaInteractorIF.DrawFromStock, "ds", "drawstock").
	Add(usecase.SambaInteractorIF.SkipMeld, "sm", "skipmeld").
	Add(usecase.SambaInteractorIF.GoOut, "go", "goout").
	Add(usecase.SambaInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.SambaInteractorIF.ActionLog, "log", "l")

// sambaArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var sambaArgfulCommands = []string{
	"dd", "drawdiscard", "m", "meld", "d", "discard",
	"sd", "setdifficulty", "sl", "setlimit",
}

// SambaCuiController サンバCUIコントローラークラス
type SambaCuiController struct {
	ci usecase.SambaInteractorIF
}

// NewSambaCuiController コンストラクタ
func NewSambaCuiController(ci usecase.SambaInteractorIF) *SambaCuiController {
	return &SambaCuiController{ci: ci}
}

// Exec コマンド実行
func (c *SambaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		append(sambaNoArgCommands.Names(), sambaArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := sambaNoArgCommands.Lookup(cmd); ok {
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
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.SambaCpuDifficulty(v)
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
