//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoppelkopfCuiController ドッペルコップのCUIコントローラークラス
type DoppelkopfCuiController struct {
	di usecase.DoppelkopfInteractorIF
}

// NewDoppelkopfCuiController コンストラクタ
func NewDoppelkopfCuiController(di usecase.DoppelkopfInteractorIF) *DoppelkopfCuiController {
	return &DoppelkopfCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit              → ゲーム終了 ("bye.")
//	r / reset             → ゲームリセット (設定保持)
//	<n> / play <n>        → カードをプレイ (プレイフェーズ)
//	a / announce          → Re/Kontra を宣言する
//	n / next              → 次のトリックへ
//	nr / nextround        → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sb / setchips <n>     → 基本チップ設定
//	h / hint              → ヒント表示
//	log / l               → 棋譜表示
func (c *DoppelkopfCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"play", "a", "announce",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sb", "setchips", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "a", "announce":
				return c.di.Announce(), true
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.DoppelkopfCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			case "sb", "setchips":
				return cuiutil.WithParsedIntKeys(args, "baseChipsRequired", "invalidBaseChips1OrMore", 1, math.MaxInt, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.BaseChips = v
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
