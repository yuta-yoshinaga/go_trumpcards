package controller

import (
	"fmt"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CassinoCuiController カシノ CUI コントローラークラス。
type CassinoCuiController struct {
	ci usecase.CassinoInteractorIF
}

// NewCassinoCuiController コンストラクタ。
func NewCassinoCuiController(ci usecase.CassinoInteractorIF) *CassinoCuiController {
	return &CassinoCuiController{ci: ci}
}

// cassinoRuleKeys sr コマンドで操作できるルール一覧。
var cassinoRuleKeys = []string{"multibuild", "sweepbonus"}

// getCassinoRule returns the current value of a boolean rule flag.
func getCassinoRule(cfg *domain.CassinoConfig, key string) (value bool, ok bool) {
	switch key {
	case "multibuild":
		return cfg.MultiBuildEnabled, true
	case "sweepbonus":
		return cfg.SweepBonusEnabled, true
	default:
		return false, false
	}
}

// formatCassinoRuleList returns the rules with their current ON/OFF status.
func formatCassinoRuleList(cfg *domain.CassinoConfig) string {
	var b strings.Builder
	for _, key := range cassinoRuleKeys {
		val, _ := getCassinoRule(cfg, key)
		status := "OFF"
		if val {
			status = "ON"
		}
		fmt.Fprintf(&b, "  %-12s %s\n", key, status)
	}
	return b.String()
}

// setCassinoRule toggles a rule on/off. Returns false if the key is unknown.
func setCassinoRule(cfg *domain.CassinoConfig, key string, value bool) bool {
	switch key {
	case "multibuild":
		if cfg != nil {
			cfg.MultiBuildEnabled = value
		}
		return true
	case "sweepbonus":
		if cfg != nil {
			cfg.SweepBonusEnabled = value
		}
		return true
	default:
		return false
	}
}

// parseIntListArg は文字列引数を整数スライスへ変換し、失敗時にエラーメッセージを返す。
func parseIntListArg(args []string) ([]int, []string) {
	return cuiutil.ParseIntSlice(args)
}

// Exec コマンド実行。
//
//	take  <h> <t1 t2 ...> [b <bi>...]
//	build <h> <value> <t1 t2 ...>
//	trail <h>
//	reset / r / next / n
//	sd <0-2> / sr <key> <0|1>
//	log / l
func (c *CassinoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ci.GetConfig()
			return c.ci.ResetWithConfig(cfg)
		},
		[]string{
			"take", "t", "build", "b", "trail", "tr", "next", "n",
			"h", "hint", "sd", "setdifficulty", "sr", "setrule", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "take":
				return c.handleTake(args)
			case "b", "build":
				return c.handleBuild(args)
			case "tr", "trail":
				return c.handleTrail(args)
			case "n", "next":
				return c.ci.NextRound(), true
			case "h", "hint":
				return c.ci.Hint(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.CassinoCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sr", "setrule":
				if len(args) >= 1 && args[0] == "list" {
					cfg := c.ci.GetConfig()
					return "Rules:\n" + formatCassinoRuleList(&cfg), true
				}
				if len(args) < 2 {
					return "Usage: sr <rule> <0|1> | sr list\nRules: " + strings.Join(cassinoRuleKeys, ", "), true
				}
				if !setCassinoRule(nil, args[0], false) {
					return fmt.Sprintf("Unknown rule: %s.", args[0]), true
				}
				if args[1] != "0" && args[1] != "1" {
					return fmt.Sprintf("Invalid value: %s. Please enter 0 or 1.", args[1]), true
				}
				cfg := c.ci.GetConfig()
				setCassinoRule(&cfg, args[0], args[1] == "1")
				return c.ci.ResetWithConfig(cfg), true
			default:
				return handleCuiLog(cmd, c.ci.ActionLog)
			}
		},
	)
}

// handleTake は `t <h> <t1 t2 ...> [b <bi>...]` を処理する。
// "b" 区切り以降はビルド捕獲インデックスとして扱う。
func (c *CassinoCuiController) handleTake(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: t <handIdx> <tableIdx...> [b <buildIdx...>]", true
	}
	handStr := args[0]
	handIdx, _, ok := cuiutil.ParseIntArg([]string{handStr}, "hand index is required", "Invalid hand index: %s", 0, 51)
	if !ok {
		return "Invalid hand index: " + handStr, true
	}
	rest := args[1:]
	tableArgs := rest
	buildArgs := []string{}
	for i, a := range rest {
		if a == "b" {
			tableArgs = rest[:i]
			buildArgs = rest[i+1:]
			break
		}
	}
	tableIdxs, skippedT := parseIntListArg(tableArgs)
	buildIdxs, skippedB := parseIntListArg(buildArgs)
	skipped := append(skippedT, skippedB...)
	return cuiutil.PrependSkippedWarning(c.ci.Take(handIdx, tableIdxs, buildIdxs), skipped), true
}

// handleBuild は `b <h> <value> <t1 t2 ...>` を処理する。
func (c *CassinoCuiController) handleBuild(args []string) (string, bool) {
	if len(args) < 3 {
		return "Usage: b <handIdx> <value> <tableIdx...>", true
	}
	handIdx, _, ok := cuiutil.ParseIntArg([]string{args[0]}, "hand index is required", "Invalid hand index: %s", 0, 51)
	if !ok {
		return "Invalid hand index: " + args[0], true
	}
	value, _, ok := cuiutil.ParseIntArg([]string{args[1]}, "build value is required", "Invalid build value: %s", 2, 10)
	if !ok {
		return "Invalid build value: " + args[1], true
	}
	tableIdxs, skipped := parseIntListArg(args[2:])
	return cuiutil.PrependSkippedWarning(c.ci.Build(handIdx, tableIdxs, value), skipped), true
}

// handleTrail は `tr <h>` を処理する。
func (c *CassinoCuiController) handleTrail(args []string) (string, bool) {
	if len(args) < 1 {
		return "Usage: tr <handIdx>", true
	}
	handIdx, _, ok := cuiutil.ParseIntArg([]string{args[0]}, "hand index is required", "Invalid hand index: %s", 0, 51)
	if !ok {
		return "Invalid hand index: " + args[0], true
	}
	return c.ci.Trail(handIdx), true
}
