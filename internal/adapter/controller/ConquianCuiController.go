//go:build !js || !wasm || extra

package controller

import (
	"math"
	"strconv"
	"strings"

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
				groups, targets := parseConquianMeld(args)
				return c.ci.MeldWithTargets(groups, targets), true
			case "d", "discard":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ci.Discard)
			case "nr", "nextround":
				return c.ci.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ci.GetConfig()
					cfg.CpuDifficulty = domain.ConquianCpuDifficulty(v)
					return c.ci.ResetWithConfig(cfg)
				})
			case "sw", "setwins":
				return cuiutil.WithParsedIntKeys(args, "targetWinsRequired", "invalidTargetWins1OrMore", 1, math.MaxInt, func(v int) string {
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

// parseConquianMeld parses groups and optional explicit extension targets.
// Each group may end in @meldIndex, for example "m 0,1,2;3@1".
func parseConquianMeld(args []string) ([][]int, []int) {
	if len(args) == 0 {
		return nil, nil
	}
	groups := make([][]int, 0)
	targets := make([]int, 0)
	for _, raw := range strings.Split(strings.Join(args, " "), ";") {
		parts := strings.SplitN(raw, "@", 2)
		indices := strings.FieldsFunc(parts[0], func(r rune) bool { return r == ',' || r == ' ' })
		group := make([]int, 0, len(indices))
		for _, s := range indices {
			if n, err := strconv.Atoi(s); err == nil {
				group = append(group, n)
			}
		}
		groups = append(groups, group)
		target := -1
		if len(parts) == 2 {
			target, _ = strconv.Atoi(strings.TrimSpace(parts[1]))
		}
		targets = append(targets, target)
	}
	return groups, targets
}
