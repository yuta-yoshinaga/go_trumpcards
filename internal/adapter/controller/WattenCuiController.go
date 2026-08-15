//go:build !js || !wasm || extra

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// WattenCuiController ヴァッテンCUIコントローラークラス
type WattenCuiController struct {
	wi usecase.WattenInteractorIF
}

// NewWattenCuiController コンストラクタ
func NewWattenCuiController(wi usecase.WattenInteractorIF) *WattenCuiController {
	return &WattenCuiController{wi: wi}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit             → ゲーム終了 ("bye.")
//	r / reset            → ゲームリセット (設定保持)
//	d / declare <r> <s>  → Schlag ランク + 切り札スートを宣言 (人間ディーラー)
//	p / play <i>         → カードをプレイ
//	rz / raise           → ステークを引き上げる
//	resp / respond <h|f> → レイズに応答 (h=hold, f=fold)
//	nr / nextround       → 次のディールへ
//	sd / setdifficulty <0-2> → CPU難易度設定
//	st / settarget <n>   → ターゲットスコア設定 (デフォルト15)
//	h / hint             → ヒント表示
//	log / l              → 棋譜表示
func (c *WattenCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.wi.GetConfig()
			return c.wi.ResetWithConfig(cfg)
		},
		[]string{
			"d", "declare",
			"p", "play",
			"rz", "raise",
			"resp", "respond",
			"nr", "nextround",
			"sd", "setdifficulty", "st", "settarget",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "d", "declare":
				return c.handleDeclare(args)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.wi.Play)
			case "rz", "raise":
				return c.wi.Raise(), true
			case "resp", "respond":
				return c.handleRespond(args)
			case "nr", "nextround":
				return c.wi.NextRound(), true
			case "sd", "setdifficulty":
				return cuiutil.WithParsedInt(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2, func(v int) string {
					cfg := c.wi.GetConfig()
					cfg.CpuDifficulty = domain.WattenCpuDifficulty(v)
					return c.wi.ResetWithConfig(cfg)
				})
			case "st", "settarget":
				return cuiutil.WithParsedInt(args, "Target score is required.", "Invalid target score: %s. Please enter 1 or more.", 1, math.MaxInt, func(v int) string {
					cfg := c.wi.GetConfig()
					cfg.TargetScore = v
					return c.wi.ResetWithConfig(cfg)
				})
			default:
				return handleCuiHintAndLog(cmd, c.wi.Hint, c.wi.ActionLog)
			}
		},
	)
}

// handleDeclare は `d <rank> <suit>` を解析して宣言する。
func (c *WattenCuiController) handleDeclare(args []string) (string, bool) {
	rank, errMsg, ok := cuiutil.ParseIntArg(args, "Rank and suit are required (e.g. 'd 10 3').", "Invalid rank: %s.", 1, 13)
	if !ok {
		return errMsg, true
	}
	suit, errMsg, ok := cuiutil.ParseIntArg(args[1:], "Suit is required (1-4).", "Invalid suit: %s.", 1, 4)
	if !ok {
		return errMsg, true
	}
	return c.wi.Declare(rank, suit), true
}

// handleRespond は `resp <h|f>` を解析して応答する。
func (c *WattenCuiController) handleRespond(args []string) (string, bool) {
	if len(args) < 1 {
		return "Response is required (h=hold, f=fold).", true
	}
	switch args[0] {
	case "h", "hold":
		return c.wi.Respond(true), true
	case "f", "fold":
		return c.wi.Respond(false), true
	default:
		return "Invalid response: " + args[0] + " (use h=hold, f=fold).", true
	}
}
