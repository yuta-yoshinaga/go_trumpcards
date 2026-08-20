//go:build !js || !wasm || solo

package controller

import (
	"fmt"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TarocchiniCuiController タロッキーニのCUIコントローラークラス
type TarocchiniCuiController struct {
	di usecase.TarocchiniInteractorIF
}

// NewTarocchiniCuiController コンストラクタ
func NewTarocchiniCuiController(di usecase.TarocchiniInteractorIF) *TarocchiniCuiController {
	return &TarocchiniCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                 → ゲーム終了 ("bye.")
//	r / reset                → ゲームリセット (設定保持)
//	scarto <i0> <i1>         → スカルトで 2 枚を捨てる (ディーラーのみ)
//	discard <i0> <i1>        → scarto のエイリアス
//	play <n>                 → カードをプレイ (プレイフェーズ)
//	n / next                 → 次のトリックへ
//	nr / nextround           → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint                 → ヒント表示
//	log / l                  → 棋譜表示
func (c *TarocchiniCuiController) Exec(command string) string {
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
					cfg.CpuDifficulty = domain.TarocchiniCpuDifficulty(v)
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
// **枚数は TarocchiniSurplus から出す。**メッセージに数字を直接書くと、余剰枚数が
// 変わったときに案内だけが古くなる。
func (c *TarocchiniCuiController) execScarto(args []string) (string, bool) {
	if len(args) < domain.TarocchiniSurplus {
		return invalidArg("cardIndicesRequiredScartoTwo", "n", fmt.Sprint(domain.TarocchiniSurplus)), true
	}
	indices, skipped := cuiutil.ParseIntSlice(args)
	if len(skipped) > 0 {
		return invalidArg("invalidCardIndex", "val", skipped[0]), true
	}
	return c.di.Discard(indices), true
}
