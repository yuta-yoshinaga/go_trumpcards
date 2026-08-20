//go:build !js || !wasm || classic

package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoudizhuCuiController 斗地主CUIコントローラークラス
type DoudizhuCuiController struct {
	dgi usecase.DoudizhuInteractorIF
}

// NewDoudizhuCuiController コンストラクタ
func NewDoudizhuCuiController(dgi usecase.DoudizhuInteractorIF) *DoudizhuCuiController {
	return &DoudizhuCuiController{dgi: dgi}
}

// Exec CUIコマンドを実行する
func (c *DoudizhuCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.dgi.GetConfig()
			return c.dgi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "b", "bid", "sd", "setdifficulty",
			"log", "l",
		},
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
				return c.dgi.Play(indices), true
			case "b", "bid":
				if len(args) == 0 {
					return c.dgi.Bid(0), true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > domain.DoudizhuMaxBid {
					return invalidArg("invalidBidValue03Pass"), true
				}
				return c.dgi.Bid(v), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequiredAlt", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.dgi.GetConfig()
					cfg.CpuDifficulty = domain.DoudizhuCpuDifficulty(v)
					return c.dgi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.dgi.ActionLog)
			}
		},
	)
}
