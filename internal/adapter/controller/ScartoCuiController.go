//go:build !js || !wasm || extra4

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// ScartoCuiController スカルト (Scarto) のCUIコントローラークラス
type ScartoCuiController struct {
	di usecase.ScartoInteractorIF
}

// NewScartoCuiController コンストラクタ
func NewScartoCuiController(di usecase.ScartoInteractorIF) *ScartoCuiController {
	return &ScartoCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                          → ゲーム終了 ("bye.")
//	r / reset                         → ゲームリセット (設定保持)
//	scarto <i0> <i1> <i2>             → スカルトで 3 枚を捨てる (親のみ)
//	discard <i0> <i1> <i2>            → scarto のエイリアス
//	<n> / play <n>                    → カードをプレイ (プレイフェーズ)
//	n / next                          → 次のトリックへ
//	nr / nextround                    → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2>          → CPU難易度設定
//	h / hint                          → ヒント表示
//	log / l                           → 棋譜表示
func (c *ScartoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"scarto", "discard", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "scarto", "discard":
				return c.execScarto(args)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.ScartoCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execScarto scarto / discard サブコマンドを解釈する (3 枚のインデックス)。
func (c *ScartoCuiController) execScarto(args []string) (string, bool) {
	if len(args) < domain.ScartoSurplus {
		return invalidArg("threeIndicesRequiredScarto"), true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.di.Discard(indices), true
}
