package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FiftyOneCuiController フィフティワンCUIコントローラークラス
type FiftyOneCuiController struct {
	fi usecase.FiftyOneInteractorIF
}

// NewFiftyOneCuiController コンストラクタ
func NewFiftyOneCuiController(fi usecase.FiftyOneInteractorIF) *FiftyOneCuiController {
	return &FiftyOneCuiController{fi: fi}
}

// Exec コマンド実行
func (c *FiftyOneCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.fi.Reset(c.fi.GetConfig()) },
		[]string{
			"p", "play", "a", "all", "stop",
			"sd", "setdifficulty",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				if len(args) < 2 {
					return "Usage: p <handIdx> <tableIdx>", true
				}
				handIdx, err1 := strconv.Atoi(args[0])
				tableIdx, err2 := strconv.Atoi(args[1])
				if err1 != nil {
					return invalidArg("invalidHandIndexRaw", "val", fmt.Sprint(args[0])), true
				}
				if err2 != nil {
					return invalidArg("invalidTableIndexRaw", "val", fmt.Sprint(args[1])), true
				}
				return c.fi.ExchangeOne(handIdx, tableIdx), true
			case "a", "all":
				return c.fi.ExchangeAll(), true
			case "stop":
				return c.fi.Stop(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.fi.GetConfig()
					cfg.CpuDifficulty = domain.FiftyOneCpuDifficulty(v)
					return c.fi.Reset(cfg)
				})
			default:
				return handleCuiLog(cmd, c.fi.ActionLog)
			}
		},
	)
}
