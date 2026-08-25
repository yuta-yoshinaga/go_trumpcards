//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CostlyColoursCuiController はコストリー・カラーズの CUI コントローラー。
type CostlyColoursCuiController struct {
	di usecase.CostlyColoursInteractorIF
}

// NewCostlyColoursCuiController コンストラクタ。
func NewCostlyColoursCuiController(di usecase.CostlyColoursInteractorIF) *CostlyColoursCuiController {
	return &CostlyColoursCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	mog / m                  → 交換に応じる
//	nomog / nm               → 交換を断る (相手に 1 点)
//	p / play <idx>           → 手札を 1 枚出す
//	nd / nextdeal            → 次のディールへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	st / settarget <31-121>  → 目標点
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *CostlyColoursCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"mog", "m", "nomog", "nm", "p", "play", "nd", "nextdeal",
			"sd", "setdifficulty", "st", "settarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "mog", "m":
				return c.di.Mog(true), true
			case "nomog", "nm":
				// **断る手は別コマンドにする。** 引数の有無で分けると、
				// 打ち間違いが黙って交換になる。
				return c.di.Mog(false), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "nd", "nextdeal":
				return c.di.NextDeal(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.CostlyColoursCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired",
					"costlycolours.invalidTargetScore",
					domain.CostlyColoursMinTarget, domain.CostlyColoursMaxTarget,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.TargetScore = v
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
