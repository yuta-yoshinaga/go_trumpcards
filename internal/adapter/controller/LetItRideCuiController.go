//go:build !js || !wasm || casino

package controller

import (
	"math"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LetItRideCuiController レット・イット・ライドCUIコントローラークラス
type LetItRideCuiController struct {
	ci usecase.LetItRideInteractorIF
	// pullPending は "p" が一度打たれ、確認待ちであることを表す。
	//
	// **Pull は取り消せない。**Web は専用ダイアログでリスクの前後を見せてから
	// 実行するのに、CUI は "p" 一発で確定していた (#4699)。ここで一段挟む。
	pullPending bool
}

// NewLetItRideCuiController コンストラクタ
func NewLetItRideCuiController(ci usecase.LetItRideInteractorIF) *LetItRideCuiController {
	return &LetItRideCuiController{
		ci: ci,
	}
}

// Exec ゲーム実行
// コマンド例: "r", "b 100", "p", "l", "log", "q"
func (lc *LetItRideCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			// **r / reset は gameHandler を通らない** (execCuiCommand が先に拾う)。
			// ここで消さないと、リセット直後の "y" が配り直した卓に Pull を
			// 走らせる (#6076)。
			lc.pullPending = false
			return lc.ci.Reset()
		},
		[]string{"b", "bet", "p", "pull", "y", "yes", "l", "letitride", "log"},
		func(cmd string, args []string) (string, bool) {
			// **確認待ちは "y" 以外のどのコマンドでも取り消す。**残したままだと、
			// あとで打った "y" が意図しない Pull を確定させてしまう。
			pending := lc.pullPending
			if cmd != "y" && cmd != "yes" {
				lc.pullPending = false
			}

			switch cmd {
			case "b", "bet":
				amount, errMsg, ok := cuiutil.ParseIntArgKeys(args, "betAmountRequired", "invalidBetAmount", 1, math.MaxInt)
				if !ok {
					return errMsg, true
				}
				return lc.ci.Bet(amount), true
			case "p", "pull":
				lc.pullPending = true
				return lc.ci.PullConfirm(), true
			case "y", "yes":
				if !pending {
					// **日本語モードでも英語で返っていた。**
					return invalidArg("letitride.nothingToConfirm"), true
				}
				lc.pullPending = false
				return lc.ci.Pull(), true
			case "l", "letitride":
				return lc.ci.LetItRide(), true
			default:
				return handleCuiLog(cmd, lc.ci.ActionLog)
			}
		},
	)
}
