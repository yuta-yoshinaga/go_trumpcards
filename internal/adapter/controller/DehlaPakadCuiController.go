//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DehlaPakadCuiController はデーラ・パカドの CUI コントローラー。
type DehlaPakadCuiController struct {
	di usecase.DehlaPakadInteractorIF
}

// NewDehlaPakadCuiController コンストラクタ。
func NewDehlaPakadCuiController(di usecase.DehlaPakadInteractorIF) *DehlaPakadCuiController {
	return &DehlaPakadCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	t / trump <1-4>          → 切り札を宣言する (親の右隣のみ)
//	p / play <idx>           → カードを出す
//	nh / nexthand            → 次のハンドへ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	sk / setkots <1-5>       → 勝利に必要なコート数
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *DehlaPakadCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"t", "trump", "p", "play", "nh", "nexthand",
			"sd", "setdifficulty", "sk", "setkots", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "suitRequired", "invalidTrumpSuit",
					domain.CardDesignSpade, domain.CardDesignDiamond,
					func(v int) string { return c.di.SelectTrump(v) })
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, func(v int) string { return c.di.Play(v) })
			case "nh", "nexthand":
				return c.di.NextHand(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.DehlaPakadCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "sk", "setkots":
				return cuiutil.WithParsedIntKeys(args, "targetKotsRequired", "invalidTargetKots",
					domain.DehlaPakadMinKots, domain.DehlaPakadMaxKots,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.TargetKots = v
						return c.di.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
