//go:build !js || !wasm || extra4

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SchafkopfCuiController シャーフコップのCUIコントローラークラス
type SchafkopfCuiController struct {
	si usecase.SchafkopfInteractorIF
}

// NewSchafkopfCuiController コンストラクタ
func NewSchafkopfCuiController(si usecase.SchafkopfInteractorIF) *SchafkopfCuiController {
	return &SchafkopfCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit              → ゲーム終了 ("bye.")
//	r / reset             → ゲームリセット (設定保持)
//	pick / p              → ブラインドを取る (ピックフェーズ)
//	pass                  → パス (ピックフェーズ)
//	c <suit> / call <suit>   → 呼びスートを指定 (呼びフェーズ; 1=♠ 2=♣ 3=♥)
//	<n> / play <n>        → カードをプレイ (プレイフェーズ)
//	n / next              → 次のトリックへ
//	nr / nextround        → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sb / setchips <n>     → 基本チップ設定
//	h / hint              → ヒント表示
//	log / l               → 棋譜表示
func (c *SchafkopfCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"pick", "p", "wenz", "w", "solo", "so", "pass", "c", "call", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sb", "setchips", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "pick", "p":
				return c.si.Declare(true, domain.SchafkopfContractRufspiel, 0), true
			case "wenz", "w":
				return c.si.Declare(true, domain.SchafkopfContractWenz, 0), true
			case "solo", "so":
				// Solo は切り札スートを引数に取る。既定値を置くと、指定を
				// 忘れた宣言が黙って別のスートの Solo になる。
				return cuiutil.WithParsedIntKeys(args, "suitRequiredThree", "invalidSuitThree",
					domain.CardDesignSpade, domain.CardDesignMax, func(suit int) string {
						return c.si.Declare(true, domain.SchafkopfContractSolo, suit)
					})
			case "pass":
				return c.si.Declare(false, domain.SchafkopfContractRufspiel, 0), true
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
					cfg.CpuDifficulty = domain.SchafkopfCpuDifficulty(v)
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
