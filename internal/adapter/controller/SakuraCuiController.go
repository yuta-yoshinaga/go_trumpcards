//go:build !js || !wasm || extra3

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SakuraCuiController はさくら (肥後花) の CUI コントローラークラス。
type SakuraCuiController struct {
	si usecase.SakuraInteractorIF
}

// NewSakuraCuiController コンストラクタ。
func NewSakuraCuiController(si usecase.SakuraInteractorIF) *SakuraCuiController {
	return &SakuraCuiController{si: si}
}

// Exec コマンド実行。
//
//	q / quit             → ゲーム終了 ("bye.")
//	r / reset            → ゲームリセット (設定保持)
//	p / play <h> [f]     → 手札 h を出す。f は 2 枚一致時に取る場札インデックス
//	nr / nextround / n   → 次のラウンドへ
//	ss / setseats <2-4>  → 席数設定
//	sr / setrounds <1-12>→ ラウンド数設定
//	h / hint             → ヒント表示
//	log / l              → 棋譜表示
func (c *SakuraCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			return c.si.ResetWithConfig(c.si.GetConfig())
		},
		[]string{
			"p", "play", "n", "next", "nr", "nextround",
			"ss", "setseats", "sr", "setrounds", "h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.handlePlay(args), true
			case "n", "next", "nr", "nextround":
				return c.si.NextRound(), true
			case "ss", "setseats":
				return cuiutil.WithParsedInt(args,
					"Number of seats is required (2-4).",
					"Invalid number of seats: %s. Please enter 2-4.",
					domain.SakuraMinSeats, domain.SakuraMaxSeats, func(v int) string {
						cfg := c.si.GetConfig()
						cfg.Seats = v
						return c.si.ResetWithConfig(cfg)
					})
			case "sr", "setrounds":
				return cuiutil.WithParsedInt(args,
					"Number of rounds is required (1-12).",
					"Invalid number of rounds: %s. Please enter 1-12.",
					domain.SakuraMinRounds, domain.SakuraMaxRounds, func(v int) string {
						cfg := c.si.GetConfig()
						cfg.Rounds = v
						return c.si.ResetWithConfig(cfg)
					})
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// handlePlay は "p <handIdx> [fieldIdx]" を解析して Play を呼ぶ。
func (c *SakuraCuiController) handlePlay(args []string) string {
	if len(args) == 0 {
		return invalidArg("cardIndexRequiredField")
	}
	handIdx, err := strconv.Atoi(args[0])
	if err != nil {
		return invalidArg("invalidCardIndex", "val", args[0])
	}
	fieldIdx := -1
	if len(args) >= 2 {
		if v, err := strconv.Atoi(args[1]); err == nil {
			fieldIdx = v
		}
	}
	return c.si.Play(handIdx, fieldIdx)
}
