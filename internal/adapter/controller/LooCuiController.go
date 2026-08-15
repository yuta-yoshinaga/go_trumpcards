//go:build !js || !wasm || extra3

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LooCuiController はルー (Loo) の CUI コントローラークラス。
type LooCuiController struct {
	li usecase.LooInteractorIF
}

// NewLooCuiController コンストラクタ。
func NewLooCuiController(li usecase.LooInteractorIF) *LooCuiController {
	return &LooCuiController{li: li}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	d <0-1> / play / pass    → 参加 (1/play) または降り (0/pass)
//	p / play <n>             → カードをプレイ
//	nr / nextround / n       → 次のディールへ (精算)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *LooCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.li.GetConfig()
			return c.li.ResetWithConfig(cfg)
		},
		[]string{
			"d", "decide", "play", "pass", "p",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "decide":
				return cuiutil.WithParsedInt(args, "Decision is required (0=pass, 1=play).", "Invalid decision: %s. Please enter 0 or 1.", 0, 1, func(v int) string {
					return c.li.Decide(v == 1)
				})
			case "play":
				return c.li.Decide(true), true
			case "pass":
				return c.li.Decide(false), true
			case "p":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", 0, domain.LooHandSize-1, c.li.Play)
			case "n", "next", "nr", "nextround":
				return c.li.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.li.GetConfig()
					cfg.CpuDifficulty = domain.LooCpuDifficulty(v)
					return c.li.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.li.Hint, c.li.ActionLog)
			}
		},
	)
}
