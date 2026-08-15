//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ChinchonCuiController チンチョンCUIコントローラークラス
type ChinchonCuiController struct {
	ci usecase.ChinchonInteractorIF
}

// NewChinchonCuiController コンストラクタ
func NewChinchonCuiController(ci usecase.ChinchonInteractorIF) *ChinchonCuiController {
	return &ChinchonCuiController{ci: ci}
}

// Exec コマンド実行
//
// レイオフコマンド (lo/layoff) はスペースまたはカンマ区切りで手札インデックスを複数指定する
// (引数なしでレイオフ終了)。共有ヘルパー parseIntList を利用する。
func (c *ChinchonCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard",
			"d", "discard", "k", "knock", "lo", "layoff",
			"nr", "nextround",
			"sd", "setdifficulty", "sp", "setplayers", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "k", "knock":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Knock)
			case "lo", "layoff":
				return c.ci.Layoff(parseIntList(args)), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.ChinchonCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sp", "setplayers":
				return cuiutil.WithParsedInt(args, "Player count is required (2-4).", "Invalid player count: %s. Please enter 2-4.", 2, 4, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PlayerCount = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
