//go:build !js || !wasm || extra3

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SheepsheadCuiController シープスヘッドのCUIコントローラークラス
type SheepsheadCuiController struct {
	si usecase.SheepsheadInteractorIF
}

// NewSheepsheadCuiController コンストラクタ
func NewSheepsheadCuiController(si usecase.SheepsheadInteractorIF) *SheepsheadCuiController {
	return &SheepsheadCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit              → ゲーム終了 ("bye.")
//	r / reset             → ゲームリセット (設定保持)
//	pick / p              → ブラインドを取る (ピックフェーズ)
//	pass                  → パス (ピックフェーズ)
//	b <i> <j> / bury <i> <j> → 2枚を埋める (埋めフェーズ)
//	c <suit> / call <suit>   → 呼びスートを指定 (呼びフェーズ; 1=♠ 2=♣ 3=♥)
//	<n> / play <n>        → カードをプレイ (プレイフェーズ)
//	n / next              → 次のトリックへ
//	nr / nextround        → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sb / setchips <n>     → 基本チップ設定
//	h / hint              → ヒント表示
//	log / l               → 棋譜表示
func (c *SheepsheadCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"pick", "p", "pass", "b", "bury", "c", "call", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sb", "setchips", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "pick", "p":
				return c.si.Pick(true), true
			case "pass":
				return c.si.Pick(false), true
			case "b", "bury":
				return c.handleBury(args)
			case "c", "call":
				return cuiutil.WithParsedIntKeys(args, "suitRequiredThree", "invalidSuitThree", domain.CardDesignSpade, domain.CardDesignHeart, c.si.Call)
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "n", "next":
				return c.si.NextTrick(), true
			case "nr", "nextround":
				return c.si.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.CpuDifficulty = domain.SheepsheadCpuDifficulty(v)
					return c.si.ResetWithConfig(cfg)
				})
			case "sb", "setchips":
				return cuiutil.WithParsedIntKeys(args, "baseChipsRequired", "invalidBaseChips1OrMore", 1, math.MaxInt, func(v int) string {
					cfg := c.si.GetConfig()
					cfg.BaseChips = v
					return c.si.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// handleBury は "b <i> <j>" コマンドを処理する。
func (c *SheepsheadCuiController) handleBury(args []string) (string, bool) {
	if len(args) < 2 {
		return "Usage: b <idx1> <idx2> — specify two card indices to bury.", true
	}
	idx1, _, ok1 := cuiutil.ParseIntArgKeys(args[:1], "firstCardIndexRequired", "invalidCardIndexPlain", cuiutil.NoMin, cuiutil.NoMax)
	idx2, _, ok2 := cuiutil.ParseIntArgKeys(args[1:2], "secondCardIndexRequired", "invalidCardIndexPlain", cuiutil.NoMin, cuiutil.NoMax)
	if !ok1 || !ok2 {
		return invalidArg("invalidCardIndicesUsageB"), true
	}
	return c.si.Bury([]int{idx1, idx2}), true
}
