//go:build !js || !wasm || extra

package controller

import (
	"github.com/yuta-yoshinaga/go_trumpcards/internal/adapter/controller/cuiutil"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/domain"
	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// SergeantMajorCuiController サージェントメジャーCUIコントローラークラス
type SergeantMajorCuiController struct {
	si usecase.SergeantMajorInteractorIF
}

// NewSergeantMajorCuiController コンストラクタ
func NewSergeantMajorCuiController(si usecase.SergeantMajorInteractorIF) *SergeantMajorCuiController {
	return &SergeantMajorCuiController{si: si}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit                  → ゲーム終了 ("bye.")
//	r / reset                 → ゲームリセット (設定保持)
//	t / trump <s>             → 切り札を宣言する (1:♠ 2:♣ 3:♥ 4:♦、親のみ)
//	d / discard <i> <i> <i> <i> → キティのぶん4枚を捨てる (親のみ)
//	p / play <i>              → 手札の i 番目を出す
//	n / next                  → 次のラウンドへ
//	g / giveup                → 投了
//	h / hint                  → ヒント表示
//	log / l                   → 棋譜表示
//
// **ノルマを宣言するコマンドは無い。** 8・5・3 は席順で決まります。
func (c *SergeantMajorCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.si.ResetWithConfig(c.si.GetConfig()) },
		[]string{"t", "trump", "d", "discard", "p", "play", "n", "next", "g", "giveup", "h", "hint", "log", "l"},
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "t", "trump":
				return cuiutil.WithParsedIntKeys(args, "suitRequired", "invalidSuit",
					domain.CardDesignSpade, domain.CardDesignMax, c.si.DeclareTrump)
			case "d", "discard":
				return c.discard(args)
			case "p", "play":
				return cuiutil.WithParsedIntKeys(args, "cardIndexRequired", "invalidCardIndex", cuiutil.NoMin, cuiutil.NoMax, c.si.Play)
			case "n", "next":
				return c.si.NextRound(), true
			case "g", "giveup":
				return c.si.GiveUp(), true
			default:
				return handleCuiHintAndLog(cmd, c.si.Hint, c.si.ActionLog)
			}
		},
	)
}

// discard は捨て札のインデックスを 4 つ取る。
//
// **既定値で埋めない。** 埋めると捨てていない札が捨てられる。
func (c *SergeantMajorCuiController) discard(args []string) (string, bool) {
	if len(args) < domain.SergeantMajorKittySize {
		return "Four card indices are required.", true
	}
	indices := make([]int, 0, domain.SergeantMajorKittySize)
	for i := range domain.SergeantMajorKittySize {
		v, errMsg, ok := cuiutil.ParseIntArgKeys(args[i:], "cardIndexRequired", "invalidCardIndex",
			cuiutil.NoMin, cuiutil.NoMax)
		if !ok {
			return errMsg, true
		}
		indices = append(indices, v)
	}
	return c.si.Discard(indices), true
}
