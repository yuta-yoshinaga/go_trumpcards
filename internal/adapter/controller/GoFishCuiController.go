package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GoFishCuiController Go FishCUIコントローラークラス
type GoFishCuiController struct {
	gi usecase.GoFishInteractorIF
}

// NewGoFishCuiController コンストラクタ
func NewGoFishCuiController(gi usecase.GoFishInteractorIF) *GoFishCuiController {
	return &GoFishCuiController{gi: gi}
}

// Exec コマンド実行
// ask コマンドは "ask <targetIdx> <rank>" の形式。
// sd コマンドで CPU 難易度を変更 (0=Easy, 1=Normal, 2=Hard)。
func (c *GoFishCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.gi.Reset(c.gi.GetConfig()) },
		[]string{
			"ask", "sd", "setdifficulty",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ask":
				if len(args) < 2 {
					return invalidArg("usageAskTargetidxRank"), true
				}
				targetIdx, err1 := strconv.Atoi(args[0])
				rank, err2 := strconv.Atoi(args[1])
				if err1 != nil {
					return invalidArg("invalidTargetIndexRaw", "val", fmt.Sprint(args[0])), true
				}
				if err2 != nil {
					return invalidArg("invalidRankRaw", "val", fmt.Sprint(args[1])), true
				}
				return c.gi.Ask(targetIdx, rank), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.gi.GetConfig()
					cfg.CpuDifficulty = domain.GoFishCpuDifficulty(v)
					return c.gi.Reset(cfg)
				})
			default:
				return handleCuiLog(cmd, c.gi.ActionLog)
			}
		},
	)
}
