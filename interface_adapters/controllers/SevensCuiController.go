package controllers

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/usecases"
)

// SevensCuiController 7並べCUIコントローラークラス
type SevensCuiController struct {
	sgi usecases.SevensInteractorIF
}

// NewSevensCuiController コンストラクタ
func NewSevensCuiController(sgi usecases.SevensInteractorIF) *SevensCuiController {
	return &SevensCuiController{sgi: sgi}
}

// Exec コマンド実行
// play コマンドは "p [インデックス]" の形式でカードインデックスを指定。
// インデックスなし ("p") の場合はパス扱い (idx = -1)。
// 例: "p"  → パス / "p 2" → 2番のカードを出す
func (c *SevensCuiController) Exec(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "コマンドが不明です: " + command
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return c.sgi.Reset()
	case "p", "play":
		idx := -1 // デフォルトはパス
		if len(fields) > 1 {
			if parsed, err := strconv.Atoi(fields[1]); err == nil {
				idx = parsed
			}
		}
		return c.sgi.Play(idx)
	default:
		return "コマンドが不明です: " + command
	}
}
