package controller

import (
	"strconv"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// DoubtCuiController ダウトCUIコントローラークラス
type DoubtCuiController struct {
	di usecase.DoubtInteractorIF
}

// NewDoubtCuiController コンストラクタ
func NewDoubtCuiController(di usecase.DoubtInteractorIF) *DoubtCuiController {
	return &DoubtCuiController{di: di}
}

// Exec コマンド実行
// コマンド一覧:
//
//	q / quit         → ゲーム終了 ("bye.")
//	r / reset        → ゲームリセット
//	p <値> <idx...>  → カードを出す (値=宣言する値, idx=カードインデックス)
//	play <値> <idx...> (同上)
//	d <idx...>       → ダウト (idx=ダウトするプレイヤーインデックス)
//	doubt <idx...>   (同上)
//	s / skip         → ダウトをスキップ
func (c *DoubtCuiController) Exec(command string) string {
	return execCuiCommand(
		command,
		func(_ []string) string { return c.di.Reset() },
		func(command string) string { return "コマンドが不明です: " + command },
		func(cmd string, args []string) (string, bool) {
			switch cmd {
			case "p", "play":
				claimedValue := 0
				if len(args) > 0 {
					if parsed, err := strconv.Atoi(args[0]); err == nil {
						claimedValue = parsed
					}
				}
				cardIndices := []int{}
				if len(args) > 1 {
					for _, f := range args[1:] {
						if parsed, err := strconv.Atoi(f); err == nil {
							cardIndices = append(cardIndices, parsed)
						}
					}
				}
				return c.di.Play(cardIndices, claimedValue), true
			case "d", "doubt":
				doubterIndices := []int{}
				for _, f := range args {
					if parsed, err := strconv.Atoi(f); err == nil {
						doubterIndices = append(doubterIndices, parsed)
					}
				}
				return c.di.ResolveDoubt(doubterIndices), true
			case "s", "skip":
				return c.di.SkipDoubt(), true
			}
			return "", false
		},
	)
}
