//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MusCuiController ムスのCUIコントローラークラス
type MusCuiController struct {
	mi usecase.MusInteractorIF
}

// NewMusCuiController コンストラクタ
func NewMusCuiController(mi usecase.MusInteractorIF) *MusCuiController {
	return &MusCuiController{mi: mi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                  → ゲーム終了 ("bye.")
//	r / reset                 → ゲームリセット (設定保持)
//	m / mus                   → Mus を宣言する (Mus フェーズ)
//	c / cut / corte           → Corte を宣言する (Mus フェーズ)
//	d <i>... / discard <i>... → 指定インデックスの札を交換 (Discard フェーズ)
//	paso                      → パス (賭けフェーズ)
//	e <n> / envido <n>        → Envido (賭けフェーズ; <n>=賭け額)
//	ordago                    → Ordago 宣言 (賭けフェーズ)
//	q / quiero                → 賭けを受ける (賭けフェーズ; q は quit と衝突しないよう注意)
//	nq / noquiero             → 賭けを降りる (賭けフェーズ)
//	n / next                  → 次のラウンドへ (RoundEnd フェーズ)
//	sd [0-2]                  → CPU難易度設定
//	h / hint                  → ヒント表示
//	l / log                   → 棋譜表示
func (c *MusCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.mi.GetConfig()
			return c.mi.ResetWithConfig(cfg)
		},
		[]string{
			"m", "mus", "c", "cut", "corte",
			"d", "discard",
			"paso",
			"e", "envido",
			"ordago",
			"quiero",
			"nq", "noquiero",
			"n", "next",
			"sd", "setdifficulty",
			"h", "hint", "l", "log",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "m", "mus":
				return c.mi.Mus(true), true
			case "c", "cut", "corte":
				return c.mi.Mus(false), true
			case "d", "discard":
				return c.handleDiscard(args)
			case "paso":
				return c.mi.Bet(domain.MusActionPaso, 0), true
			case "e", "envido":
				return cuiutil.WithParsedIntKeys(args, "amountRequiredEGE2", "invalidAmount", 1, 40, func(v int) string {
					return c.mi.Bet(domain.MusActionEnvido, v)
				})
			case "ordago":
				return c.mi.Bet(domain.MusActionOrdago, 0), true
			case "quiero":
				return c.mi.Bet(domain.MusActionQuiero, 0), true
			case "nq", "noquiero":
				return c.mi.Bet(domain.MusActionNoQuiero, 0), true
			case "n", "next":
				return c.mi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.mi.GetConfig()
					cfg.CpuDifficulty = domain.MusCpuDifficulty(v)
					return c.mi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.mi.Hint, c.mi.ActionLog)
			}
		},
	)
}

// handleDiscard は "d <i>..." コマンドを処理する。引数なしは0枚交換。
func (c *MusCuiController) handleDiscard(args []string) (string, bool) {
	indices := make([]int, 0, len(args))
	for _, a := range args {
		v, _, ok := cuiutil.ParseIntArgKeys([]string{a}, "indexRequired", "invalidIndexPlain", 0, 3)
		if !ok {
			return "Invalid card index. Usage: d <idx>...", true
		}
		indices = append(indices, v)
	}
	return c.mi.Discard(indices), true
}
