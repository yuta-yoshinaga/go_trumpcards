package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TwoTenJackCuiController ツーテンジャックCUIコントローラークラス
type TwoTenJackCuiController struct {
	ti usecase.TwoTenJackInteractorIF
}

// NewTwoTenJackCuiController コンストラクタ
func NewTwoTenJackCuiController(ti usecase.TwoTenJackInteractorIF) *TwoTenJackCuiController {
	return &TwoTenJackCuiController{ti: ti}
}

// Exec コマンド実行
//
//	q / quit         → ゲーム終了
//	r / reset        → ゲームリセット (設定保持)
//	d / declare <s>  → トランプスート宣言 (1=S, 2=C, 3=H, 4=D)
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	nr / nextround   → 次のラウンドへ
//	sd <0-2>         → CPU難易度設定
//	sl <n>           → ポイント上限設定
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *TwoTenJackCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ti.GetConfig()
			return c.ti.ResetWithConfig(cfg)
		},
		[]string{
			"d", "declare", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "declare":
				return cuiutil.WithParsedInt(args, "Trump suit is required (1=S, 2=C, 3=H, 4=D).", "Invalid trump suit: %s.", domain.CardDesignSpade, domain.CardDesignDiamond, c.ti.DeclareTrump)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ti.Play)
			case "n", "next":
				return c.ti.NextTrick(), true
			case "nr", "nextround":
				return c.ti.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.CpuDifficulty = domain.TwoTenJackCpuDifficulty(v)
					return c.ti.ResetWithConfig(cfg)
				})
			case "sl", "setlimit":
				return cuiutil.WithParsedIntKeys(args, "pointLimitRequired", "invalidPointLimit", 1, math.MaxInt, func(v int) string {
					cfg := c.ti.GetConfig()
					cfg.PointLimit = v
					return c.ti.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ti.Hint, c.ti.ActionLog)
			}
		},
	)
}
