package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SpeedCuiController スピードCUIコントローラークラス
type SpeedCuiController struct {
	si usecase.SpeedInteractorIF
}

// NewSpeedCuiController コンストラクタ
func NewSpeedCuiController(si usecase.SpeedInteractorIF) *SpeedCuiController {
	return &SpeedCuiController{si: si}
}

// Exec コマンド実行
//
//	q / quit         → ゲーム終了
//	r / reset        → ゲームリセット
//	p <card> <pile>  → カードを出す (card=手札idx, pile=台札idx)
//	play (同上)
//	f / flip         → 膠着時にカードをめくる
//	sd / setdifficulty <0-2> → CPU難易度設定
//	h / hint         → ヒント表示
//	log / l          → 棋譜表示
func (c *SpeedCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.si.GetConfig()
			return c.si.ResetWithConfig(cfg)
		},
		[]string{
			"p", "play", "f", "flip", "sd", "setdifficulty",
			"h", "hint", "log", "l",
		},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				cardIdx, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 0, 0, "invalidCardIndex")
				if !ok {
					return errMsg, true
				}
				pileIdx, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 1, 0, "invalidPileIndex")
				if !ok {
					return errMsg, true
				}
				return c.si.Play(cardIdx, pileIdx), true
			case "f", "flip":
				return c.si.Flip(), true
			case "sd", "setdifficulty":
				val, errMsg, ok := cuiutil.ParseOptionalIntKeys(args, 0, int(domain.SpeedCpuDifficultyNormal), "invalidCpuDifficulty")
				if !ok {
					return errMsg, true
				}
				cfg := c.si.GetConfig()
				cfg.CpuDifficulty = domain.SpeedCpuDifficulty(val)
				return c.si.ResetWithConfig(cfg), true
			case "h", "hint":
				return c.si.Hint(), true
			case "log", "l":
				return c.si.ActionLog(), true
			default:
				return "", false
			}
		},
	)
}
