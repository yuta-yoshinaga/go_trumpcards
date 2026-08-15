package controller

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ShitheadCuiController シットヘッドCUIコントローラークラス
type ShitheadCuiController struct {
	si usecase.ShitheadInteractorIF
}

// NewShitheadCuiController コンストラクタ
func NewShitheadCuiController(si usecase.ShitheadInteractorIF) *ShitheadCuiController {
	return &ShitheadCuiController{si: si}
}

// shitheadRuleKeys sr コマンドで操作できるルール一覧
var shitheadRuleKeys = []string{"two", "seven", "eight", "ten", "fourburn"}

// getShitheadRule returns the current value of a boolean rule flag.
func getShitheadRule(cfg *domain.ShitheadConfig, key string) (value bool, ok bool) {
	switch key {
	case "two":
		return cfg.MagicTwo, true
	case "seven":
		return cfg.MagicSeven, true
	case "eight":
		return cfg.MagicEight, true
	case "ten":
		return cfg.MagicTen, true
	case "fourburn":
		return cfg.FourOfAKindBurn, true
	default:
		return false, false
	}
}

// formatShitheadRuleList returns the rules with their current ON/OFF status.
func formatShitheadRuleList(cfg *domain.ShitheadConfig) string {
	var b strings.Builder
	for _, key := range shitheadRuleKeys {
		val, _ := getShitheadRule(cfg, key)
		status := "OFF"
		if val {
			status = "ON"
		}
		fmt.Fprintf(&b, "  %-10s %s\n", key, status)
	}
	return b.String()
}

// setShitheadRule toggles a rule on/off. Returns false if the key is unknown.
func setShitheadRule(cfg *domain.ShitheadConfig, key string, value bool) bool {
	switch key {
	case "two":
		if cfg != nil {
			cfg.MagicTwo = value
		}
		return true
	case "seven":
		if cfg != nil {
			cfg.MagicSeven = value
		}
		return true
	case "eight":
		if cfg != nil {
			cfg.MagicEight = value
		}
		return true
	case "ten":
		if cfg != nil {
			cfg.MagicTen = value
		}
		return true
	case "fourburn":
		if cfg != nil {
			cfg.FourOfAKindBurn = value
		}
		return true
	default:
		return false
	}
}

// Exec コマンド実行。"p [indices...]" でカードを出す ("p" 単体はピックアップ)。
func (c *ShitheadCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "sd", "setdifficulty", "sr", "setrule", "log", "l",
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
				return c.si.Play(indices), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.CpuDifficulty = domain.ShitheadCpuDifficulty(v)
					return c.si.ResetWithConfig(cfg)
				})
			case "sr", "setrule":
				if len(args) >= 1 && args[0] == "list" {
					cfg := c.si.GetConfig()
					return "Rules:\n" + formatShitheadRuleList(&cfg), true
				}
				if len(args) < 2 {
					return "Usage: sr <rule> <0|1> | sr list\nRules: " + strings.Join(shitheadRuleKeys, ", "), true
				}
				if !setShitheadRule(nil, args[0], false) {
					return fmt.Sprintf("Unknown rule: %s.", args[0]), true
				}
				if args[1] != "0" && args[1] != "1" {
					return fmt.Sprintf("Invalid value: %s. Please enter 0 or 1.", args[1]), true
				}
				cfg := c.si.GetConfig()
				setShitheadRule(&cfg, args[0], args[1] == "1")
				return c.si.ResetWithConfig(cfg), true
			default:
				return handleCuiLog(cmd, c.si.ActionLog)
			}
		},
	)
}
