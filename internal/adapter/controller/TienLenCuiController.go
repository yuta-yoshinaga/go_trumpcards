//go:build !js || !wasm || solo

package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TienLenCuiController Tien Len CUIコントローラークラス
type TienLenCuiController struct {
	tli usecase.TienLenInteractorIF
}

// NewTienLenCuiController コンストラクタ
func NewTienLenCuiController(tli usecase.TienLenInteractorIF) *TienLenCuiController {
	return &TienLenCuiController{tli: tli}
}

// Exec コマンド実行
func (c *TienLenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.tli.GetConfig()
			return c.tli.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "sd", "setdifficulty", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices, skipped := cuiutil.ParseIntSlice(args)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return c.tli.Play(indices), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.tli.GetConfig()
					cfg.CpuDifficulty = domain.TienLenCpuDifficulty(v)
					return c.tli.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.tli.Hint, c.tli.ActionLog)
			}
		},
	)
}
