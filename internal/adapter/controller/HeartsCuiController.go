package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// HeartsCuiController ハーツCUIコントローラークラス
type HeartsCuiController struct {
	hi usecase.HeartsInteractorIF
}

// NewHeartsCuiController コンストラクタ
func NewHeartsCuiController(hi usecase.HeartsInteractorIF) *HeartsCuiController {
	return &HeartsCuiController{hi: hi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	pass <i1> <i2> <i3> → カード3枚を交換
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	nr / nextround   → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n> → ポイント上限設定
//	log / l          → 棋譜表示
func (c *HeartsCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.hi.GetConfig()
			return c.hi.ResetWithConfig(cfg)
		},
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "pass":
				indices := []int{}
				for _, f := range args {
					if parsed, err := strconv.Atoi(f); err == nil {
						indices = append(indices, parsed)
					}
				}
				if len(indices) != 3 {
					return "Pass requires exactly 3 card indices.", true
				}
				return c.hi.Pass(indices), true
			case "p", "play":
				if len(args) < 1 {
					return "Card index is required.", true
				}
				idx, err := strconv.Atoi(args[0])
				if err != nil {
					return fmt.Sprintf("Invalid card index: %s.", args[0]), true
				}
				return c.hi.Play(idx), true
			case "n", "next":
				return c.hi.NextTrick(), true
			case "nr", "nextround":
				return c.hi.NextRound(), true
			case "sd", "setdifficulty":
				if len(args) < 1 {
					return "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 2 {
					return fmt.Sprintf("Invalid CPU difficulty: %s. Please enter 0-2.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.CpuDifficulty = domain.HeartsCpuDifficulty(v)
				return c.hi.ResetWithConfig(cfg), true
			case "sl", "setlimit":
				if len(args) < 1 {
					return "Point limit is required.", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 1 {
					return fmt.Sprintf("Invalid point limit: %s. Please enter 1 or more.", args[0]), true
				}
				cfg := c.hi.GetConfig()
				cfg.PointLimit = v
				return c.hi.ResetWithConfig(cfg), true
			case "log", "l":
				return c.hi.ActionLog(), true
			}
			return "", false
		},
	)
}
