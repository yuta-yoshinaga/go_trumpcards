//go:build !js || !wasm || extra

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HachiHachiCuiController は八八 (Hachi-Hachi) の CUI コントローラークラス。
type HachiHachiCuiController struct {
	ki usecase.HachiHachiInteractorIF
}

// NewHachiHachiCuiController コンストラクタ。
func NewHachiHachiCuiController(ki usecase.HachiHachiInteractorIF) *HachiHachiCuiController {
	return &HachiHachiCuiController{ki: ki}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	p / play <h> [f]         → 手札 h を出す。f は 2 枚一致時の捕獲場札インデックス
//	nr / nextround / n       → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *HachiHachiCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ki.GetConfig()
			return c.ki.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.handlePlay(args), true
			case "n", "next", "nr", "nextround":
				return c.ki.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ki.GetConfig()
					cfg.CpuDifficulty = domain.HachiHachiCpuDifficulty(v)
					return c.ki.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ki.Hint, c.ki.ActionLog)
			}
		},
	)
}

// handlePlay は "p <handIdx> [fieldIdx]" を解析して Play を呼ぶ。
func (c *HachiHachiCuiController) handlePlay(args []string) string {
	if len(args) == 0 {
		return invalidArg("cardIndexRequiredField")
	}
	handIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[0])
	}
	fieldIdx := -1
	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[1]); err == nil {
			fieldIdx = v
		}
	}
	return c.ki.Play(handIdx, fieldIdx)
}
