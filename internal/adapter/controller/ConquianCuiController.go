//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ConquianCuiController コンキャンCUIコントローラークラス
type ConquianCuiController struct {
	ci usecase.ConquianInteractorIF
}

// NewConquianCuiController コンストラクタ
func NewConquianCuiController(ci usecase.ConquianInteractorIF) *ConquianCuiController {
	return &ConquianCuiController{ci: ci}
}

// Exec コマンド実行
//
// メルドコマンド (m/meld) はグループを ';' で区切り、各グループ内のインデックスは
// スペースまたはカンマ区切りで指定する (例: "m 0,1,2;3" → [[0,1,2],[3]])。
// 共有ヘルパー parseMeldGroups (Canasta と共通) を利用する。
func (c *ConquianCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard",
			"m", "meld", "d", "discard",
			"nr", "nextround",
			"sd", "setdifficulty", "sw", "setwins", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "m", "meld":
				return c.ci.Meld(parseMeldGroups(args)), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.ConquianCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sw", "setwins":
				return cuiutil.WithParsedInt(args, "Target wins is required.", "Invalid target wins: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.TargetWins = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
