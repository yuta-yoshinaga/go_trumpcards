package controller

import (
	"math"
	"strings"

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
//	sh / sethesitation → CPU の迷い時間演出 (0=OFF, 1=ON)
func (c *DoubtCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg, nil)
		},
		[]string{
			"p", "play", "d", "doubt", "s", "skip", "sw", "setwindow",
			"sm", "setmemory", "smetaai", "smai", "rp", "resetprofile", "sp", "setpenalty",
			"sh", "sethesitation",
			"log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				claimedValue, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 0, 0, "invalidClaimedValue")
				if !ok {
					return errMsg, true
				}
				var cardArgs []string
				if len(args) > 1 {
					cardArgs = args[1:]
				}
				indices, skipped := cuiutil.ParseIntSlice(cardArgs)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return c.di.Play(indices, claimedValue, 0), true
			case "d", "doubt":
				indices, skipped := cuiutil.ParseIntSlice(args)
				// Refuse before playing. PrependSkippedWarning ran the move first and
				// put the warning above the new board, so a mistyped index was dropped
				// and the remaining ones played as a different, legal move (issue #5390).
				if len(skipped) > 0 {
					return invalidArg("invalidCardIndex", "val", strings.Join(skipped, ", ")), true
				}
				return c.di.ResolveDoubt(indices), true
			case "s", "skip":
				return c.di.SkipDoubt(), true
			case "sw", "setwindow":
				return cuiutil.WithParsedIntKeys(args, "doubtWindowSecondsRequired160", "invalidDoubtWindow160", 1, 60, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.DoubtWindowSec = v
					return c.di.ResetWithConfig(cfg, nil)
				})
			case "sm", "setmemory":
				return cuiutil.WithParsedIntKeys(args, "cpuMemoryLevelRequired0Easy1Normal2Hard", "invalidCpuMemoryLevel02", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuMemoryLevel = domain.DoubtMemoryLevel(v)
					return c.di.ResetWithConfig(cfg, nil)
				})
			case "smetaai", "smai":
				return cuiutil.WithParsedIntKeys(args, "metaAiFlagRequired0Off1On", "invalidMetaAiFlag01", 0, 1, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuMetaAI = v == 1
					return c.di.ResetWithConfig(cfg, nil)
				})
			case "sh", "sethesitation":
				return cuiutil.WithParsedIntKeys(args, "hesitationFlagRequired0Off1On", "invalidHesitationFlag01", 0, 1, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuHesitationEnabled = v == 1
					return c.di.ResetWithConfig(cfg, nil)
				})
			case "rp", "resetprofile":
				return c.di.ResetProfile(), true
			case "sp", "setpenalty":
				return cuiutil.WithParsedIntKeys(args, "penaltyDrawLimitRequired0Unlimited0Limit", "invalidPenaltyDrawLimit0OrMore", 0, math.MaxInt, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.PenaltyDrawLimit = v
					return c.di.ResetWithConfig(cfg, nil)
				})
			default:
				return handleCuiLog(cmd, c.di.ActionLog)
			}
		},
	)
}
