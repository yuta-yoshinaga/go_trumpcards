//go:build !js || !wasm || extra2

package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// LaughAndLieDownCuiController ラフ・アンド・ライダウンCUIコントローラークラス
type LaughAndLieDownCuiController struct {
	li usecase.LaughAndLieDownInteractorIF
}

// NewLaughAndLieDownCuiController コンストラクタ
func NewLaughAndLieDownCuiController(li usecase.LaughAndLieDownInteractorIF) *LaughAndLieDownCuiController {
	return &LaughAndLieDownCuiController{li: li}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit          → ゲーム終了 ("bye.")
//	r / reset         → ゲームリセット (設定保持)
//	p / play <i>      → 手札の i 番目で場札 1 枚を取る
//	p / play <i> 3    → 同じ札で場札 3 枚を取る (場に 3 枚以上あるときのみ)
//	h / hint          → ヒント表示
//	log / l           → 棋譜表示
func (c *LaughAndLieDownCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string {
			cfg := c.li.GetConfig()
			return c.li.ResetWithConfig(cfg)
		},
		[]string{"p", "play", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				return c.play(args)
			default:
				return handleCuiHintAndLog(cmd, c.li.Hint, c.li.ActionLog)
			}
		},
	)
}

// play は `p <i> [take]` を解釈する。
//
// 取得枚数は省略できる。1 枚取りが普通で、3 枚取りは場に 3 枚以上あるときだけの
// 選択肢なので、毎回打たせるのは無駄が多い。
func (c *LaughAndLieDownCuiController) play(args []string) (string, bool) {
	if len(args) == 0 {
		return invalidArg("cardIndexRequired"), true
	}
	handIdx, err := strconv.Atoi(args[0])
	if err != nil || handIdx < 0 {
		return invalidArg("invalidCardIndex", "val", args[0]), true
	}
	take := 1
	if len(args) > 1 {
		take, err = strconv.Atoi(args[1])
		if err != nil {
			return "Invalid take count: " + args[1] + ".", true
		}
	}
	return c.li.Play(handIdx, take), true
}
