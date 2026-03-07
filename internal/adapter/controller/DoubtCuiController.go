package controller

import (
	"fmt"
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoubtCuiController ダウトCUIコントローラークラス
type DoubtCuiController struct {
	di usecase.DoubtInteractorIF
}

// NewDoubtCuiController コンストラクタ
func NewDoubtCuiController(di usecase.DoubtInteractorIF) *DoubtCuiController {
	return &DoubtCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	p <値> <idx...>  → カードを出す (値=宣言する値, idx=カードインデックス)
//	play <値> <idx...> (同上)
//	d <idx...>       → ダウト (idx=ダウトするプレイヤーインデックス)
//	doubt <idx...>   (同上)
//	s / skip         → ダウトをスキップ
//	sw / setwindow   → ダウト待機秒数設定 (1-60)
//	sm / setmemory   → CPU記憶力設定 (0=Easy, 1=Normal, 2=Hard)
//	sp / setpenalty  → ペナルティドロー上限設定 (0=無制限, >0=上限)
func (c *DoubtCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		unknownCommandMessage,
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				claimedValue := 0
				if len(args) > 0 {
					if parsed, err := strconv.Atoi(args[0]); err == nil {
						claimedValue = parsed
					}
				}
				cardIndices := []int{}
				if len(args) > 1 {
					for _, f := range args[1:] {
						if parsed, err := strconv.Atoi(f); err == nil {
							cardIndices = append(cardIndices, parsed)
						}
					}
				}
				return c.di.Play(cardIndices, claimedValue), true
			case "d", "doubt":
				doubterIndices := []int{}
				for _, f := range args {
					if parsed, err := strconv.Atoi(f); err == nil {
						doubterIndices = append(doubterIndices, parsed)
					}
				}
				return c.di.ResolveDoubt(doubterIndices), true
			case "s", "skip":
				return c.di.SkipDoubt(), true
			case "sw", "setwindow":
				if len(args) < 1 {
					return "Doubt window seconds is required (1-60).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 1 || v > 60 {
					return fmt.Sprintf("Invalid doubt window: %s. Please enter 1-60.", args[0]), true
				}
				cfg := c.di.GetConfig()
				cfg.DoubtWindowSec = v
				return c.di.ResetWithConfig(cfg), true
			case "sm", "setmemory":
				if len(args) < 1 {
					return "CPU memory level is required (0=Easy, 1=Normal, 2=Hard).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 || v > 2 {
					return fmt.Sprintf("Invalid CPU memory level: %s. Please enter 0-2.", args[0]), true
				}
				cfg := c.di.GetConfig()
				cfg.CpuMemoryLevel = domain.DoubtMemoryLevel(v)
				return c.di.ResetWithConfig(cfg), true
			case "sp", "setpenalty":
				if len(args) < 1 {
					return "Penalty draw limit is required (0=unlimited, >0=limit).", true
				}
				v, err := strconv.Atoi(args[0])
				if err != nil || v < 0 {
					return fmt.Sprintf("Invalid penalty draw limit: %s. Please enter 0 or more.", args[0]), true
				}
				cfg := c.di.GetConfig()
				cfg.PenaltyDrawLimit = v
				return c.di.ResetWithConfig(cfg), true
			}
			return "", false
		},
	)
}
