package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
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
		[]string{"p", "play", "d", "doubt", "s", "skip", "sw", "setwindow", "sm", "setmemory", "smetaai", "smai", "rp", "resetprofile", "sp", "setpenalty"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				claimedValue := cuiutil.ParseOptionalInt(args, 0, 0)
				var cardArgs []string
				if len(args) > 1 {
					cardArgs = args[1:]
				}
				return c.di.Play(cuiutil.ParseIntSlice(cardArgs), claimedValue), true
			case "d", "doubt":
				return c.di.ResolveDoubt(cuiutil.ParseIntSlice(args)), true
			case "s", "skip":
				return c.di.SkipDoubt(), true
			case "sw", "setwindow":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Doubt window seconds is required (1-60).", "Invalid doubt window: %s. Please enter 1-60.", 1, 60)
				if !ok {
					return errMsg, true
				}
				cfg := c.di.GetConfig()
				cfg.DoubtWindowSec = v
				return c.di.ResetWithConfig(cfg), true
			case "sm", "setmemory":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "CPU memory level is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU memory level: %s. Please enter 0-2.", 0, 2)
				if !ok {
					return errMsg, true
				}
				cfg := c.di.GetConfig()
				cfg.CpuMemoryLevel = domain.DoubtMemoryLevel(v)
				return c.di.ResetWithConfig(cfg), true
			case "smetaai", "smai":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Meta-AI flag is required (0=OFF, 1=ON).", "Invalid meta-AI flag: %s. Please enter 0-1.", 0, 1)
				if !ok {
					return errMsg, true
				}
				cfg := c.di.GetConfig()
				cfg.CpuMetaAI = v == 1
				return c.di.ResetWithConfig(cfg), true
			case "rp", "resetprofile":
				return c.di.ResetProfile(), true
			case "sp", "setpenalty":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Penalty draw limit is required (0=unlimited, >0=limit).", "Invalid penalty draw limit: %s. Please enter 0 or more.", 0, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				cfg := c.di.GetConfig()
				cfg.PenaltyDrawLimit = v
				return c.di.ResetWithConfig(cfg), true
			}
			return "", false
		},
	)
}
