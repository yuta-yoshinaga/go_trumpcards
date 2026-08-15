//go:build !js || !wasm || solo

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// MinchiateCuiController ミンキアーテのCUIコントローラークラス
type MinchiateCuiController struct {
	di usecase.MinchiateInteractorIF
}

// NewMinchiateCuiController コンストラクタ
func NewMinchiateCuiController(di usecase.MinchiateInteractorIF) *MinchiateCuiController {
	return &MinchiateCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	scarto <i0> ... <i12>    → スカルトで余剰札を捨てる (ディーラーのみ, MinchiateSurplus 枚)
//	discard <i0> ...         → scarto のエイリアス
//	play <n>                 → カードをプレイ (プレイフェーズ)
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *MinchiateCuiController) Exec(command string) string {
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
					cfg.CpuDifficulty = domain.MinchiateCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}

// execScarto scarto / discard サブコマンドを解釈する。
//
// **枚数は MinchiateSurplus から出す。**メッセージに数字を直接書くと、余剰枚数が
// 変わったときに案内だけが古くなる。
func (c *MinchiateCuiController) execScarto(args []string) (string, bool) {
	if len(args) < domain.MinchiateSurplus {
		return invalidArg("cardIndicesRequiredScartoN", "n", fmt.Sprint(domain.MinchiateSurplus)), true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.di.Discard(indices), true
}
