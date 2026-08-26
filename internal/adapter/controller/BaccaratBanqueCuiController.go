//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// BaccaratBanqueCuiController はバカラ・バンクの CUI コントローラー。
type BaccaratBanqueCuiController struct {
	di usecase.BaccaratBanqueInteractorIF
}

// NewBaccaratBanqueCuiController コンストラクタ。
func NewBaccaratBanqueCuiController(di usecase.BaccaratBanqueInteractorIF) *BaccaratBanqueCuiController {
	return &BaccaratBanqueCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	draw / d                 → 3 枚目を引く
//	stand / s                → 3 枚目を引かない
//	nextcoup / nc            → 次のクーへ
//	retire                   → バンクを降りる
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	sc / setchips <100-100000>  → 元手
//	sb / setbet <10-500>     → 1 つの子の張り額
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *BaccaratBanqueCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"draw", "d", "stand", "s", "nextcoup", "nc", "retire",
			"sd", "setdifficulty", "sc", "setchips", "sb", "setbet",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "draw", "d":
				return c.di.Draw(), true
			case "stand", "s":
				return c.di.Stand(), true
			case "nextcoup", "nc":
				return c.di.NextCoup(), true
			case "retire":
				return c.di.Retire(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.BaccaratBanqueCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "sc", "setchips":
				return cuiutil.WithParsedIntKeys(args, "chipsRequired", "baccaratbanque.invalidChips",
					domain.BaccaratBanqueMinChips, domain.BaccaratBanqueMaxChips,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.StartChips = v
						return c.di.ResetWithConfig(cfg)
					})
			case "sb", "setbet":
				return cuiutil.WithParsedIntKeys(args, "betRequired", "baccaratbanque.invalidBet",
					domain.BaccaratBanqueMinBet, domain.BaccaratBanqueMaxBet,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.BetAmount = v
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
