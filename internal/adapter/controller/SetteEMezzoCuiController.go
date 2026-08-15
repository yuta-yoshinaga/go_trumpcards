//go:build !js || !wasm || extra2

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SetteEMezzoCuiController セッテ・エ・メッツォ CUI コントローラークラス
type SetteEMezzoCuiController struct {
	si usecase.SetteEMezzoInteractorIF
}

// NewSetteEMezzoCuiController コンストラクタ
func NewSetteEMezzoCuiController(si usecase.SetteEMezzoInteractorIF) *SetteEMezzoCuiController {
	return &SetteEMezzoCuiController{si: si}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "h", "s", "matta 3", "bh", "bs", "log", "q"
func (sc *SetteEMezzoCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return sc.si.Reset() },
		[]string{"b", "bet", "deal", "h", "hit", "s", "stand", "matta", "bh", "bs", "log"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args,
					"betAmountRequired", "invalidBetAmount",
					domain.SetteEMezzoMinBet, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return sc.si.Bet(amount), true
			case "deal":
				return sc.si.Deal(), true
			case "h", "hit":
				return sc.si.Hit(), true
			case "s", "stand":
				return sc.si.Stand(), true
			case "matta":
				return sc.handleMatta(args), true
			case "bh":
				return sc.si.BankerHit(), true
			case "bs":
				return sc.si.BankerStand(), true
			default:
				if cmd == "log" || cmd == "l" {
					return sc.si.ActionLog(), true
				}
				return "", false
			}
		},
	)
}

// handleMatta マッタの値を受け取る。
//
// 入力は**点数**（0.5 または 1〜7）で受け、内部の半点単位へ直す。プレイヤーに
// 「15 と入力してください」と言わせないため、変換はここで吸収する。
func (sc *SetteEMezzoCuiController) handleMatta(args []string) string {
	if len(args) == 0 {
		return "Matta value is required (0.5 or 1-7)."
	}
	if args[0] == "0.5" {
		return sc.si.Matta(1)
	}
	points, errMsg, ok := cuiutil.ParseIntArg(args,
		"Matta value is required (0.5 or 1-7).",
		"Invalid matta value. Enter 0.5 or a whole number from 1 to 7.", 1, 7)
	if !ok {
		return errMsg
	}
	return sc.si.Matta(points * 2)
}
