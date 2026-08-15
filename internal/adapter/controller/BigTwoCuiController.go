package controller

import (
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BigTwoCuiController Big Two CUIコントローラークラス
type BigTwoCuiController struct {
	bti usecase.BigTwoInteractorIF
}

// NewBigTwoCuiController コンストラクタ
func NewBigTwoCuiController(bti usecase.BigTwoInteractorIF) *BigTwoCuiController {
	return &BigTwoCuiController{bti: bti}
}

// Exec コマンド実行
func (c *BigTwoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.bti.GetConfig()
			return c.bti.ResetWithConfig(cfg)
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
				return c.bti.Play(indices), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.bti.GetConfig()
					cfg.CpuDifficulty = domain.BigTwoCpuDifficulty(v)
					return c.bti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.bti.ActionLog)
			}
		},
	)
}
