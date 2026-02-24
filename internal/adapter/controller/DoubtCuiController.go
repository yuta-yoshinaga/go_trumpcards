package controller

import (
	"strconv"
	"strings"

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
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "コマンドが不明です: " + command
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return c.di.Reset()
	case "p", "play":
		claimedValue := 0
		if len(fields) > 1 {
			if parsed, err := strconv.Atoi(fields[1]); err == nil {
				claimedValue = parsed
			}
		}
		cardIndices := []int{}
		if len(fields) > 2 {
			for _, f := range fields[2:] {
				if parsed, err := strconv.Atoi(f); err == nil {
					cardIndices = append(cardIndices, parsed)
				}
			}
		}
		return c.di.Play(cardIndices, claimedValue)
	case "d", "doubt":
		doubterIndices := []int{}
		for _, f := range fields[1:] {
			if parsed, err := strconv.Atoi(f); err == nil {
				doubterIndices = append(doubterIndices, parsed)
			}
		}
		return c.di.ResolveDoubt(doubterIndices)
	case "s", "skip":
		return c.di.SkipDoubt()
	default:
		return "コマンドが不明です: " + command
	}
}
