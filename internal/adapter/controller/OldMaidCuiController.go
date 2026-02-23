package controller

import (
	"strconv"
	"strings"

	"github.com/yuta-yoshinaga/go_trumpcards/internal/usecase"
)

// OldMaidCuiController ババ抜きCUIコントローラークラス
type OldMaidCuiController struct {
	omi usecase.OldMaidInteractorIF
}

// NewOldMaidCuiController コンストラクタ
func NewOldMaidCuiController(omi usecase.OldMaidInteractorIF) *OldMaidCuiController {
	return &OldMaidCuiController{omi: omi}
}

// Exec コマンド実行
// draw コマンドは "d N" または "draw N" の形式でカードインデックスを指定可能。
// 例: "d 2" → インデックス2のカードを引く / "d" → ランダムに引く
func (c *OldMaidCuiController) Exec(command string) string {
	fields := strings.Fields(command)
	if len(fields) == 0 {
		return "コマンドが不明です: " + command
	}
	switch fields[0] {
	case "q", "quit":
		return "bye."
	case "r", "reset":
		return c.omi.Reset()
	case "d", "draw":
		cardIdx := -1
		if len(fields) >= 2 {
			if idx, err := strconv.Atoi(fields[1]); err == nil {
				cardIdx = idx
			}
		}
		return c.omi.Draw(cardIdx)
	default:
		return "コマンドが不明です: " + command
	}
}
