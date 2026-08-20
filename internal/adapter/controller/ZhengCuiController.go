//go:build !js || !wasm || solo

package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ZhengCuiController 争上游 CUIコントローラークラス
type ZhengCuiController struct {
	zi usecase.ZhengInteractorIF
}

// NewZhengCuiController コンストラクタ
func NewZhengCuiController(zi usecase.ZhengInteractorIF) *ZhengCuiController {
	return &ZhengCuiController{zi: zi}
}

// Exec コマンド実行
func (c *ZhengCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.zi.GetConfig()
			return c.zi.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "sd", "setdifficulty", "log", "l"},
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
				return c.zi.Play(indices), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.zi.GetConfig()
					cfg.CpuDifficulty = domain.ZhengCpuDifficulty(v)
					return c.zi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.zi.ActionLog)
			}
		},
	)
}
