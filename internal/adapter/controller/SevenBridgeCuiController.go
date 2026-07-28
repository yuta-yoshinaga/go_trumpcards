//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SevenBridgeCuiController セブンブリッジ CUI コントローラー
type SevenBridgeCuiController struct {
	ci usecase.SevenBridgeInteractorIF
}

// NewSevenBridgeCuiController コンストラクタ
func NewSevenBridgeCuiController(ci usecase.SevenBridgeInteractorIF) *SevenBridgeCuiController {
	return &SevenBridgeCuiController{ci: ci}
}

// Exec コマンドを実行する
func (c *SevenBridgeCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock",
			"pon", "chi",
			"m", "meld",
			"lo", "layoff",
			"d", "discard",
			"nr", "nextround",
			"h", "hint",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "pon":
				return c.ci.ClaimPon(parseIntList(args)), true
			case "chi":
				return c.ci.ClaimChi(parseIntList(args)), true
			case "m", "meld":
				return c.ci.Meld(parseIntList(args)), true
			case "lo", "layoff":
				idx := parseIntList(args)
				if len(idx) < 3 {
					return "Usage: lo <targetPlayerIdx> <meldIdx> <cardIndex>", true
				}
				return c.ci.Layoff(idx[0], idx[1], idx[2]), true
			case "d", "discard":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "h", "hint":
				return c.ci.Hint(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.SevenBridgeCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}
