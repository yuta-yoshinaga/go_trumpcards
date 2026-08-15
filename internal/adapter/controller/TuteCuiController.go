//go:build !js || !wasm || casino

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// TuteCuiController トゥーテのCUIコントローラークラス
type TuteCuiController struct {
	di usecase.TuteInteractorIF
}

// NewTuteCuiController コンストラクタ
func NewTuteCuiController(di usecase.TuteInteractorIF) *TuteCuiController {
	return &TuteCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                   → ゲーム終了 ("bye.")
//	r / reset                  → ゲームリセット (設定保持)
//	<n> / play <n>             → カードをプレイ (プレイフェーズ)
//	m <suit> / marriage <suit> → 結婚宣言 (suit: 1=♠ 2=♣ 3=♥ 4=♦)
//	tute                       → Tute を宣言する (4枚の K または Q)
//	n / next                   → 次のトリックへ
//	nr / nextround             → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2>   → CPU難易度設定
//	h / hint                   → ヒント表示
//	log / l                    → 棋譜表示
func (c *TuteCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.di.GetConfig()
			return c.di.ResetWithConfig(cfg)
		},
		[]string{
			"play", "m", "marriage", "tute",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.di.Play)
			case "m", "marriage":
				return cuiutil.WithParsedIntKeys(args, "suitRequiredSymbolsPlain", "invalidSuitRange", 1, domain.CardDesignMax, c.di.DeclareMarriage)
			case "tute":
				return c.di.DeclareTute(), true
			case "n", "next":
				return c.di.NextTrick(), true
			case "nr", "nextround":
				return c.di.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.di.GetConfig()
					cfg.CpuDifficulty = domain.TuteCpuDifficulty(v)
					return c.di.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.di.Hint, c.di.ActionLog)
			}
		},
	)
}
