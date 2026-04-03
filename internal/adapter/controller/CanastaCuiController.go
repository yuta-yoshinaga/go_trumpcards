package controller

import (
	"math"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CanastaCuiController カナスタCUIコントローラークラス
type CanastaCuiController struct {
	ci usecase.CanastaInteractorIF
}

// NewCanastaCuiController コンストラクタ
func NewCanastaCuiController(ci usecase.CanastaInteractorIF) *CanastaCuiController {
	return &CanastaCuiController{ci: ci}
}

// Exec コマンド実行
func (c *CanastaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard",
			"m", "meld", "sm", "skipmeld",
			"d", "discard", "go", "goout",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				indices := parseIntList(args)
				return c.ci.DrawFromDiscard(indices), true
			case "m", "meld":
				groups := parseMeldGroups(args)
				return c.ci.Meld(groups), true
			case "sm", "skipmeld":
				return c.ci.SkipMeld(), true
			case "d", "discard":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "go", "goout":
				return c.ci.GoOut(), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CanastaCpuDifficulty(v)
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

// parseMeldGroups メルドグループを解析する ("0,1,2;3,4,5" → [[0,1,2],[3,4,5]])
func parseMeldGroups(args []string) [][]int {
	joined := strings.Join(args, " ")
	if strings.TrimSpace(joined) == "" {
		return nil
	}

	// ";" or " -- " をグループセパレータとして使用
	joined = strings.ReplaceAll(joined, " -- ", ";")
	parts := strings.Split(joined, ";")

	var groups [][]int
	for _, part := range parts {
		indices := parseIntList(strings.Fields(strings.TrimSpace(part)))
		if len(indices) > 0 {
			groups = append(groups, indices)
		}
	}
	return groups
}
