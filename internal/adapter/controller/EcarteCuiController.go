//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// EcarteCuiController エカルテCUIコントローラークラス
type EcarteCuiController struct {
	ei usecase.EcarteInteractorIF
}

// NewEcarteCuiController コンストラクタ
func NewEcarteCuiController(ei usecase.EcarteInteractorIF) *EcarteCuiController {
	return &EcarteCuiController{ei: ei}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	pr / propose             → 交換を提案する (elder)
//	st / stand               → 交換せず勝負する (elder)
//	a / accept               → 提案を承諾する (親)
//	rf / refuse              → 提案を拒否する (親)
//	d / discard <i j k>      → 捨て札を選んで引き直す
//	p / play <i>             → カードをプレイ
//	n / next / nextround     → 次のディールへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	tg / settarget <n>       → ターゲットスコア設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *EcarteCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.ei.GetConfig()
			return c.ei.ResetWithConfig(cfg)
		},
		[]string{
			"pr", "propose", "st", "stand", "a", "accept", "rf", "refuse",
			"d", "discard", "p", "play",
			"n", "next", "nextround",
			"sd", "setdifficulty", "tg", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "pr", "propose":
				return c.ei.Propose(), true
			case "st", "stand":
				return c.ei.Stand(), true
			case "a", "accept":
				return c.ei.Respond(true), true
			case "rf", "refuse":
				return c.ei.Respond(false), true
			case "d", "discard":
				indices, skipped := cuiutil.ParseBoundedIntSlice(args, 0, domain.EcarteHandSize-1)
				return cuiutil.PrependSkippedWarning(c.ei.Discard(indices), skipped), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.ei.Play)
			case "n", "next", "nextround":
				return c.ei.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.ei.GetConfig()
					cfg.CpuDifficulty = domain.EcarteCpuDifficulty(v)
					return c.ei.ResetWithConfig(cfg)
				})
			case "tg", "settarget":
				return cuiutil.WithParsedIntKeys(args, "targetScoreRequired", "invalidTargetScore", 1, math.MaxInt, func(v int) string {
					cfg := c.ei.GetConfig()
					cfg.TargetScore = v
					return c.ei.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.ei.Hint, c.ei.ActionLog)
			}
		},
	)
}
