//go:build !js || !wasm || solo

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// CometCuiController はコメットの CUI コントローラー。
type CometCuiController struct {
	di usecase.CometInteractorIF
}

// NewCometCuiController コンストラクタ。
func NewCometCuiController(di usecase.CometInteractorIF) *CometCuiController {
	return &CometCuiController{di: di}
}

// Exec コマンド実行。
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	p / play <idx>           → 手札を 1 枚出す
//	pass                     → パス (出せる札が無いときだけ)
//	nr / nextround           → 次の局へ
//	sd / setdifficulty <0-2> → CPU 難易度設定
//	sp / setplayers <2-5>    → 席数
//	st / settarget <20-200>  → 目標点
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *CometCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.ResetWithConfig(c.di.GetConfig()) },
		[]string{
			"p", "play", "pass", "nr", "nextround",
			"sd", "setdifficulty", "sp", "setplayers", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex",
					cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "pass":
				return c.di.Pass(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.CpuDifficulty = domain.CometCpuDifficulty(v)
						return c.di.ResetWithConfig(cfg)
					})
			case "sp", "setplayers":
				return cuiutil.WithParsedIntKeys(args, "playerCountRequired", "comet.invalidPlayerCount",
					domain.CometMinPlayers, domain.CometMaxPlayers,
					func(v int) string {
						cfg := c.di.GetConfig()
						cfg.Players = v
						return c.di.ResetWithConfig(cfg)
					})
			case "st", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "comet.invalidTargetScore",
					domain.CometMinTarget, domain.CometMaxTarget,
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
