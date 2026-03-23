package controller

import (
	"math"
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// GinRummyCuiController ジンラミーCUIコントローラークラス
type GinRummyCuiController struct {
	ci usecase.GinRummyInteractorIF
}

// NewGinRummyCuiController コンストラクタ
func NewGinRummyCuiController(ci usecase.GinRummyInteractorIF) *GinRummyCuiController {
	return &GinRummyCuiController{ci: ci}
}

// Exec コマンド実行
func (c *GinRummyCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"ds", "drawstock", "dd", "drawdiscard", "d", "discard",
			"k", "knock", "lo", "layoff",
			"nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "ds", "drawstock":
				return c.ci.DrawFromStock(), true
			case "dd", "drawdiscard":
				return c.ci.DrawFromDiscard(), true
			case "d", "discard":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "k", "knock":
				return cuiutil.WithParsedInt(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax, c.ci.Knock)
			case "lo", "layoff":
				indices := parseIntList(args)
				return c.ci.Layoff(indices), true
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.GinRummyCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedInt(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.PointLimit = v
					return c.ci.ResetWithConfig(cfg)
				})
			case "log", "l":
				return c.ci.ActionLog(), true
			}
			return "", false
		},
	)
}

// parseIntList 引数をintスライスに変換する (空→nil)
func parseIntList(args []string) []int {
	if len(args) == 0 {
		return nil
	}
	var result []int
	for _, a := range args {
		for _, s := range strings.Split(a, ",") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			v, err := strconv.Atoi(s)
			if err == nil {
				result = append(result, v)
			}
		}
	}
	return result
}
