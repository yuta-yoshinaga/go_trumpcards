package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// pageOneNoArgCommands maps no-arg CUI commands to PageOne interactor methods.
var pageOneNoArgCommands = cuiutil.NewCommandMap[usecase.PageOneInteractorIF]().
	Add(usecase.PageOneInteractorIF.Draw, "d", "draw").
	Add(usecase.PageOneInteractorIF.Declare, "dc", "declare").
	Add(usecase.PageOneInteractorIF.SkipDeclare, "sk", "skip").
	Add(usecase.PageOneInteractorIF.NextRound, "nr", "nextround").
	Add(usecase.PageOneInteractorIF.Hint, "h", "hint").
	Add(usecase.PageOneInteractorIF.ActionLog, "log", "l")

// pageOneArgfulCommands lists alias names for argful commands handled in the
// Exec switch.
var pageOneArgfulCommands = []string{
	"p", "play",
	"sd", "setdifficulty", "sl", "setlimit",
}

// PageOneCuiController ページワンCUIコントローラークラス
type PageOneCuiController struct {
	ci usecase.PageOneInteractorIF
}

// NewPageOneCuiController コンストラクタ
func NewPageOneCuiController(ci usecase.PageOneInteractorIF) *PageOneCuiController {
	return &PageOneCuiController{ci: ci}
}

// Exec コマンド実行
func (c *PageOneCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		append(pageOneNoArgCommands.Names(), pageOneArgfulCommands...),
		func(cmd string, args []string) (string, bool) {
			if fn, ok := pageOneNoArgCommands.Lookup(cmd); ok {
				return fn(c.ci), true
			}
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Play)
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.PageOneCpuDifficulty(v)
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
