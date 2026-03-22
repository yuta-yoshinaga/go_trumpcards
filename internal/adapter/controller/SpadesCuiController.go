package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpadesCuiController スペードCUIコントローラークラス
type SpadesCuiController struct {
	si usecase.SpadesInteractorIF
}

// NewSpadesCuiController コンストラクタ
func NewSpadesCuiController(si usecase.SpadesInteractorIF) *SpadesCuiController {
	return &SpadesCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット (設定保持)
//	b / bid <n>      → ビッドを宣言
//	p / play <i>     → カードをプレイ
//	n / next         → 次のトリックへ
//	nr / nextround   → 次のラウンドへ (スコアリング)
//	sd / setdifficulty <0-2> → CPU難易度設定
//	sl / setlimit <n> → ポイント上限設定
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *SpadesCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"b", "bid", "p", "play", "n", "next", "nr", "nextround",
			"sd", "setdifficulty", "sl", "setlimit", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bid":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Bid value is required (0-13).", "Invalid bid value: %s.", 0, 13)
				if !ok {
					return errMsg, true
				}
				return c.si.Bid(v), true
			case "p", "play":
				idx, errMsg, ok := cuiutil.ParseIntArg(args, "Card index is required.", "Invalid card index: %s.", cuiutil.NoMin, cuiutil.NoMax)
				if !ok {
					return errMsg, true
				}
				return c.si.Play(idx), true
			case "n", "next":
				return c.si.NextTrick(), true
			case "nr", "nextround":
				return c.si.NextRound(), true
			case "sd", "setdifficulty":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "CPU difficulty is required (0=Easy, 1=Normal, 2=Hard).", "Invalid CPU difficulty: %s. Please enter 0-2.", 0, 2)
				if !ok {
					return errMsg, true
				}
				cfg := c.si.GetConfig()
				cfg.CpuDifficulty = domain.SpadesCpuDifficulty(v)
				return c.si.ResetWithConfig(cfg), true
			case "sl", "setlimit":
				v, errMsg, ok := cuiutil.ParseIntArg(args, "Point limit is required.", "Invalid point limit: %s. Please enter 1 or more.", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				cfg := c.si.GetConfig()
				cfg.PointLimit = v
				return c.si.ResetWithConfig(cfg), true
			case "h", "hint":
				return c.si.Hint(), true
			case "log", "l":
				return c.si.ActionLog(), true
			}
			return "", false
		},
	)
}
