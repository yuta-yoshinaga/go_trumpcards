//go:build !js || !wasm || solo

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// FiveHundredCuiController 500 CUIコントローラークラス
type FiveHundredCuiController struct {
	fi usecase.FiveHundredInteractorIF
}

// NewFiveHundredCuiController コンストラクタ
func NewFiveHundredCuiController(fi usecase.FiveHundredInteractorIF) *FiveHundredCuiController {
	return &FiveHundredCuiController{fi: fi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit           → ゲーム終了 ("bye.")
//	r / reset          → ゲームリセット (設定保持)
//	b / bid <tricks> <suit> → 切り札ビッド (suit: 1=S,2=C,3=H,4=D, tricks: 6-10)
//	bnt <tricks>       → ノートランプビッド (6-10)
//	m / misere         → ミゼール
//	om / openmisere    → オープンミゼール
//	pa / pass          → パス
//	e / exchange <i> <j> <k> → キティ交換 (捨てるカード3枚)
//	p / play <i> [s]   → カードをプレイ (NTでジョーカーリード時 s=指名スート 1-4)
//	n / next           → 次のトリックへ
//	nr / nextround     → 次のラウンドへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n> → 勝利スコア設定
//	h / hint           → ヒント表示
//	log / l            → 棋譜表示
func (c *FiveHundredCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.fi.GetConfig()
			return c.fi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "bnt",
			"m", "misere", "om", "openmisere",
			"pa", "pass",
			"e", "exchange", "p", "play",
			"n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				if len(args) < 2 {
					return "Usage: bid <tricks> <suit>  (tricks 6-10, suit 1=S 2=C 3=H 4=D)\n", true
				}
				tricks, errMsg, ok := cuiutil.ParseIntArg(args[:1], "", "Invalid tricks: %s.", 6, 10)
				if !ok {
					return errMsg, true
				}
				suit, errMsg, ok := cuiutil.ParseIntArg(args[1:2], "", "Invalid suit: %s.", 1, 4)
				if !ok {
					return errMsg, true
				}
				return c.fi.Bid(domain.FiveHundredContractSuit, tricks, suit), true
			case "bnt":
				return cuiutil.WithParsedInt(args, "Tricks is required (6-10).", "Invalid tricks: %s.", 6, 10, func(t int) string {
					return c.fi.Bid(domain.FiveHundredContractNoTrump, t, -1)
				})
			case "m", "misere":
				return c.fi.Bid(domain.FiveHundredContractMisere, 0, -1), true
			case "om", "openmisere":
				return c.fi.Bid(domain.FiveHundredContractOpenMisere, 0, -1), true
			case "pa", "pass":
				return c.fi.Pass(), true
			case "e", "exchange":
				if len(args) < 3 {
					return "Usage: exchange <i> <j> <k>  (three card indices to discard)\n", true
				}
				idxs := make([]int, 3)
				for i := 0; i < 3; i++ {
					v, errMsg, ok := cuiutil.ParseIntArgKeys(args[i:i+1], "", "invalidCardIndex", 0, math.MaxInt)
					if !ok {
						return errMsg, true
					}
					idxs[i] = v
				}
				return c.fi.ExchangeKitty(idxs), true
			case "p", "play":
				cardIdx, errMsg, ok := cuiutil.ParseIntArgKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				jokerSuit := cuiutil.ParseOptionalInt(args, 1, -1)
				return c.fi.Play(cardIdx, jokerSuit), true
			case "n", "next":
				return c.fi.NextTrick(), true
			case "nr", "nextround":
				return c.fi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedIntKeys(args, "cpuDifficultyRequired", "invalidCpuDifficulty", 0, 2, func(v int) string {
					cfg := c.fi.GetConfig()
					cfg.CpuDifficulty = domain.FiveHundredCpuDifficulty(v)
					return c.fi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedInt(args, "Target score is required.", "Invalid target score: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.fi.GetConfig()
					cfg.TargetScore = v
					return c.fi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.fi.Hint, c.fi.ActionLog)
			}
		},
	)
}
