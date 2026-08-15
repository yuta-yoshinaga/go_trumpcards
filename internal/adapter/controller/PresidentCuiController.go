package controller

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// PresidentCuiController プレジデントCUIコントローラークラス
type PresidentCuiController struct {
	pi usecase.PresidentInteractorIF
}

// NewPresidentCuiController コンストラクタ
func NewPresidentCuiController(pi usecase.PresidentInteractorIF) *PresidentCuiController {
	return &PresidentCuiController{pi: pi}
}

// presidentRuleKeys sr コマンドで操作できるルール一覧
var presidentRuleKeys = []string{"revolution", "exchange", "passflush"}

// getPresidentRule returns the current value of a boolean rule flag.
func getPresidentRule(cfg *domain.PresidentConfig, key string) (value bool, ok bool) {
	switch key {
	case "revolution":
		return cfg.RevolutionEnabled, true
	case "exchange":
		return cfg.CardExchangeEnabled, true
	case "passflush":
		return cfg.PassFieldFlushEnabled, true
	default:
		return false, false
	}
}

// formatPresidentRuleList returns the rules with their current ON/OFF status.
func formatPresidentRuleList(cfg *domain.PresidentConfig) string {
	var b strings.Builder
	for _, key := range presidentRuleKeys {
		val, _ := getPresidentRule(cfg, key)
		status := "OFF"
		if val {
			status = "ON"
		}
		fmt.Fprintf(&b, "  %-12s %s\n", key, status)
	}
	return b.String()
}

// setPresidentRule toggles a rule on/off. Returns false if the key is unknown.
func setPresidentRule(cfg *domain.PresidentConfig, key string, value bool) bool {
	switch key {
	case "revolution":
		if cfg != nil {
			cfg.RevolutionEnabled = value
		}
		return true
	case "exchange":
		if cfg != nil {
			cfg.CardExchangeEnabled = value
		}
		return true
	case "passflush":
		if cfg != nil {
			cfg.PassFieldFlushEnabled = value
		}
		return true
	default:
		return false
	}
}

// Exec コマンド実行
// play コマンドは "p 0 2" または "play 0 2" の形式でカードインデックスを指定。
// インデックスなしの場合はパス扱い。例: "p" → パス / "p 0 2" → 0番と2番のカードを出す
func (c *PresidentCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.pi.GetConfig()
			return c.pi.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "h", "hint", "sd", "setdifficulty", "sr", "setrule", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				indices, skipped := cuiutil.ParseIntSlice(args)
				return cuiutil.PrependSkippedWarning(c.pi.Play(indices), skipped), true
			case "h", "hint":
				return c.pi.Hint(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.pi.GetConfig()
					cfg.CpuDifficulty = domain.PresidentCpuDifficulty(v)
					return c.pi.ResetWithConfig(cfg)
				})
			case "sr", "setrule":
				if len(args) >= 1 && args[0] == "list" {
					cfg := c.pi.GetConfig()
					return "Rules:\n" + formatPresidentRuleList(&cfg), true
				}
				if len(args) < 2 {
					return "Usage: sr <rule> <0|1> | sr list\nRules: " + strings.Join(presidentRuleKeys, ", "), true
				}
				if !setPresidentRule(nil, args[0], false) {
					return fmt.Sprintf("Unknown rule: %s.", args[0]), true
				}
				if args[1] != "0" && args[1] != "1" {
					return fmt.Sprintf("Invalid value: %s. Please enter 0 or 1.", args[1]), true
				}
				cfg := c.pi.GetConfig()
				setPresidentRule(&cfg, args[0], args[1] == "1")
				return c.pi.ResetWithConfig(cfg), true
			default:
				return handleCuiLog(cmd, c.pi.ActionLog)
			}
		},
	)
}
