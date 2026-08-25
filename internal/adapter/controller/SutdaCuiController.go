//go:build !js || !wasm || extra2

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SutdaCuiController はソッタの CUI コントローラー。
type SutdaCuiController struct {
	di usecase.SutdaInteractorIF
}

// NewSutdaCuiController コンストラクタ。
func NewSutdaCuiController(di usecase.SutdaInteractorIF) *SutdaCuiController {
	return &SutdaCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	c / call                 → コール (差額 0 ならチェック)
//	b / raise                → レイズ (1 単位上げる)
//	f / fold                 → フォールド (降りる)
//	nh / nexthand            → 次のハンドへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	ss / setseats <2-5>      → 席数設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *SutdaCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"c", "call", "b", "raise", "f", "fold", "nh", "nexthand",
			"sd", "setdifficulty", "ss", "setseats", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "c", "call":
				return c.di.Action(domain.SutdaActionCall), true
			case "b", "raise":
				return c.di.Action(domain.SutdaActionRaise), true
			case "f", "fold":
				return c.di.Action(domain.SutdaActionFold), true
			case "nh", "nexthand":
				return c.di.NextHand(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.SutdaCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "ss", "setseats":
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired", "invalidPlayerCount25",
					domain.SutdaMinSeats, domain.SutdaMaxSeats,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.Seats = v
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
