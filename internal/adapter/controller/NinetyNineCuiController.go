package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// NinetyNineCuiController ナインティナインCUIコントローラークラス
type NinetyNineCuiController struct {
	oi usecase.NinetyNineInteractorIF
}

// NewNinetyNineCuiController コンストラクタ
func NewNinetyNineCuiController(oi usecase.NinetyNineInteractorIF) *NinetyNineCuiController {
	return &NinetyNineCuiController{oi: oi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit               → ゲーム終了 ("bye.")
//	r / reset              → ゲームリセット (設定保持)
//	b / bid <i> <j> <k>    → 3枚を伏せてビッドを宣言
//	p / play <i>           → カードをプレイ
//	n / next               → 次のトリックへ
//	nr / nextround         → 次のディールへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n>     → 目標スコア設定
//	h / hint               → ヒント表示
//	log / l                → 棋譜表示
func (c *NinetyNineCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.oi.GetConfig()
			return c.oi.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "st", "settarget", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				if len(args) < domain.NinetyNineBurySize {
					return "Bury 3 card indices are required (e.g. 'bid 0 1 2').", true
				}
				idxs, skipped := cuiutil.ParseIntSlice(args)
				if len(skipped) > 0 || len(idxs) < domain.NinetyNineBurySize {
					return "Invalid bury indices. Please enter 3 valid card indices.", true
				}
				return c.oi.Bid(idxs[:domain.NinetyNineBurySize]), true
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.oi.Play)
			case "n", "next":
				return c.oi.NextTrick(), true
			case "nr", "nextround":
				return c.oi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.oi.GetConfig()
					cfg.CpuDifficulty = domain.NinetyNineCpuDifficulty(v)
					return c.oi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedInt(args, "Target score is required (10-1000).", "Invalid target score: %s. Please enter 10-1000.", 10, 1000, func(v int) string {
					cfg := c.oi.GetConfig()
					cfg.TargetScore = v
					return c.oi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.oi.Hint, c.oi.ActionLog)
			}
		},
	)
}
